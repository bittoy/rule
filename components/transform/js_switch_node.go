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
//        "type": "jsSwitch",
//        "name": "脚本路由",
//        "debugMode": false,
//        "configuration": {
//          "jsScript": "return ['one','two'];"
//        }
//      }
import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bittoy/rule/types"
	"github.com/bittoy/rule/utils/maps"

	"github.com/dop251/goja"
)

// JsSwitchReturnFormatErr JavaScript脚本必须返回数组
var JsSwitchReturnFormatErr = errors.New("return the value is not string")

// init 注册JsSwitchNode组件
func init() {
	Registry.Add(&JsSwitchNode{})
}

// JsSwitchNodeConfiguration JsSwitchNode配置结构
type JsSwitchNodeConfiguration struct {
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
	Script string `json:"script"`
}

// JsSwitchNode 使用JavaScript确定消息路由路径的开关节点
type JsSwitchNode struct {
	// Config 节点配置
	Config JsSwitchNodeConfiguration

	pool *sync.Pool
}

// Type 返回组件类型
func (x *JsSwitchNode) Type() types.NodeType {
	return types.RuleSubTypeJsSwitch
}

// New 创建新实例
func (x *JsSwitchNode) New() types.Node {
	return &JsSwitchNode{Config: JsSwitchNodeConfiguration{
		Script: `return default;`,
	}}
}

// Init 初始化节点
func (x *JsSwitchNode) Init(config types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.Config)
	if err != nil {
		return err
	}

	jsScript := fmt.Sprintf("function jsSwitch(msg) { %s } jsSwitch;", x.Config.Script)
	program, err := goja.Compile("jsSwitch.js", jsScript, true)
	if err != nil {
		return fmt.Errorf("new js vm err: script:%s", jsScript)
	}

	// vars, err := base.NodeUtils.GetVars(configuration)
	// if err != nil {
	// 	return err
	// }

	x.pool = &sync.Pool{
		New: func() any {
			vm := goja.New()
			// 每个 vm 运行时执行 prog
			_, err := vm.RunProgram(program)
			if err != nil {
				panic(fmt.Sprintf("failed to run program in new VM: %v", err))
			}
			return vm
		},
	}
	return nil
}

// OnMsg 处理消息，执行JavaScript脚本确定路由路径
func (x *JsSwitchNode) OnMsg(ctx context.Context, msg types.RuleMsg) (string, error) {
	vm := x.pool.Get().(*goja.Runtime)
	defer x.pool.Put(vm)

	fnVal := vm.Get("jsSwitch")
	if fnVal == nil {
		return "", errors.New("function jsSwitch not found")
	}

	f, ok := goja.AssertFunction(fnVal)
	if !ok {
		return "", errors.New("jsSwitch is not a function")
	}

	// Execute function
	res, err := f(goja.Undefined(), vm.ToValue(msg.GetInput()))
	if err != nil {
		return "", err
	}

	if result, ok := res.Export().(string); ok {
		return result, nil
	}
	return "", JsSwitchReturnFormatErr
}

// Destroy 清理资源
func (x *JsSwitchNode) Destroy() {

}
