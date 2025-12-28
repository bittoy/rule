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

import (
	"sort"
)

// Aspect defines the base interface for implementing Aspect-Oriented Programming (AOP) in RuleGo.
// AOP provides cross-cutting functionality that can intercept and enhance rule chain execution
// without modifying the original business logic of components.
//
// Aspect 定义在 RuleGo 中实现面向切面编程（AOP）的基础接口。
// AOP 提供横切功能，可以拦截和增强规则链执行而不修改组件的原始业务逻辑。
//
// Engine Instance Level:
// 引擎实例级别：
//
// Aspects are registered at the rule engine level and each engine instance gets its own
// aspect instances through the New() method during initialization. This ensures proper
// isolation between different rule engine instances.
//
// 切面在规则引擎级别注册，每个引擎实例在初始化期间通过 New() 方法获得自己的
// 切面实例。这确保了不同规则引擎实例之间的适当隔离。
//
// Aspect Categories:
// 切面类别：
//
//   - Engine Lifecycle Aspects: OnChainBeforeInit, OnNodeBeforeInit, OnCreated, OnReload, OnDestroy
//     引擎生命周期切面：OnChainBeforeInit、OnNodeBeforeInit、OnCreated、OnReload、OnDestroy
//   - Chain Execution Aspects: Start, End, Completed
//     链执行切面：Start、End、Completed
//   - Node Execution Aspects: Before, After, Around
//     节点执行切面：Before、After、Around
//
// Execution Order:
// 执行顺序：
//
//  1. Engine Level (during rule engine operations):
//     引擎级别（规则引擎操作期间）：
//     OnChainBeforeInit -> OnNodeBeforeInit -> OnCreated -> OnReload -> OnDestroy
//
//  2. Chain Level (for each message processing):
//     链级别（每次消息处理）：
//     Start (onStart) -> [Node Processing] -> End (onEnd) -> Completed (onAllNodeCompleted)
//
//  3. Node Level (for each node execution in rule_context.go executeAroundAop):
//     节点级别（rule_context.go executeAroundAop 中每个节点执行）：
//     Before -> Around -> [Node.OnMsg] -> After
//
// Built-in Aspects:
// 内置切面：
//
// RuleGo includes built-in aspects that are automatically registered:
// RuleGo 包含自动注册的内置切面：
//   - Validator: Data validation and schema checking
//     Validator：数据验证和模式检查
//   - Debug: Debug information collection and logging
//     Debug：调试信息收集和日志记录
//   - MetricsAspect: Performance metrics and monitoring
//     MetricsAspect：性能指标和监控
type Aspect interface {
	// Order returns the execution priority of the aspect.
	// Lower values indicate higher priority and earlier execution in the aspect chain.
	//
	// Order 返回切面的执行优先级。
	// 较小的值表示更高的优先级和在切面链中更早的执行。
	//
	// Returns:
	// 返回：
	//   - int: Priority value, lower numbers execute first
	//     int：优先级值，较小的数字先执行
	Order() int

	// New creates a new instance of the aspect for a specific rule engine instance.
	// This method is called during rule engine initialization (in initBuiltinsAspects and initChain)
	// to ensure each rule engine has its own aspect instance with isolated state.
	//
	// New 为特定的规则引擎实例创建切面的新实例。
	// 此方法在规则引擎初始化期间调用（在 initBuiltinsAspects 和 initChain 中），
	// 确保每个规则引擎都有自己的切面实例和隔离的状态。
	//
	// Implementation Requirements:
	// 实现要求：
	//   - Create a completely independent instance
	//     创建完全独立的实例
	//   - Copy necessary configuration
	//     复制必要的配置
	//   - Ensure no shared mutable state between instances
	//     确保实例之间没有共享的可变状态
	//
	// Returns:
	// 返回：
	//   - Aspect: New aspect instance for the rule engine
	//     Aspect：规则引擎的新切面实例
	New() Aspect
}

// NodeAspect defines the base interface for aspects that operate at the individual node level.
// These aspects can intercept and modify the execution of specific nodes based on PointCut criteria.
//
// NodeAspect 定义在单个节点级别操作的切面的基础接口。
// 这些切面可以基于 PointCut 条件拦截和修改特定节点的执行。
//
// Node aspects are executed during message processing through nodes and provide
// fine-grained control over individual node behavior.
//
// 节点切面在消息通过节点处理期间执行，提供对单个节点行为的细粒度控制。
type NodeAspect interface {
	Aspect

	// PointCut determines whether this aspect should be applied to a specific node execution.
	// This method enables selective aspect application based on runtime conditions.
	//
	// PointCut 确定此切面是否应应用于特定的节点执行。
	// 此方法基于运行时条件启用选择性切面应用。
	//
	// Parameters:
	// 参数：
	//   - ctx: Rule execution context
	//     ctx：规则执行上下文
	//   - msg: Message being processed
	//     msg：正在处理的消息
	//   - relationType: Connection type between nodes
	//     relationType：节点间的连接类型
	//
	// Returns:
	// 返回：
	//   - bool: true to apply aspect, false to skip
	//     bool：true 应用切面，false 跳过
	PointCut(ctx RuleContext, msg RuleMsg, relationType string) bool
}

type ChainAspect interface {
	Aspect

	// PointCut determines whether this aspect should be applied to a specific node execution.
	// This method enables selective aspect application based on runtime conditions.
	//
	// PointCut 确定此切面是否应应用于特定的节点执行。
	// 此方法基于运行时条件启用选择性切面应用。
	//
	// Parameters:
	// 参数：
	//   - ctx: Rule execution context
	//     ctx：规则执行上下文
	//   - msg: Message being processed
	//     msg：正在处理的消息
	//   - relationType: Connection type between nodes
	//     relationType：节点间的连接类型
	//
	// Returns:
	// 返回：
	//   - bool: true to apply aspect, false to skip
	//     bool：true 应用切面，false 跳过
	PointCut(ctx RuleContext, msg RuleMsg) bool
}

// BeforeAspect defines the interface for aspects that execute before node message processing.
// These aspects are executed in rule_context.go executeAroundAop() before the node's OnMsg method.
//
// BeforeAspect 定义在节点消息处理之前执行的切面接口。
// 这些切面在 rule_context.go executeAroundAop() 中节点的 OnMsg 方法之前执行。
//
// Execution Flow:
// 执行流程：
//  1. Message arrives at node
//     消息到达节点
//  2. BeforeAspect.Before() is called
//     调用 BeforeAspect.Before()
//  3. Modified message is passed to node OnMsg()
//     修改后的消息传递给节点 OnMsg()
type NodeBeforeAspect interface {
	NodeAspect

	// Before is executed before the node's OnMsg method processes the message.
	// The returned message will be used as input for the node's OnMsg method.
	//
	// Before 在节点的 OnMsg 方法处理消息之前执行。
	// 返回的消息将用作节点 OnMsg 方法的输入。
	//
	// Parameters:
	// 参数：
	//   - ctx: Rule execution context
	//     ctx：规则执行上下文
	//   - msg: Original message to be processed
	//     msg：要处理的原始消息
	//   - relationType: Connection type that led to this node execution
	//     relationType：导致此节点执行的连接类型
	//
	// Returns:
	// 返回：
	//   - RuleMsg: Modified message for node processing
	//     RuleMsg：用于节点处理的修改后消息
	Before(ctx RuleContext, msg RuleMsg, relationType string) (RuleMsg, error)
}

// AfterAspect defines the interface for aspects that execute after node message processing.
// These aspects are executed in rule_context.go executeAfterAop() after the node's OnMsg method.
//
// AfterAspect 定义在节点消息处理之后执行的切面接口。
// 这些切面在 rule_context.go executeAfterAop() 中节点的 OnMsg 方法之后执行。
//
// Execution Flow:
// 执行流程：
//  1. Node processes message with OnMsg()
//     节点使用 OnMsg() 处理消息
//  2. AfterAspect.After() is called with result/error
//     使用结果/错误调用 AfterAspect.After()
//  3. Modified message is passed to next node
//     修改后的消息传递给下一个节点
type NodeAfterAspect interface {
	NodeAspect

	// After is executed after the node's OnMsg method completes processing.
	// The returned message will be used for subsequent processing.
	//
	// After 在节点的 OnMsg 方法完成处理后执行。
	// 返回的消息将用于后续处理。
	//
	// Parameters:
	// 参数：
	//   - ctx: Rule execution context
	//     ctx：规则执行上下文
	//   - msg: Message that was processed by the node
	//     msg：被节点处理的消息
	//   - err: Error returned by the node processing, nil if successful
	//     err：节点处理返回的错误，成功时为 nil
	//   - relationType: Connection type for the next node execution
	//     relationType：下一个节点执行的连接类型
	//
	// Returns:
	// 返回：
	//   - RuleMsg: Modified message for next processing
	//     RuleMsg：用于下一步处理的修改后消息
	After(ctx RuleContext, msg RuleMsg, relationType string) (RuleMsg, error)
}

type ChainBeforeAspect interface {
	ChainAspect
	Before(ctx RuleContext, msg RuleMsg) (RuleMsg, error)
}

// EndAspect defines the interface for aspects executed when a rule chain branch ends.
// These aspects are called in engine.go onEnd() method when a branch of execution completes.
//
// EndAspect 定义在规则链分支结束时执行的切面接口。
// 这些切面在 engine.go onEnd() 方法中当执行分支完成时调用。
type ChainAfterAspect interface {
	ChainAspect
	After(ctx RuleContext, msg RuleMsg) (RuleMsg, error)
}

// StartAspect defines the interface for aspects executed before rule chain message processing.
// These aspects are called in engine.go onStart() method before any node processing begins.
//
// StartAspect 定义在规则链消息处理之前执行的切面接口。
// 这些切面在 engine.go onStart() 方法中在任何节点处理开始之前调用。
type ChainAggregationBeforeAspect interface {
	ChainBeforeAspect
}

// EndAspect defines the interface for aspects executed when a rule chain branch ends.
// These aspects are called in engine.go onEnd() method when a branch of execution completes.
//
// EndAspect 定义在规则链分支结束时执行的切面接口。
// 这些切面在 engine.go onEnd() 方法中当执行分支完成时调用。
type ChainAggregationAfterAspect interface {
	ChainAfterAspect
}

// OnNodeBeforeInitAspect defines the interface for aspects executed before rule node initialization.
// These aspects are called during node initialization in the rule chain setup process.
//
// OnNodeBeforeInitAspect 定义在规则节点初始化之前执行的切面接口。
// 这些切面在规则链设置过程中的节点初始化期间调用。
type OnNodeBeforeInitAspect interface {
	NodeAspect

	// OnNodeBeforeInit is executed before rule node initialization.
	// If an error is returned, the node creation will fail.
	//
	// OnNodeBeforeInit 在规则节点初始化之前执行。
	// 如果返回错误，节点创建将失败。
	//
	// Parameters:
	// 参数：
	//   - config: Rule engine configuration
	//     config：规则引擎配置
	//   - def: Rule node definition to be initialized
	//     def：要初始化的规则节点定义
	//
	// Returns:
	// 返回：
	//   - error: Error to prevent node creation, nil to continue
	//     error：阻止节点创建的错误，nil 表示继续
	OnNodeBeforeInit(config Config, def *BaseInfo) error
}

// OnChainBeforeInitAspect defines the interface for aspects executed before rule chain initialization.
// These aspects are called in engine.go initChain() method before the rule chain is created.
//
// OnChainBeforeInitAspect 定义在规则链初始化之前执行的切面接口。
// 这些切面在 engine.go initChain() 方法中规则链创建之前调用。
type OnChainBeforeInitAspect interface {
	NodeAspect

	// OnChainBeforeInit is executed before rule chain initialization.
	// If an error is returned, the chain creation will fail.
	//
	// OnChainBeforeInit 在规则链初始化之前执行。
	// 如果返回错误，链创建将失败。
	//
	// Parameters:
	// 参数：
	//   - config: Rule engine configuration
	//     config：规则引擎配置
	//   - def: Rule chain definition to be initialized
	//     def：要初始化的规则链定义
	//
	// Returns:
	// 返回：
	//   - error: Error to prevent chain creation, nil to continue
	//     error：阻止链创建的错误，nil 表示继续
	OnChainBeforeInit(config Config, def *Chain) error
}

// OnChainBeforeInitAspect defines the interface for aspects executed before rule chain initialization.
// These aspects are called in engine.go initChain() method before the rule chain is created.
//
// OnChainBeforeInitAspect 定义在规则链初始化之前执行的切面接口。
// 这些切面在 engine.go initChain() 方法中规则链创建之前调用。
type OnChainAggregationBeforeInitAspect interface {
	NodeAspect

	// OnChainBeforeInit is executed before rule chain initialization.
	// If an error is returned, the chain creation will fail.
	//
	// OnChainBeforeInit 在规则链初始化之前执行。
	// 如果返回错误，链创建将失败。
	//
	// Parameters:
	// 参数：
	//   - config: Rule engine configuration
	//     config：规则引擎配置
	//   - def: Rule chain definition to be initialized
	//     def：要初始化的规则链定义
	//
	// Returns:
	// 返回：
	//   - error: Error to prevent chain creation, nil to continue
	//     error：阻止链创建的错误，nil 表示继续
	OnChainAggregationBeforeInit(config Config, def *ChainAggregation) error
}

type AspectList []Aspect

// GetChainAspects 获取规则链执行类型增强点切面列表
func (list AspectList) GetChainAspects() ([]ChainBeforeAspect, []ChainAfterAspect) {
	//从小到大排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order() < list[j].Order()
	})

	var beforeAspects []ChainBeforeAspect
	var afterAspects []ChainAfterAspect
	for _, item := range list {
		if a, ok := item.(ChainBeforeAspect); ok {
			beforeAspects = append(beforeAspects, a)
		}
		if a, ok := item.(ChainAfterAspect); ok {
			afterAspects = append(afterAspects, a)
		}
	}

	return beforeAspects, afterAspects
}

// GetChainAspects 获取规则链执行类型增强点切面列表
func (list AspectList) GetChainAggregationAspects() ([]ChainAggregationBeforeAspect, []ChainAggregationAfterAspect) {
	//从小到大排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order() < list[j].Order()
	})

	var beforeAspects []ChainAggregationBeforeAspect
	var afterAspects []ChainAggregationAfterAspect
	for _, item := range list {
		if a, ok := item.(ChainAggregationBeforeAspect); ok {
			beforeAspects = append(beforeAspects, a)
		}
		if a, ok := item.(ChainAggregationAfterAspect); ok {
			afterAspects = append(afterAspects, a)
		}
	}

	return beforeAspects, afterAspects
}

// GetNodeAspects 获取节点执行类型增强点切面列表
func (list AspectList) GetNodeAspects() ([]NodeBeforeAspect, []NodeAfterAspect) {

	//从小到大排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order() < list[j].Order()
	})

	var beforeAspects []NodeBeforeAspect
	var afterAspects []NodeAfterAspect

	for _, item := range list {
		if a, ok := item.(NodeBeforeAspect); ok {
			beforeAspects = append(beforeAspects, a)
		}
		if a, ok := item.(NodeAfterAspect); ok {
			afterAspects = append(afterAspects, a)
		}
	}

	return beforeAspects, afterAspects
}

// GetNodeAspects 获取节点执行类型增强点切面列表
func (list AspectList) GetOnNodeBeforeInitAspects() []OnNodeBeforeInitAspect {

	//从小到大排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order() < list[j].Order()
	})

	var onNodeBeforeInitAspects []OnNodeBeforeInitAspect

	for _, item := range list {
		if a, ok := item.(OnNodeBeforeInitAspect); ok {
			onNodeBeforeInitAspects = append(onNodeBeforeInitAspects, a)
		}
	}

	return onNodeBeforeInitAspects
}

// GetNodeAspects 获取节点执行类型增强点切面列表
func (list AspectList) GetOnChainBeforeInitAspects() []OnChainBeforeInitAspect {

	//从小到大排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order() < list[j].Order()
	})

	var onChainBeforeInitAspect []OnChainBeforeInitAspect

	for _, item := range list {
		if a, ok := item.(OnChainBeforeInitAspect); ok {
			onChainBeforeInitAspect = append(onChainBeforeInitAspect, a)
		}
	}

	return onChainBeforeInitAspect
}

// GetNodeAspects 获取节点执行类型增强点切面列表
func (list AspectList) GetOnChainAggregationBeforeInitAspects() []OnChainAggregationBeforeInitAspect {

	//从小到大排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Order() < list[j].Order()
	})

	var onChainAggregationBeforeInitAspects []OnChainAggregationBeforeInitAspect

	for _, item := range list {
		if a, ok := item.(OnChainAggregationBeforeInitAspect); ok {
			onChainAggregationBeforeInitAspects = append(onChainAggregationBeforeInitAspects, a)
		}
	}

	return onChainAggregationBeforeInitAspects
}
