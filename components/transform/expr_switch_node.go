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

package transform

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/bittoy/rule/types"
	"github.com/bittoy/rule/utils/maps"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

func init() {
	Registry.Add(&ExprSwitchNode{})
}

// SwitchNodeConfiguration SwitchNode配置结构
// SwitchNodeConfiguration defines the configuration structure for the SwitchNode component.
type ExprSwitchNodeConfiguration struct {
	// JsScript JavaScript脚本，用于确定消息路由路径
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
	// student=="3" ? "A" : ((score > 75 && level == "B")|| student == "C") ? "B" : (score > 60) ? "C" : "Default"
	Script string `json:"script"`

	Cases []types.Case `json:"cases"`
}

// SwitchNode 基于表达式评估提供条件消息路由的过滤组件
// SwitchNode provides conditional message routing based on expression evaluation.
//
// 核心算法：
// Core Algorithm:
// 1. 初始化时编译所有case表达式为优化程序 - Compile all case expressions to optimized programs during initialization
// 2. 按顺序评估每个case表达式 - Evaluate each case expression sequentially
// 3. 第一个评估为true的case决定路由 - First case that evaluates to true determines routing
// 4. 无匹配时路由到"Default"关系 - Route to "Default" relation if no matches
//
// 评估逻辑 - Evaluation logic:
//   - 按配置中出现的顺序评估case - Cases evaluated in configuration order
//   - 在第一个成功匹配时停止评估 - Evaluation stops at first successful match
//   - 布尔true结果触发路由到case关系 - Boolean true result triggers routing to case relation
//   - 无匹配导致路由到默认关系 - No matches result in routing to default relation
//
// 表达式语言特性 - Expression language features:
//   - 算术运算符：+, -, *, /, % - Arithmetic operators
//   - 比较运算符：==, !=, <, <=, >, >= - Comparison operators
//   - 逻辑运算符：&&, ||, ! - Logical operators
//   - 字符串操作：contains, startsWith, endsWith - String operations
//   - 数学函数：abs, ceil, floor, round - Mathematical functions
//
// 性能优化 - Performance optimization:
//   - 表达式在初始化期间编译一次 - Expressions compiled once during initialization
//   - 早期终止减少不必要的评估 - Early termination reduces unnecessary evaluations
//   - 按概率排序case以获得最佳性能 - Order cases by probability for optimal performance
type ExprSwitchNode struct {
	// Config 开关节点配置
	// Config holds the switch node configuration
	Config ExprSwitchNodeConfiguration

	// program 用于高效评估的编译表达式
	// program is the compiled expression for efficient evaluation
	program *vm.Program
}

// Type 返回组件类型
// Type returns the component type identifier.
func (x *ExprSwitchNode) Type() types.NodeType {
	return types.RuleSubTypeExprSwitch
}

// New 创建新实例
// New creates a new instance.
func (x *ExprSwitchNode) New() types.Node {
	return &ExprSwitchNode{Config: ExprSwitchNodeConfiguration{
		Script: "",
	}}
}

// Init 初始化组件，编译所有case表达式
// Init initializes the component.
func (x *ExprSwitchNode) Init(config types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.Config)
	if err != nil {
		return err
	}

	var script = strings.TrimSpace(x.Config.Script)
	if len(script) == 0 {
		caseScript, err := genExprScriptByCases(x.Config.Cases)
		if err != nil {
			return err
		}
		script = caseScript
	}

	program, err := expr.Compile(script, expr.AllowUndefinedVariables(), expr.AsKind(reflect.String))
	if err != nil {
		return err
	}

	x.program = program

	return nil
}

// OnMsg 处理消息，按顺序评估case表达式并路由到第一个匹配的case或默认关系
// OnMsg processes incoming messages by evaluating case expressions sequentially.
func (x *ExprSwitchNode) OnMsg(ctx context.Context, msg types.RuleMsg) (string, error) {
	out, err := vm.Run(x.program, msg.GetInput())
	if err != nil {
		return "", err
	}
	if result, ok := out.(string); ok {
		return result, nil
	}
	return "", errors.New("返回类型不匹配")
}

// Destroy 清理资源
// Destroy cleans up resources.
func (x *ExprSwitchNode) Destroy() {
}
