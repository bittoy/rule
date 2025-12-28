/*
 * Copyright 2024 The RuleGo Authors.
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

// Option is a function type that modifies the Config.
// Option 是修改 Config 的函数类型。
//
// The Option pattern provides a flexible and extensible way to configure RuleGo instances.
// It allows users to specify only the configuration aspects they need while maintaining
// default values for other settings.
// Option 模式提供了配置 RuleGo 实例的灵活和可扩展方式。
// 它允许用户仅指定他们需要的配置方面，同时为其他设置保持默认值。
//
// Usage Pattern:
// 使用模式：
//
//	config := NewConfig(
//	    WithPool(customPool),
//	    WithLogger(customLogger),
//	    WithOnDebug(debugHandler),
//	)
type Option func(*Config) error

// WithComponentsRegistry is an option that sets the components' registry of the Config.
// WithComponentsRegistry 是设置 Config 组件注册表的选项。
//
// The components registry manages all available node types that can be used in rule chains.
// Setting a custom registry allows for component isolation, versioning, and custom component sets.
// 组件注册表管理规则链中可使用的所有可用节点类型。
// 设置自定义注册表允许组件隔离、版本控制和自定义组件集。
//
// Use Cases:
// 使用案例：
//   - Multi-tenant applications with different component sets per tenant
//     每个租户具有不同组件集的多租户应用程序
//   - Plugin-based architectures with dynamic component loading
//     具有动态组件加载的基于插件的架构
//   - Testing environments with mock components
//     具有模拟组件的测试环境
//   - Component versioning and A/B testing
//     组件版本控制和 A/B 测试
//
// Example:
// 示例：
//
//	registry := &MyCustomRegistry{}
//	registry.Register(&MyCustomNode{})
//	config := NewConfig(WithComponentsRegistry(registry))
func WithComponentsRegistry(componentsRegistry ComponentRegistry) Option {
	return func(c *Config) error {
		c.ComponentsRegistry = componentsRegistry
		return nil
	}
}

// WithParser is an option that sets the parser of the Config.
// WithParser 是设置 Config 解析器的选项。
//
// The parser converts rule chain definitions from various formats (JSON, YAML, XML)
// into internal data structures. Custom parsers enable support for different
// configuration languages and specialized formats.
// 解析器将规则链定义从各种格式（JSON、YAML、XML）转换为内部数据结构。
// 自定义解析器支持不同的配置语言和专用格式。
//
// Parser Capabilities:
// 解析器功能：
//   - Bidirectional conversion (encode/decode)
//     双向转换（编码/解码）
//   - Format validation and error reporting
//     格式验证和错误报告
//   - Custom extension support
//     自定义扩展支持
//   - Schema validation
//     模式验证
//
// Common Parser Types:
// 常见解析器类型：
//   - JSON Parser: Default, widely supported
//     JSON 解析器：默认，广泛支持
//   - YAML Parser: Human-readable, configuration-friendly
//     YAML 解析器：人类可读，配置友好
//   - XML Parser: Enterprise integration, legacy systems
//     XML 解析器：企业集成，传统系统
//   - Binary Parser: Performance-optimized formats
//     二进制解析器：性能优化格式
//
// Custom Parser Implementation:
// 自定义解析器实现：
// Implement the Parser interface to support custom formats or add
// validation, transformation, or encryption capabilities.
// 实现 Parser 接口以支持自定义格式或添加验证、转换或加密功能。
//
// Example:
// 示例：
//
//	// YAML parser for configuration-friendly format
//	// 用于配置友好格式的 YAML 解析器
//	yamlParser := &YamlParser{}
//	config := NewConfig(WithParser(yamlParser))
//
//	// Encrypted parser for sensitive configurations
//	// 用于敏感配置的加密解析器
//	encryptedParser := &EncryptedJsonParser{Key: secretKey}
//	config := NewConfig(WithParser(encryptedParser))
func WithParser(parser Parser) Option {
	return func(c *Config) error {
		c.Parser = parser
		return nil
	}
}

// WithLogger is an option that sets the logger of the Config.
// WithLogger 是设置 Config 日志记录器的选项。
//
// Example:
// 示例：
//
//	// Logrus integration
//	// Logrus 集成
//	logrusLogger := &LogrusLogger{Logger: logrus.New()}
//	config := NewConfig(WithLogger(logrusLogger))
//
//	// Custom logger with monitoring integration
//	// 具有监控集成的自定义日志记录器
//	monitoringLogger := &MonitoringLogger{Service: "rulego"}
//	config := NewConfig(WithLogger(monitoringLogger))
func WithLogger(logger Logger) Option {
	return func(c *Config) error {
		c.Logger = logger
		return nil
	}
}

func WithProperties(properties Properties) Option {
	return func(c *Config) error {
		c.Properties = properties
		return nil
	}
}

type CallbackOption func(*Callbacks) error

func NewCallbacks(opts ...CallbackOption) Callbacks {
	c := &Callbacks{}

	for _, opt := range opts {
		_ = opt(c)
	}
	return *c
}

func WithOnNew(onNew OnNew) CallbackOption {
	return func(c *Callbacks) error {
		c.OnNew = onNew
		return nil
	}
}

func WithOnUpdated(onUpdated OnUpdated) CallbackOption {
	return func(c *Callbacks) error {
		c.OnUpdated = onUpdated
		return nil
	}
}

func WithOnDeleted(onDeleted OnDeleted) CallbackOption {
	return func(c *Callbacks) error {
		c.OnDeleted = onDeleted
		return nil
	}
}
