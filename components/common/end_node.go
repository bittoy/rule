/*
 * Copyright 2025 The RuleGo Authors.
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

package common

import (
	"context"
	"errors"
	"reflect"

	"rule/types"
	"rule/utils/maps"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// init registers the EndNode component with the default registry.
func init() {
	Registry.Add(&EndNode{})
}

// ExprAssignNodeConfiguration ExprAssignNode配置结构
type EndNodeConfiguration struct {
	// ExprScript JavaScript脚本，用于确定消息路由路径
	// 函数参数：msg, metadata, msgType, dataType
	// 必须返回字符串数组，表示路由关系类型
	//
	// 内置变量：
	//   - $ctx: 上下文对象，提供缓存操作
	//   - global: 全局配置属性
	//   - vars: 规则链变量
	//   - UDF函数: 用户自定义函数
	//
	// 示例: "return ['route1', 'route2'];"
	Script string
}

// EndNode 结束节点组件，用于触发规则链的结束回调。如果规则链设置了结束节点组件，则会替代默认的分支结束行为，只有运行到结束节点组件时，才会触发结束回调
// EndNode is an end node component that triggers the end callback of the rule chain. If the rule chain has an end node component set, it will replace the default branch ending behavior.
//
// 功能说明：
// Function Description:
// 1. 接收消息并触发DoOnEnd回调 - Receives messages and triggers DoOnEnd callback
// 2. 使用上一个节点传入的关系类型 - Uses the relation type passed from the previous node
// 3. 不会继续传递消息到下一个节点 - Does not continue passing messages to next nodes
//
// 使用场景：
// Use Cases:
// - 规则链的明确结束点 - Explicit end point of rule chains
// - 触发特定的结束处理逻辑 - Trigger specific end processing logic
// - 替代默认的分支结束行为 - Replace default branch ending behavior
type EndNode struct {
	// Config 节点配置
	Config EndNodeConfiguration

	// program 用于高效评估的编译表达式
	// program is the compiled expression for efficient evaluation
	program *vm.Program
}

// Type 返回组件类型
// Type returns the component type identifier.
func (x *EndNode) Type() types.NodeType {
	return types.RuleSubTypeEnd
}

// New creates a new instance.
func (x *EndNode) New() types.Node {
	return &EndNode{Config: EndNodeConfiguration{
		Script: `{}`,
	}}
}

// Init initializes the component.
func (x *EndNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.Config)
	if err != nil {
		return err
	}

	program, err := expr.Compile(x.Config.Script, expr.AllowUndefinedVariables(), expr.AsKind(reflect.Map))
	if err != nil {
		return err
	}
	x.program = program

	return nil
}

// OnMsg processes the incoming message and triggers the end callback.
func (x *EndNode) OnMsg(ctx context.Context, rCtx types.RuleContext, msg types.RuleMsg) error {
	out, err := vm.Run(x.program, msg.GetInput())
	if err != nil {
		return types.NewEngineError(rCtx, msg, err)
	}
	if formatData, ok := out.(map[string]any); ok {
		msg.ClearInnerData()
		msg.SetChainOutput(formatData)
	} else {
		return types.NewEngineError(rCtx, msg, errors.New("返回类型不匹配"))
	}
	return nil
}

func (x *EndNode) Destroy() {
}
