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

package types

// Config defines the configuration for the rule engine.
// Config 定义规则引擎的配置。
//
// This structure contains all the necessary configuration parameters for initializing
// and running a RuleGo rule engine instance. It provides control over execution behavior,
// resource management, debugging, scripting, and integration with external systems.
// 此结构包含初始化和运行 RuleGo 规则引擎实例所需的所有配置参数。
// 它提供对执行行为、资源管理、调试、脚本和与外部系统集成的控制。
//
// Configuration Categories:
// Usage Example:
// 使用示例：
//
//	config := NewConfig(
//	    WithPool(myPool),
//	    WithLogger(myLogger),
//	    WithOnDebug(debugHandler),
//	)
//	engine := rulego.New("chainId", chainDSL, rulego.WithConfig(config))
type Config struct {
	// ComponentsRegistry is the component registry for managing available rule chain components.
	// ComponentsRegistry 是管理可用规则链组件的组件注册表。
	//
	// 主要特性 - Key Features:
	//   - 组件隔离：支持不同引擎实例使用独立的组件集合 - Component isolation: supports different engine instances using independent component sets
	//   - 动态管理：运行时注册/注销组件和插件加载 - Dynamic management: runtime component registration/unregistration and plugin loading
	//   - 可视化支持：提供UI配置工具所需的组件元数据 - Visual support: provides component metadata for UI configuration tools
	//
	// 配置示例 - Configuration Examples:
	//
	//	// 使用自定义组件注册表 - Use custom component registry
	//	customRegistry := components.NewRegistry()
	//	customRegistry.Register(&MyCustomNode{})
	//	config := rulego.NewConfig(types.WithComponentsRegistry(customRegistry))
	//
	//	// 动态加载插件 - Dynamic plugin loading
	//	registry.RegisterPlugin("myPlugin", "./plugins/custom.so")
	//
	// 默认使用 `rulego.Registry`，包含所有标准组件。详细功能请参见 ComponentRegistry 接口文档。
	// Defaults to `rulego.Registry` with all standard components. See ComponentRegistry interface for detailed functionality.
	//
	ComponentsRegistry ComponentRegistry
	// Parser is the rule chain parser interface, defaulting to `rulego.JsonParser`.
	// Parser 是规则链解析器接口，默认为 `rulego.JsonParser`。
	//
	// The parser converts rule chain definitions from various formats (JSON, YAML, XML)
	// into internal data structures that the engine can execute.
	// 解析器将规则链定义从各种格式（JSON、YAML、XML）转换为
	// 引擎可以执行的内部数据结构。
	//
	// Custom parsers can be implemented to support:
	// 可以实现自定义解析器来支持：
	//   - Domain-specific configuration languages
	//     领域特定的配置语言
	//   - Legacy configuration formats
	//     传统配置格式
	//   - Compressed or encrypted rule definitions
	//     压缩或加密的规则定义
	//   - Runtime rule generation from databases
	//     从数据库运行时生成规则
	Parser Parser
	// Logger is the logging interface, defaulting to `DefaultLogger()`.
	// Logger 是日志接口，默认为 `DefaultLogger()`。
	//
	// The logger provides structured logging capabilities for the rule engine,
	// supporting different log levels and output formats.
	// 日志记录器为规则引擎提供结构化日志功能，
	Logger Logger
	// Properties are global properties in key-value format.
	// Rule chain node configurations can replace values with ${global.propertyKey}.
	// Replacement occurs during node initialization and only once.
	// Properties 是键值格式的全局属性。
	// 规则链节点配置可以用 ${global.propertyKey} 替换值。
	// 替换在节点初始化期间发生，只发生一次。
	//
	// Example usage in rule configuration:
	// 规则配置中的示例用法：
	//   {
	//     "type": "restApiCall",
	//     "configuration": {
	//       "restEndpointUrlPattern": "${global.apiBaseUrl}/users"
	//     }
	//   }
	Properties Properties
	// Udf is a map for registering custom Golang functions and native scripts that can be called at runtime by script engines like JavaScript Lua.
	// Function names can be repeated for different script types.
	// Udf 是用于注册自定义 Golang 函数和原生脚本的映射，可以在运行时被 JavaScript Lua 等脚本引擎调用。
	// 不同脚本类型的函数名可以重复。
	//
	// UDF (User Defined Functions) extend the scripting capabilities by providing:
	// UDF（用户定义函数）通过提供以下功能扩展脚本功能：
	//   - Access to Go standard library functions
	//     访问 Go 标准库函数
	//   - Custom business logic implementation
	//     自定义业务逻辑实现
	//   - Integration with external systems
	//     与外部系统集成
	//   - Performance-critical operations in native code
	//     原生代码中的性能关键操作
	//
	// Function registration example:
	// 函数注册示例：
	//   config.RegisterUdf("encrypt", func(data string) string {
	//       // Custom encryption logic
	//       return encryptedData
	//   })
	Udf map[string]interface{}
}

// RegisterUdf registers a custom function. Function names can be repeated for different script types.
// RegisterUdf 注册自定义函数。不同脚本类型的函数名可以重复。
//
// This method provides a convenient way to register User Defined Functions (UDFs) that can be
// called from script components. It handles function name resolution and conflict prevention
// for different script engines.
// 此方法提供了注册用户定义函数（UDF）的便捷方式，这些函数可以从脚本组件中调用。
// 它处理不同脚本引擎的函数名解析和冲突预防。
//
// Function Registration Process:
// 函数注册过程：
//  1. Initialize Udf map if not already created
//     如果尚未创建则初始化 Udf 映射
//  2. Check if value is a Script type with specific engine
//     检查值是否是具有特定引擎的 Script 类型
//  3. Resolve naming conflicts using script type prefixes
//     使用脚本类型前缀解决命名冲突
//  4. Store function with resolved name
//     使用解析的名称存储函数
//
// Examples:
// 示例：
//
//	// Register a Go function for all script types
//	// 为所有脚本类型注册 Go 函数
//	config.RegisterUdf("stringUtils", myStringUtilsFunc)
//
//	// Register a JavaScript-specific function
//	// 注册 JavaScript 特定函数
//	config.RegisterUdf("jsHelper", Script{
//	    Type: "Js",
//	    Content: "function jsHelper(data) { return data.toUpperCase(); }"
//	})
//
//	// Register a Lua-specific function
//	// 注册 Lua 特定函数
//	config.RegisterUdf("luaHelper", Script{
//	    Type: "Lua",
//	    Content: "function luaHelper(data) return string.upper(data) end"
//	})
func (c *Config) RegisterUdf(name string, value interface{}) {
	if c.Udf == nil {
		c.Udf = make(map[string]interface{})
	}
	if script, ok := value.(Script); ok {
		if script.Type != AllScript {
			// Resolve function name conflicts for different script types.
			// 解决不同脚本类型的函数名冲突。
			name = script.Type + ScriptFuncSeparator + name
		}
	}
	c.Udf[name] = value
}

// NewConfig creates a new Config with default values and applies the provided options.
// NewConfig 创建具有默认值的新 Config 并应用提供的选项。
//
// This function implements the functional options pattern, allowing for flexible
// and extensible configuration. It sets reasonable defaults while enabling
// 此函数实现函数式选项模式，允许灵活和可扩展的配置。
//
// Usage Examples:
// 使用示例：
//
//	// Basic configuration with defaults
//	// 具有默认值的基本配置
//	config := NewConfig()
//
//	// Configuration with custom options
//	// 具有自定义选项的配置
//	config := NewConfig(
//	    WithPool(customPool),
//	    WithLogger(customLogger),
//	    WithScriptMaxExecutionTime(5 * time.Second),
//	    WithEndpointEnabled(false),
//	)
func NewConfig(opts ...Option) Config {
	c := &Config{
		Logger:     DefaultLogger(),
		Properties: NewProperties(),
	}

	for _, opt := range opts {
		_ = opt(c)
	}
	return *c
}
