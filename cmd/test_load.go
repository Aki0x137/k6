package cmd

import (
	"archive/tar"
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.k6.io/k6/cmd/state"
	"go.k6.io/k6/errext"
	"go.k6.io/k6/errext/exitcodes"
	"go.k6.io/k6/event"
	"go.k6.io/k6/js"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/lib/fsext"
	"go.k6.io/k6/loader"
	"go.k6.io/k6/metrics"
	"gopkg.in/guregu/null.v3"
)

const (
	testTypeJS      = "js"
	testTypeArchive = "archive"
)

// loadedTest contains all of data, details and dependencies of a loaded
// k6 test, but without any config consolidation.
type loadedTest struct {
	sourceRootPath     string // contains the raw string the user supplied
	pwd                string
	source             *loader.SourceData
	fs                 fsext.Fs
	fileSystems        map[string]fsext.Fs
	preInitState       *lib.TestPreInitState
	initRunner         lib.Runner // TODO: rename to something more appropriate
	keyLogger          io.Closer
	consolidatedConfig Config
	derivedConfig      Config
}

func loadTest(
	gs *state.GlobalState, cmd *cobra.Command, args []string,
	cliConfigGetter func(flags *pflag.FlagSet) (Config, error),
) (*loadedTest, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("k6 needs at least one argument to load the test")
	}

	sourceRootPath := args[0]
	gs.Logger.Debugf("Resolving and reading test '%s'...", sourceRootPath)
	src, fileSystems, pwd, err := readSource(gs, sourceRootPath)
	if err != nil {
		return nil, err
	}
	resolvedPath := src.URL.String()
	gs.Logger.Debugf(
		"'%s' resolved to '%s' and successfully loaded %d bytes!",
		sourceRootPath, resolvedPath, len(src.Data),
	)

	gs.Logger.Debugf("Gathering k6 runtime options...")
	runtimeOptions, err := getRuntimeOptions(cmd.Flags(), gs.Env)
	if err != nil {
		return nil, err
	}

	var cliConfig Config
	if cliConfigGetter != nil {
		gs.Logger.Debug("Parsing CLI flags...")
		var err error
		cliConfig, err = cliConfigGetter(cmd.Flags())
		if err != nil {
			return nil, err
		}
	}

	keyLogger, err := openSSLKeyLogFile(gs, pwd, runtimeOptions.KeyWriter)
	if err != nil {
		return nil, err
	}

	registry := metrics.NewRegistry()
	state := &lib.TestPreInitState{
		Logger:         gs.Logger,
		RuntimeOptions: runtimeOptions,
		Registry:       registry,
		BuiltinMetrics: metrics.RegisterBuiltinMetrics(registry),
		Events:         event.NewEventSystem(),
		LookupEnv: func(key string) (string, bool) {
			val, ok := gs.Env[key]
			return val, ok
		},
	}
	if keyLogger != nil {
		state.KeyLogger = &syncWriter{w: keyLogger}
	}

	gs.Logger.Debugf("Initializing k6 runner for '%s' (%s)...", sourceRootPath, resolvedPath)
	initRunner, err := initializeFirstRunner(gs, state, src, pwd, fileSystems)
	if err != nil {
		return nil, fmt.Errorf("could not initialize '%s': %w", sourceRootPath, err)
	}
	gs.Logger.Debug("Runner successfully initialized!")

	gs.Logger.Debug("Consolidating config layers...")
	consolidatedConfig, err := getConsolidatedConfig(gs, cliConfig, initRunner.GetOptions())
	if err != nil {
		return nil, err
	}

	// return consolidateDeriveAndValidateConfig(gs, cmd, cliConfigGetter)

	gs.Logger.Debug("Parsing thresholds and validating config...")
	// Parse the thresholds, only if the --no-threshold flag is not set.
	// If parsing the threshold expressions failed, consider it as an
	// invalid configuration error.
	if !runtimeOptions.NoThresholds.Bool {
		for metricName, thresholdsDefinition := range consolidatedConfig.Options.Thresholds {
			err = thresholdsDefinition.Parse()
			if err != nil {
				return nil, errext.WithExitCodeIfNone(err, exitcodes.InvalidConfig)
			}

			err = thresholdsDefinition.Validate(metricName, registry)
			if err != nil {
				return nil, errext.WithExitCodeIfNone(err, exitcodes.InvalidConfig)
			}
		}
	}

	derivedConfig, err := deriveAndValidateConfig(consolidatedConfig, initRunner.IsExecutable, gs.Logger)
	if err != nil {
		return nil, err
	}

	test := &loadedTest{
		pwd:                pwd,
		sourceRootPath:     sourceRootPath,
		source:             src,
		fs:                 gs.FS,
		fileSystems:        fileSystems,
		preInitState:       state,
		initRunner:         initRunner,
		consolidatedConfig: consolidatedConfig,
		derivedConfig:      derivedConfig,
	}

	return test, nil
}

func initializeFirstRunner(gs *state.GlobalState, preInitState *lib.TestPreInitState, source *loader.SourceData, pwd string, fileSystems map[string]fsext.Fs) (runner lib.Runner, err error) {
	testPath := source.URL.String()
	logger := gs.Logger.WithField("test_path", testPath)

	testType := preInitState.RuntimeOptions.TestType.String
	if testType == "" {
		logger.Debug("Detecting test type for...")
		testType = detectTestType(source.Data)
	}

	switch testType {
	case testTypeJS:
		logger.Debug("Trying to load as a JS test...")
		runner, err = js.New(preInitState, source, fileSystems)
		// TODO: should we use common.UnwrapGojaInterruptedError() here?
		if err != nil {
			return nil, fmt.Errorf("could not load JS test '%s': %w", testPath, err)
		}
	case testTypeArchive:
		logger.Debug("Trying to load test as an archive bundle...")

		var arc *lib.Archive
		arc, err := lib.ReadArchive(bytes.NewReader(source.Data))
		if err != nil {
			return nil, fmt.Errorf("could not load test archive bundle '%s': %w", testPath, err)
		}
		logger.Debugf("Loaded test as an archive bundle with type '%s'!", arc.Type)

		switch arc.Type {
		case testTypeJS:
			logger.Debug("Evaluating JS from archive bundle...")
			runner, err = js.NewFromArchive(preInitState, arc)
			if err != nil {
				return nil, fmt.Errorf("could not load JS from test archive bundle '%s': %w", testPath, err)
			}
		default:
			return nil, fmt.Errorf("archive '%s' has an unsupported test type '%s'", testPath, arc.Type)
		}
	default:
		return nil, fmt.Errorf("unknown or unspecified test type '%s' for '%s'", testType, testPath)
	}

	return runner, nil
}

// readSource is a small wrapper around loader.ReadSource returning
// result of the load and filesystems map
func readSource(gs *state.GlobalState, filename string) (*loader.SourceData, map[string]fsext.Fs, string, error) {
	pwd, err := gs.Getwd()
	if err != nil {
		return nil, nil, "", err
	}

	filesystems := loader.CreateFilesystems(gs.FS)
	src, err := loader.ReadSource(gs.Logger, filename, pwd, filesystems, gs.Stdin)
	return src, filesystems, pwd, err
}

func detectTestType(data []byte) string {
	if _, err := tar.NewReader(bytes.NewReader(data)).Next(); err == nil {
		return testTypeArchive
	}
	return testTypeJS
}

func (lt *loadedTest) consolidateDeriveAndValidateConfig(
	gs *state.GlobalState, cmd *cobra.Command,
	cliConfGetter func(flags *pflag.FlagSet) (Config, error), // TODO: obviate
) (*loadedAndConfiguredTest, error) {
	var cliConfig Config
	if cliConfGetter != nil {
		gs.Logger.Debug("Parsing CLI flags...")
		var err error
		cliConfig, err = cliConfGetter(cmd.Flags())
		if err != nil {
			return nil, err
		}
	}

	gs.Logger.Debug("Consolidating config layers...")
	consolidatedConfig, err := getConsolidatedConfig(gs, cliConfig, lt.initRunner.GetOptions())
	if err != nil {
		return nil, err
	}

	gs.Logger.Debug("Parsing thresholds and validating config...")
	// Parse the thresholds, only if the --no-threshold flag is not set.
	// If parsing the threshold expressions failed, consider it as an
	// invalid configuration error.
	if !lt.preInitState.RuntimeOptions.NoThresholds.Bool {
		for metricName, thresholdsDefinition := range consolidatedConfig.Options.Thresholds {
			err = thresholdsDefinition.Parse()
			if err != nil {
				return nil, errext.WithExitCodeIfNone(err, exitcodes.InvalidConfig)
			}

			err = thresholdsDefinition.Validate(metricName, lt.preInitState.Registry)
			if err != nil {
				return nil, errext.WithExitCodeIfNone(err, exitcodes.InvalidConfig)
			}
		}
	}

	derivedConfig, err := deriveAndValidateConfig(consolidatedConfig, lt.initRunner.IsExecutable, gs.Logger)
	if err != nil {
		return nil, err
	}

	return &loadedAndConfiguredTest{
		loadedTest:         lt,
		consolidatedConfig: consolidatedConfig,
		derivedConfig:      derivedConfig,
	}, nil
}

// loadedAndConfiguredTest contains the whole loadedTest, as well as the
// consolidated test config and the full test run state.
// type loadedAndConfiguredTest struct {
// 	*loadedTest
// 	consolidatedConfig Config
// 	derivedConfig      Config
// }

// func loadAndConfigureTest(
// 	gs *state.GlobalState, cmd *cobra.Command, args []string,
// 	cliConfigGetter func(flags *pflag.FlagSet) (Config, error),
// ) (*loadedAndConfiguredTest, error) {
// 	test, err := loadTest(gs, cmd, args, cliConfigGetter)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return test, nil
// }

func openSSLKeyLogFile(gs *state.GlobalState, path string, filename null.String) (keyLogFile io.Writer, err error) {
	if !filename.Valid {
		return nil, nil
	}

	keylogFilename := filename.String
	// if path is absolute - no point doing anything
	if !filepath.IsAbs(keylogFilename) {
		// filepath.Abs could be used but it will get the pwd from `os` package instead of what is configured in GlobalState.
		// This is against our general approach of not using `os` directly and makes testing harder.
		keylogFilename = filepath.Join(path, keylogFilename)
	}
	gs.Logger.Warnf("SSLKEYLOGFILE was specified, logging TLS connection keys to '%s'...",
		filename)
	keyLogFile, err = gs.FS.OpenFile(keylogFilename, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("couldn't get absolute path for keylog file: %w", err)
	}

	return keyLogFile, nil
}

// loadSystemCertPool attempts to load system certificates.
func loadSystemCertPool(logger logrus.FieldLogger) {
	if _, err := x509.SystemCertPool(); err != nil {
		logger.WithError(err).Warning("Unable to load system cert pool")
	}
}

func (lct *loadedAndConfiguredTest) buildTestRunState(
	configToReinject lib.Options,
) (*lib.TestRunState, error) {
	// This might be the full derived or just the consodlidated options
	if err := lct.initRunner.SetOptions(configToReinject); err != nil {
		return nil, err
	}

	// it pre-loads system certificates to avoid doing it on the first TLS request.
	// This is done async to avoid blocking the rest of the loading process as it will not stop if it fails.
	go loadSystemCertPool(lct.preInitState.Logger)

	return &lib.TestRunState{
		TestPreInitState: lct.preInitState,
		Runner:           lct.initRunner,
		Options:          lct.derivedConfig.Options, // we will always run with the derived options
		RunTags:          lct.preInitState.Registry.RootTagSet().WithTagsFromMap(configToReinject.RunTags),
	}, nil
}

type syncWriter struct {
	w io.Writer
	m sync.Mutex
}

func (cw *syncWriter) Write(b []byte) (int, error) {
	cw.m.Lock()
	defer cw.m.Unlock()
	return cw.w.Write(b)
}
