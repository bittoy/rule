package transform

//规则链节点配置示例：
//{
//        "id": "s1",
//        "type": "exprFilter",
//        "name": "表达式过滤器",
//        "debugMode": false,
//        "configuration": {
//          "expr": "msg.temperature > 50"
//        }
//      }
import (
	"context"
	"errors"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"rule/types"
	"rule/utils/maps"
)

// init 注册ExprFilterNode组件
// init registers the ExprFilterNode component with the default registry.
func init() {
	Registry.Add(&ExprFilterNode{})
}

// ExprFilterNodeConfiguration ExprFilterNode配置结构
// ExprFilterNodeConfiguration defines the configuration structure for the ExprFilterNode component.
type ExprFilterNodeConfiguration struct {
	// Expr 用于过滤评估的表达式，必须返回布尔值
	// Expr contains the expression to evaluate for filtering.
	// The expression has access to the following variables:
	//   - id: Message ID (string)
	//   - ts: Message timestamp (int64)
	//   - data: Original message data (string)
	//   - msg: Parsed message body (object for JSON, string otherwise)
	//   - metadata: Message metadata (object with key-value pairs)
	//   - type: Message type (string)
	//   - dataType: Message data type (string)
	//
	// The expression must evaluate to a boolean value:
	//   - true: Message passes the filter (routed to "True" relation)
	//   - false: Message fails the filter (routed to "False" relation)
	//
	// Example expressions:
	// 表达式示例：
	//   - "msg.temperature > 50"
	//   - "metadata.deviceType == 'sensor' && msg.value > 100"
	//   - "type == 'TELEMETRY' && data contains 'alarm'"
	//   - "ts > 1640995200 && msg.status == 'active'"
	Script string `json:"script"`
}

// ExprFilterNode 使用expr-lang表达式进行布尔评估来过滤消息的过滤组件
// ExprFilterNode filters messages using expr-lang expressions for boolean evaluation.
//
// 核心算法：
// Core Algorithm:
// 1. 初始化时编译表达式为优化的程序 - Compile expression to optimized program during initialization
// 2. 准备消息评估环境（id、ts、data、msg、metadata等）- Prepare message evaluation environment
// 3. 执行编译的表达式程序 - Execute compiled expression program
// 4. 根据布尔结果路由消息到True/False关系 - Route message to True/False relation based on boolean result
//
// 表达式语言特性 - Expression language features:
//   - 算术运算符：+, -, *, /, % - Arithmetic operators
//   - 比较运算符：==, !=, <, <=, >, >= - Comparison operators
//   - 逻辑运算符：&&, ||, ! - Logical operators
//   - 字符串操作：contains, startsWith, endsWith - String operations
//   - 数学函数：abs, ceil, floor, round - Mathematical functions
type ExprFilterNode struct {
	// Config 表达式过滤器配置
	// Config holds the expression filter configuration
	Config ExprFilterNodeConfiguration

	// program 用于高效评估的编译表达式
	// program is the compiled expression for efficient evaluation
	program *vm.Program
}

// Type 返回组件类型
// Type returns the component type identifier.
func (x *ExprFilterNode) Type() types.NodeType {
	return types.RuleSubTypeExprFilter
}

// New 创建新实例
// New creates a new instance.
func (x *ExprFilterNode) New() types.Node {
	return &ExprFilterNode{Config: ExprFilterNodeConfiguration{
		Script: "1==1",
	}}
}

// Init 初始化组件，验证并编译表达式
// Init initializes the component.
func (x *ExprFilterNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.Config)
	if err != nil {
		return err
	}

	program, err := expr.Compile(x.Config.Script, expr.AllowUndefinedVariables(), expr.AsBool())
	if err != nil {
		return err
	}

	x.program = program

	return nil
}

// OnMsg 处理消息，通过评估编译的表达式来过滤消息
// OnMsg processes incoming messages by evaluating the compiled expression.
func (x *ExprFilterNode) OnMsg(ctx context.Context, rCtx types.RuleContext, msg types.RuleMsg) error {
	out, err := vm.Run(x.program, msg.GetInput())
	if err != nil {
		return types.NewEngineError(rCtx, msg, err)
	}
	if result, ok := out.(bool); ok {
		if result {
			return rCtx.TellNext(ctx, msg, types.TrueRelationType)
		} else {
			return rCtx.TellNext(ctx, msg, types.FalseRelationType)
		}

	} else {
		return types.NewEngineError(rCtx, msg, errors.New("返回类型不匹配"))
	}

}

// Destroy 清理资源
// Destroy cleans up resources.
func (x *ExprFilterNode) Destroy() {
}
