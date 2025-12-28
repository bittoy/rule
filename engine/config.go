package engine

import (
	"rule/builtin/aspect"
	"rule/builtin/funcs"
	"rule/types"
)

// 这些切面在初始化期间通过 initBuiltinsAspects() 方法自动添加到规则引擎中。
// 如果提供了自定义切面，除非自定义列表中已存在相同类型的切面，否则仍会包含
// 内置切面。这确保基本功能始终可用，无需显式配置。
var BuiltinsAspects = []types.Aspect{&aspect.ChainAggregationValidator{}, &aspect.ChainValidator{}, &aspect.MetricsAspect{}}

// NewConfig creates a new Config and applies the options.
// It initializes all necessary components with sensible defaults.
//
// NewConfig 创建新的配置并应用选项。
// 它使用合理的默认值初始化所有必要的组件。
//
// Parameters:
// 参数：
//   - opts: Optional configuration functions  可选的配置函数
//
// Returns:
// 返回：
//   - types.Config: Initialized configuration  已初始化的配置
//
// Default components include:
// 默认组件包括：
//   - JSON parser for rule chain definitions  规则链定义的 JSON 解析器
//   - Default component registry with built-in components  包含内置组件的默认组件注册表
//   - User-defined functions registry  用户定义函数注册表
//   - Default cache implementation  默认缓存实现
func NewConfig(opts ...types.Option) types.Config {
	c := types.NewConfig(opts...)
	if c.Parser == nil {
		c.Parser = &JsonParser{}
	}
	if c.ComponentsRegistry == nil {
		c.ComponentsRegistry = Registry
	}
	// register all udfs
	// 注册所有用户定义函数
	for name, f := range funcs.ScriptFunc.GetAll() {
		c.RegisterUdf(name, f)
	}
	return c
}

// WithConfig is an option that sets the Config of the RuleEngine.
// WithConfig 是设置 RuleEngine 配置的选项。
func WithConfig(config types.Config) types.EngineOption {
	return func(re types.Engine) error {
		re.SetConfig(config)
		return nil
	}
}

// WithAspects creates a RuleEngineOption to set the aspects of a RuleEngine.
// Aspects provide AOP (Aspect-Oriented Programming) capabilities for cross-cutting concerns
// like logging, metrics, validation, and debugging.
//
// WithAspects 创建一个 RuleEngineOption 来设置 RuleEngine 的切面。
// 切面为日志记录、指标、验证和调试等横切关注点提供 AOP（面向切面编程）功能。
func WithAspects(aspects ...types.Aspect) types.EngineOption {
	return func(re types.Engine) error {
		re.SetAspects(aspects...) // Apply the provided aspects to the RuleEngine.
		return nil                // Return no error.
	}
}
