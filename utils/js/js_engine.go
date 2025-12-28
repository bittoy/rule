/*
 * Copyright 2023 The RuleGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package js provides JavaScript execution capabilities for the RuleGo rule engine.
//
// This package implements a JavaScript engine using the goja library, allowing
// for the execution of JavaScript code within the rule engine. It includes
// functionality for creating and managing JavaScript virtual machines,
// compiling and caching JavaScript programs, and executing JavaScript code
// with access to global variables and user-defined functions.
//
// Key components:
// - GojaJsEngine: The main struct representing the JavaScript engine.
// - NewGojaJsEngine: Function to create a new instance of the JavaScript engine.
// - PreCompileJs: Method to precompile user-defined JavaScript functions.
//
// The package supports features such as:
// - Pooling of JavaScript VMs for efficient reuse
// - Precompilation of JavaScript code for improved performance
// - Integration with the RuleGo configuration system
// - Access to global variables and functions within JavaScript code
//
// This package is crucial for components that require JavaScript execution,
// such as the JsTransformNode and JsFilterNode in the action package.
package js

import (
	"context"
	"errors"

	"github.com/bittoy/rule/types"

	"github.com/dop251/goja"
)

const (
	GlobalKey = "global"
	MetaData  = "metadata"
)

// GojaJsEngine goja js engine
type GojaJsEngine struct {
	config            types.Config
	vm                *goja.Runtime
	jsUdfProgramCache map[string]*goja.Program
}

// NewGojaJsEngine Create a new instance of the JavaScript engine
func NewGojaJsEngine(config types.Config, jsScript string, fromVars map[string]any) (*GojaJsEngine, error) {
	vm := goja.New()
	_, err := vm.RunString(jsScript)
	if err != nil {
		return nil, err
	}

	// Set global properties directly
	// if len(config.Properties.Values()) != 0 {
	// 	if err := vm.Set(GlobalKey, config.Properties.Values()); err != nil {
	// 		config.Logger.Printf("set global properties error: %s", err.Error())
	// 	}
	// }
	// if len(fromVars) != 0 {
	// 	if err := vm.Set(MetaData, fromVars); err != nil {
	// 		config.Logger.Printf("set fromVars %v error: %s", fromVars, err.Error())
	// 	}
	// }

	return &GojaJsEngine{
		config: config,
		vm:     vm,
	}, nil
}

// Execute Execute JavaScript script
func (g *GojaJsEngine) Execute(ctx context.Context, rCtx types.RuleContext, funcName string, argumentList ...any) (out interface{}, err error) {
	// Optimized parameter conversion - pre-allocate slice
	var params []goja.Value
	if len(argumentList) > 0 {
		params = make([]goja.Value, len(argumentList))
		for i, v := range argumentList {
			params[i] = g.vm.ToValue(v)
		}
	}

	f, ok := goja.AssertFunction(g.vm.Get(funcName))
	if !ok {
		return nil, errors.New(funcName + " is not a function")
	}

	// Execute function
	res, err := f(goja.Undefined(), params...)
	if err != nil {
		return nil, err
	}
	return res.Export(), nil
}

func (g *GojaJsEngine) Stop() {
}
