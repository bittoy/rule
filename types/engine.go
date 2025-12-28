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

import (
	"context"
)

// RuleEngineOption defines a function type for configuring a RuleEngine.
// It follows the functional options pattern for flexible configuration.
//
// RuleEngineOption 定义了用于配置 RuleEngine 的函数类型。
// 它遵循函数选项模式，提供灵活的配置。
//
// Example usage:
// 使用示例：
//
//	engine, err := rulego.New("myEngine", ruleChainDef,
//		types.WithConfig(myConfig),
//		types.WithAspects(debugAspect, metricsAspect))
type EngineOption func(Engine) error

// RuleEngine is the core interface for a rule engine instance.
// Each RuleEngine manages a single root rule chain and provides methods for
// message processing, configuration updates, and lifecycle management.
//
// RuleEngine 是规则引擎实例的核心接口。
// 每个 RuleEngine 管理一个根规则链，并提供消息处理、配置更新和生命周期管理的方法。
//
// Key Features:
// 主要特性：
//   - Rule chain execution and management  规则链执行和管理
//   - Dynamic configuration reloading  动态配置重载
//   - Aspect-oriented programming support  面向切面编程支持
//   - Performance metrics collection  性能指标收集
//   - Concurrent message processing  并发消息处理
//
// Lifecycle:
// 生命周期：
//  1. Create engine with New() or Load()  使用 New() 或 Load() 创建引擎
//  2. Process messages with OnMsg()  使用 OnMsg() 处理消息
//  3. Update configuration with ReloadSelf()  使用 ReloadSelf() 更新配置
//  4. Clean up resources with Stop()  使用 Stop() 清理资源
type Engine interface {
	// Id returns the unique identifier of the RuleEngine.
	// This ID is used for engine lookup and management within pools.
	// Id 返回 RuleEngine 的唯一标识符。
	// 此 ID 用于池中的引擎查找和管理。
	Id() string

	// SetConfig sets the configuration for the RuleEngine.
	// This affects logging, caching, component registry, and other engine behaviors.
	// SetConfig 设置 RuleEngine 的配置。
	// 这影响日志记录、缓存、组件注册表和其他引擎行为。
	SetConfig(config Config)

	// SetAspects sets the aspects for the RuleEngine.
	// Aspects provide cross-cutting functionality like metrics, debugging, and validation.
	// SetAspects 设置 RuleEngine 的切面。
	// 切面提供如指标、调试和验证等横切功能。
	SetAspects(aspects ...Aspect)

	// ReloadSelf reloads the RuleEngine itself with the given definition and options.
	// This completely replaces the current rule chain with a new configuration.
	// ReloadSelf 使用给定定义和选项重新加载 RuleEngine 本身。
	// 这完全用新配置替换当前规则链。
	ReloadSelf(def []byte) error

	// DSL returns the DSL (Domain Specific Language) representation of the RuleEngine.
	// This provides the complete rule chain configuration in serialized format.
	// DSL 返回 RuleEngine 的 DSL（领域特定语言）表示。
	// 这以序列化格式提供完整的规则链配置。
	DSL() []byte

	// Stop shuts down the RuleEngine and releases all resources.
	// If ctx is provided, it will wait for active messages to complete within the context deadline.
	// If ctx is no deadline, it uses a default 10-second timeout.
	// If ctx is nil, it performs immediate shutdown.
	// Stop 关闭 RuleEngine 并释放所有资源。
	// 如果提供了 ctx，它将在上下文截止时间内等待活跃消息完成。
	// 如果 ctx 没有截止时间，则使用默认的10秒超时。
	// 如果 ctx 为 nil，则执行立即停机。
	Stop()

	// OnMsg processes a message asynchronously with the given context options.
	// This is the primary method for feeding data into the rule engine.
	// OnMsg 使用给定上下文选项异步处理消息。
	// 这是向规则引擎输入数据的主要方法。
	OnMsg(ctx context.Context, msg RuleMsg) error
}
