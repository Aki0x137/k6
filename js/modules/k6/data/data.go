/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2020 Load Impact
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

package data

import (
	"errors"
	"strconv"
	"sync"

	"github.com/dop251/goja"
	"go.k6.io/k6/js/common"
	"go.k6.io/k6/js/modules"
)

type (
	// RootModule is the global module instance that will create module
	// instances for each VU.
	RootModule struct {
		shared    sharedArrays
		immutable sharedArrayBuffers
	}

	// Data represents an instance of the data module.
	Data struct {
		vu        modules.VU
		shared    *sharedArrays
		immutable *sharedArrayBuffers
	}

	sharedArrays struct {
		data map[string]sharedArray
		mu   sync.RWMutex
	}

	sharedArrayBuffers struct {
		data map[string]SharedArrayBuffer
		mu   sync.RWMutex
	}
)

var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &Data{}
)

// New returns a pointer to a new RootModule instance.
func New() *RootModule {
	return &RootModule{
		shared: sharedArrays{
			data: make(map[string]sharedArray),
		},
		immutable: sharedArrayBuffers{
			data: make(map[string]SharedArrayBuffer),
		},
	}
}

// NewModuleInstance implements the modules.Module interface to return
// a new instance for each VU.
func (rm *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &Data{
		vu:        vu,
		shared:    &rm.shared,
		immutable: &rm.immutable,
	}
}

// GetOrCreateSharedArrayBuffer fetches a shared array buffer instance with
// the provided name from the module, if it does not exist yet, it is created
// from the provided data.
//
// This method is for instance used by the init context's open method, when the
// `b` and `r` options are supplied, in order to ensure we read a binary file
// from disk only once, but allow repetitive access to its underlying data.
func (rm *RootModule) GetOrCreateSharedArrayBuffer(filename string, data []byte) SharedArrayBuffer {
	return rm.immutable.get(filename, data)
}

// Exports returns the exports of the data module.
func (d *Data) Exports() modules.Exports {
	return modules.Exports{
		Named: map[string]interface{}{
			"SharedArray": d.sharedArray,
		},
	}
}

// sharedArray is a constructor returning a shareable read-only array
// indentified by the name and having their contents be whatever the call returns
func (d *Data) sharedArray(call goja.ConstructorCall) *goja.Object {
	rt := d.vu.Runtime()

	if d.vu.State() != nil {
		common.Throw(rt, errors.New("new SharedArray must be called in the init context"))
	}

	name := call.Argument(0).String()
	if name == "" {
		common.Throw(rt, errors.New("empty name provided to SharedArray's constructor"))
	}

	fn, ok := goja.AssertFunction(call.Argument(1))
	if !ok {
		common.Throw(rt, errors.New("a function is expected as the second argument of SharedArray's constructor"))
	}

	array := d.shared.get(rt, name, fn)
	return array.wrap(rt).ToObject(rt)
}

func (s *sharedArrays) get(rt *goja.Runtime, name string, call goja.Callable) sharedArray {
	s.mu.RLock()
	array, ok := s.data[name]
	s.mu.RUnlock()
	if !ok {
		s.mu.Lock()
		defer s.mu.Unlock()
		array, ok = s.data[name]
		if !ok {
			array = getShareArrayFromCall(rt, call)
			s.data[name] = array
		}
	}

	return array
}

func getShareArrayFromCall(rt *goja.Runtime, call goja.Callable) sharedArray {
	gojaValue, err := call(goja.Undefined())
	if err != nil {
		common.Throw(rt, err)
	}
	obj := gojaValue.ToObject(rt)
	if obj.ClassName() != "Array" {
		common.Throw(rt, errors.New("only arrays can be made into SharedArray")) // TODO better error
	}
	arr := make([]string, obj.Get("length").ToInteger())

	stringify, _ := goja.AssertFunction(rt.GlobalObject().Get("JSON").ToObject(rt).Get("stringify"))
	var val goja.Value
	for i := range arr {
		val, err = stringify(goja.Undefined(), obj.Get(strconv.Itoa(i)))
		if err != nil {
			panic(err)
		}
		arr[i] = val.String()
	}

	return sharedArray{arr: arr}
}

// FIXME: not sure if this is needed. Do we want to expose the sharedArrayBuffer to users?
// func (d *Data) sharedArrayBuffer(constructor goja.ConstructorCall) goja.Value {
// 	// runtime := d.vu.Runtime()

// 	// if d.vu.State() != nil {
// 	// 	common.Throw(runtime, errors.New("new ImmutableArrayBuffer must be called in the init context"))
// 	// }

// 	// filename := constructor.Argument(0).String()
// 	// if filename == "" {
// 	// 	common.Throw(runtime, errors.New("empty filename provided to ImmutableArrayBuffer's constructor"))
// 	// }

// 	// // TODO: somehow acquire data? It looks like having a constructor just doesn't make sense.

// 	// // FIXME: nil is a leftover, some data should actually make its way here? See previous TODO.
// 	// array := d.immutable.get(filename, nil)
// 	// return array.Wrap(runtime).ToObject(runtime)
// 	// return array.Wrap(runtime).ToObject(runtime)
// 	return nil
// }

func (i *sharedArrayBuffers) get(filename string, data []byte) SharedArrayBuffer {
	i.mu.RLock()
	array, exists := i.data[filename]
	i.mu.RUnlock()
	if !exists {
		// If the array was not found, we need to try and
		// create it. Thus, we acquire a read/write lock,
		// which will be released at the end of this if
		// statement's scope.
		i.mu.Lock()
		defer i.mu.Unlock()

		// To ensure atomicity of our operation, and as we have
		// reacquired a lock on the data, we should double check
		// the pre-existence of the array (it might have been created
		// in the meantime).
		array, exists = i.data[filename]
		if !exists {
			array = SharedArrayBuffer{arr: data}
			i.data[filename] = array
		}
	}

	return array
}
