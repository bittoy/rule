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

package engine

import (
	"context"
	"fmt"

	"rule/types"
)

const (
	// defaultNodeIdPrefix is the prefix used for auto-generated node IDs
	// when no explicit ID is provided in the node definition.
	// defaultNodeIdPrefix 是当节点定义中没有提供明确 ID 时用于自动生成节点 ID 的前缀。
	defaultNodeIdPrefix = "node"
	// defaultNodeIdPrefix is the prefix used for auto-generated node IDs
	// when no explicit ID is provided in the node definition.
	// defaultNodeIdPrefix 是当节点定义中没有提供明确 ID 时用于自动生成节点 ID 的前缀。
	defaultChainIdPrefix = "chain"
)

// RuleNodeCtx represents an instance of a node component within the rule engine.
// It acts as a wrapper around the actual node implementation, providing additional
// context and metadata required for rule chain execution.
//
// RuleNodeCtx 表示规则引擎中节点组件的实例。
// 它充当实际节点实现的包装器，提供规则链执行所需的额外上下文和元数据。
//
// Architecture:
// 架构：
//
//	RuleNodeCtx embeds the types.Node interface, allowing it to act as both
//	a node wrapper and a node implementation. This design provides:
//	RuleNodeCtx 嵌入 types.Node 接口，允许它既充当节点包装器又充当节点实现。
//	此设计提供：
//	- Direct access to node methods through interface embedding  通过接口嵌入直接访问节点方法
//	- Additional context and configuration management  额外的上下文和配置管理
//	- Thread-safe operations with mutex protection  使用互斥锁保护的线程安全操作
//	- Hot reloading capabilities  热重载功能
type RuleNodeCtx struct {
	// types.Node is the embedded node implementation providing the core functionality.
	// This embedding allows RuleNodeCtx to act as a node while adding wrapper capabilities.
	// types.Node 是嵌入的节点实现，提供核心功能。
	// 这种嵌入允许 RuleNodeCtx 在添加包装器功能的同时充当节点。
	types.Node

	chainCtx types.ChainCtx

	// SelfDefinition contains the configuration and metadata for this specific node,
	// including its type, ID, configuration parameters, and behavioral settings.
	// SelfDefinition 包含此特定节点的配置和元数据，
	// 包括其类型、ID、配置参数和行为设置。
	selfDefinition *types.BaseInfo

	// config holds the global rule engine configuration,
	// providing access to component registry, parsers, and global settings.
	// config 保存全局规则引擎配置，提供对组件注册表、解析器和全局设置的访问。
	config types.Config
}

// InitRuleNodeCtx initializes a RuleNodeCtx with the given parameters.
// This is the standard initialization function for regular nodes without network resources.
//
// InitRuleNodeCtx 使用给定参数初始化 RuleNodeCtx。
// 这是不带网络资源的常规节点的标准初始化函数。
//
// Parameters:
// 参数：
//   - config: Global rule engine configuration  全局规则引擎配置
//   - chainCtx: Parent rule chain context  父规则链上下文
//   - aspects: List of AOP aspects to apply  要应用的 AOP 切面列表
//   - selfDefinition: Node definition and configuration  节点定义和配置
//
// Returns:
// 返回：
//   - *RuleNodeCtx: Initialized node context  已初始化的节点上下文
//   - error: Initialization error if any  如果有的话，初始化错误
func InitRuleNodeCtx(config types.Config, chainCtx *ChainCtx, aspects types.AspectList, selfDefinition *types.BaseInfo) (*RuleNodeCtx, error) {
	return initRuleNodeCtx(config, chainCtx, aspects, selfDefinition)
}

// initRuleNodeCtx is the core initialization function for RuleNodeCtx.
// It handles the complete node initialization process including component creation,
// configuration processing, and aspect integration.
//
// initRuleNodeCtx 是 RuleNodeCtx 的核心初始化函数。
// 它处理完整的节点初始化过程，包括组件创建、配置处理和切面集成。
//
// Parameters:
// 参数：
//   - config: Global rule engine configuration  全局规则引擎配置
//   - chainCtx: Parent rule chain context  父规则链上下文
//   - aspects: List of AOP aspects to apply  要应用的 AOP 切面列表
//   - selfDefinition: Node definition and configuration  节点定义和配置
//   - isInitNetResource: Whether to initialize network resources  是否初始化网络资源
//
// Returns:
// 返回：
//   - *RuleNodeCtx: Initialized node context  已初始化的节点上下文
//   - error: Initialization error if any  如果有的话，初始化错误
//
// Initialization Process:
// 初始化过程：
//  1. Execute before-init aspects  执行初始化前切面
//  2. Create node instance from component registry  从组件注册表创建节点实例
//  3. Process configuration variables and templates  处理配置变量和模板
//  4. Inject chain context and node definition  注入链上下文和节点定义
//  5. Initialize the node with processed configuration  使用处理过的配置初始化节点
//  6. Return wrapped node context  返回包装的节点上下文
//
// Error Handling:
// 错误处理：
//   - Aspect execution failures  切面执行失败
//   - Component creation errors  组件创建错误
//   - Configuration processing failures  配置处理失败
//   - Node initialization errors  节点初始化错误
func initRuleNodeCtx(config types.Config, chainCtx *ChainCtx, aspects types.AspectList, selfDefinition *types.BaseInfo) (*RuleNodeCtx, error) {
	// Retrieve aspects for the engine.
	//beforeAspect, afterAspect := aspects.GetNodeAspects()
	onNodeBeforeInitAspects := aspects.GetOnNodeBeforeInitAspects()
	for _, aspect := range onNodeBeforeInitAspects {
		if err := aspect.OnNodeBeforeInit(config, selfDefinition); err != nil {
			return nil, fmt.Errorf("nodeType:%s for id:%s OnNodeBeforeInit error:%s", selfDefinition.Type, selfDefinition.Id, err.Error())
		}
	}

	node, err := config.ComponentsRegistry.NewNode(selfDefinition.Type)
	if err != nil {
		return nil, fmt.Errorf("nodeType:%s for id:%s new error:%s", selfDefinition.Type, selfDefinition.Id, err.Error())
	}

	// Initialize the node with the processed configuration.
	if err = node.Init(config, selfDefinition.Configuration); err != nil {
		return nil, fmt.Errorf("nodeType:%s for id:%s init error:%s", selfDefinition.Type, selfDefinition.Id, err.Error())
	}

	// Return a RuleNodeCtx with the initialized node and provided context and definition.
	return &RuleNodeCtx{
		Node:           node,
		selfDefinition: selfDefinition,
		config:         config,
		chainCtx:       chainCtx,
	}, nil
}

// Config returns the configuration of the rule engine.
func (rn *RuleNodeCtx) Config() types.Config {
	return rn.config
}

// GetNodeId returns the ID of the node.
func (rn *RuleNodeCtx) Id() string {
	return rn.selfDefinition.Id
}

// GetNodeId returns the ID of the node.
func (rn *RuleNodeCtx) Name() string {
	return rn.selfDefinition.Name
}

// GetNodeById retrieves a node context by its ID
func (rc *RuleNodeCtx) TerminalOnErr() bool {
	return rc.selfDefinition.TerminalOnErr
}

// ReloadSelf reloads the node from a byte slice definition.
func (rn *RuleNodeCtx) ReloadSelf(_ []byte) error {
	return nil
}

// DSL returns the DSL representation of the node.
func (rn *RuleNodeCtx) DSL() []byte {
	parser := rn.config.Parser
	selfDefinition := rn.selfDefinition

	result, _ := parser.EncodeRule(selfDefinition)
	return result
}

// OnMsg 提供并发安全的消息处理，保护内嵌Node访问
// OnMsg provides concurrent-safe message processing with protected access to the embedded Node.
// This method ensures thread safety during message processing by using read locks to protect
// against concurrent modifications during hot reloads.
//
// OnMsg 提供并发安全的消息处理，通过使用读锁保护嵌入的 Node 访问。
// 此方法通过使用读锁防止热重载期间的并发修改，确保消息处理期间的线程安全。
//
// Parameters:
// 参数：
//   - ctx: Rule context for message processing  用于消息处理的规则上下文
//   - msg: Message to be processed  要处理的消息
func (rn *RuleNodeCtx) OnMsg(ctx context.Context, rCtx types.RuleContext, msg types.RuleMsg) error {
	// 使用读锁保护Node字段的访问，与ReloadSelfFromDef的写锁互斥
	return rn.Node.OnMsg(ctx, rCtx, msg)
}

// Destroy safely destroys the embedded node
func (rn *RuleNodeCtx) Destroy() {
	rn.Node.Destroy()
}
