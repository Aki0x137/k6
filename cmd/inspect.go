/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2016 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"go.k6.io/k6/js"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/lib/metrics"
	"go.k6.io/k6/lib/types"
)

func getInspectCmd(globalState *globalState) *cobra.Command {
	var addExecReqs bool

	// inspectCmd represents the inspect command
	inspectCmd := &cobra.Command{
		Use:   "inspect [file]",
		Short: "Inspect a script or archive",
		Long:  `Inspect a script or archive.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, filesystems, err := readSource(globalState, args[0])
			if err != nil {
				return err
			}

			runtimeOptions, err := getRuntimeOptions(cmd.Flags(), globalState.envVars)
			if err != nil {
				return err
			}
			registry := metrics.NewRegistry()

			var b *js.Bundle
			typ := globalState.flags.runType
			if typ == "" {
				typ = detectType(src.Data)
			}
			switch typ {
			// this is an exhaustive list
			case typeArchive:
				var arc *lib.Archive
				arc, err = lib.ReadArchive(bytes.NewBuffer(src.Data))
				if err != nil {
					return err
				}
				b, err = js.NewBundleFromArchive(globalState.logger, arc, runtimeOptions, registry)

			case typeJS:
				b, err = js.NewBundle(globalState.logger, src, filesystems, runtimeOptions, registry)
			}
			if err != nil {
				return err
			}

			// ATM, output can take 2 forms: standard (equal to lib.Options struct) and extended, with additional fields.
			inspectOutput := interface{}(b.Options)

			if addExecReqs {
				inspectOutput, err = addExecRequirements(globalState, b)
				if err != nil {
					return err
				}
			}

			data, err := json.MarshalIndent(inspectOutput, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data)) //nolint:forbidigo // yes we want to just print it

			return nil
		},
	}

	inspectCmd.Flags().SortFlags = false
	inspectCmd.Flags().AddFlagSet(runtimeOptionFlagSet(false))
	inspectCmd.Flags().StringVarP(&globalState.flags.runType, "type", "t",
		globalState.flags.runType, "override file `type`, \"js\" or \"archive\"")
	inspectCmd.Flags().BoolVar(&addExecReqs,
		"execution-requirements",
		false,
		"include calculations of execution requirements for the test")

	return inspectCmd
}

func addExecRequirements(gs *globalState, b *js.Bundle) (interface{}, error) {
	conf, err := getConsolidatedConfig(gs, Config{}, b.Options)
	if err != nil {
		return nil, err
	}

	conf, err = deriveAndValidateConfig(conf, b.IsExecutable, gs.logger)
	if err != nil {
		return nil, err
	}

	et, err := lib.NewExecutionTuple(conf.ExecutionSegment, conf.ExecutionSegmentSequence)
	if err != nil {
		return nil, err
	}

	executionPlan := conf.Scenarios.GetFullExecutionRequirements(et)
	duration, _ := lib.GetEndOffset(executionPlan)

	return struct {
		lib.Options
		TotalDuration types.NullDuration `json:"totalDuration"`
		MaxVUs        uint64             `json:"maxVUs"`
	}{
		conf.Options,
		types.NewNullDuration(duration, true),
		lib.GetMaxPossibleVUs(executionPlan),
	}, nil
}
