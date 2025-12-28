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

package transform

//规则链节点配置示例：
//{
//        "id": "s2",
//        "type": "ExprAssign",
//        "name": "脚本路由",
//        "debugMode": false,
//        "configuration": {
//          "ExprScript": "return ['one','two'];"
//        }
//      }
import (
	"context"
	"errors"
	"reflect"

	"github.com/bittoy/rule/types"
	"github.com/bittoy/rule/utils/maps"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// E x p rAssignReturnFormatErr JavaScript脚本必须返回数组
var ExprAssignReturnFormatErr = errors.New("return the value is not an array")

// init 注册ExprAssignNode组件
func init() {
	Registry.Add(&ExprAssignNode{})
}

// ExprAssignNodeConfiguration ExprAssignNode配置结构
type ExprAssignNodeConfiguration struct {
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
	Script string `json:"script"`
}

// ExprAssignNode 使用JavaScript确定消息路由路径的开关节点
type ExprAssignNode struct {
	// Config 节点配置
	Config ExprAssignNodeConfiguration

	// program 用于高效评估的编译表达式
	// program is the compiled expression for efficient evaluation
	program *vm.Program
}

// Type 返回组件类型
// Type returns the component type identifier.
func (x *ExprAssignNode) Type() types.NodeType {
	return types.RuleSubTypeExprAssign
}

// New 创建新实例
func (x *ExprAssignNode) New() types.Node {
	return &ExprAssignNode{Config: ExprAssignNodeConfiguration{
		Script: `{}`,
	}}
}

// Init 初始化节点
func (x *ExprAssignNode) Init(config types.Config, configuration types.Configuration) error {
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

// OnMsg 处理消息，执行JavaScript脚本确定路由路径
func (x *ExprAssignNode) OnMsg(ctx context.Context, msg types.RuleMsg) (next string, err error) {
	out, err := vm.Run(x.program, msg.GetInput())
	if err != nil {
		return "", err
	}
	if result, ok := out.(map[string]any); ok {
		msg.CopyInnerData(result)
		return types.DefaultRelationType, nil
	}
	return "", errors.New("返回类型不匹配")
}

// Destroy 清理资源
func (x *ExprAssignNode) Destroy() {

}
