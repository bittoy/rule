package transform

//规则链节点配置示例：
//{
//        "id": "s2",
//        "type": "jsFilter",
//        "name": "过滤",
//        "debugMode": false,
//        "configuration": {
//          "jsScript": "return msg.temperature > 50;"
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

const (
	// JsFilterFuncName JS函数名
	JsFilterFuncName = "Filter"
	// JsFilterType JsFilter组件类型
	JsFilterType = "jsFilter"
	// JsFilterFuncTemplate JS函数模板
	JsFilterFuncTemplate = "function Filter(msg, metadata, msgType, dataType) { %s }"
)

// init 注册JsFilterNode组件
func init() {
	Registry.Add(&JsFilterNode{})
}

// JsFilterNodeConfiguration JsFilterNode配置结构
type JsFilterNodeConfiguration struct {
	// JsScript JavaScript脚本，用于评估过滤条件
	// 函数参数：msg, metadata, msgType, dataType
	// 必须返回布尔值：true通过过滤，false不通过
	//
	// 内置变量：
	//   - $ctx: 上下文对象，提供缓存操作
	//   - global: 全局配置属性
	//   - vars: 规则链变量
	//   - UDF函数: 用户自定义函数
	//
	// 示例: "return msg.temperature > 25.0;"
	Script string `json:"script"`
}

// JsFilterNode 使用JavaScript评估布尔条件的过滤器节点
type JsFilterNode struct {
	// Config 节点配置
	Config JsFilterNodeConfiguration

	pool *sync.Pool
}

// Type 返回组件类型
func (x *JsFilterNode) Type() types.NodeType {
	return types.RuleSubTypeJsFilter
}

// New 创建新实例
func (x *JsFilterNode) New() types.Node {
	return &JsFilterNode{Config: JsFilterNodeConfiguration{
		Script: "return 1=1;",
	}}
}

// Init 初始化节点
func (x *JsFilterNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.Config)
	if err != nil {
		return err
	}

	jsScript := fmt.Sprintf("function jsFilter(msg) { %s } jsFilter;", x.Config.Script)
	program, err := goja.Compile("jsFilter.js", jsScript, true)
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

// OnMsg 处理消息，执行JavaScript过滤条件
func (x *JsFilterNode) OnMsg(ctx context.Context, rCtx types.RuleContext, msg types.RuleMsg) error {
	vm := x.pool.Get().(*goja.Runtime)
	defer x.pool.Put(vm)

	fnVal := vm.Get("jsSwitch")
	if fnVal == nil {
		return fmt.Errorf("function jsSwitch not found")
	}

	f, ok := goja.AssertFunction(fnVal)
	if !ok {
		return errors.New("jsSwitch is not a function")
	}

	// Execute function
	res, err := f(goja.Undefined(), vm.ToValue(msg.GetInput()))
	if err != nil {
		return types.NewEngineError(rCtx, msg, err)
	}

	if result, ok := res.Export().(bool); ok {
		if result {
			return rCtx.TellNext(ctx, msg, types.TrueRelationType)
		} else {
			return rCtx.TellNext(ctx, msg, types.FalseRelationType)
		}
	}
	return types.NewEngineError(rCtx, msg, JsSwitchReturnFormatErr)
}

// Destroy 清理资源
func (x *JsFilterNode) Destroy() {

}
