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
	"errors"
	"fmt"

	"github.com/bittoy/rule/types"
)

type ChainCtx struct {
	// SelfDefinition contains the complete rule chain definition including
	// metadata, nodes, connections, and configuration
	// SelfDefinition 包含完整的规则链定义，包括元数据、节点、连接和配置
	selfDefinition *types.Chain

	// config contains the rule engine configuration including
	// component registry, parser, and global settings
	// config 包含规则引擎配置，包括组件注册表、解析器和全局设置
	config types.Config

	// nodes maps node identifiers to their corresponding node contexts,
	// providing O(1) lookup time for node access operations
	// nodes 将节点标识符映射到其对应的节点上下文，为节点访问操作提供 O(1) 查找时间
	nodes map[string]types.NodeCtx

	// nodeRoutes maps each node to its outgoing relationships,
	// defining the flow of messages through the rule chain
	// nodeRoutes 将每个节点映射到其传出关系，定义消息通过规则链的流动
	nodeRoutes map[string][]types.RuleNodeRelation

	// rootRuleContext is the root context for message processing within this rule chain,
	// providing the entry point for message flow and execution coordination
	// rootRuleContext 是此规则链内消息处理的根上下文，为消息流和执行协调提供入口点
	rootNodeId string

	// aspects contains the list of AOP aspects applied to this rule chain,
	// providing cross-cutting concerns like logging, validation, and metrics
	// aspects 包含应用于此规则链的 AOP 切面列表，提供如日志、验证和指标等横切关注点
	aspects types.AspectList

	beforeAspects []types.NodeBeforeAspect
	afterAspects  []types.NodeAfterAspect

	configuration types.Configuration
}

func InitChainCtx(config types.Config, aspects types.AspectList, chainDef *types.Chain) (*ChainCtx, error) {
	// Initialize a new RuleChainCtx with the provided configuration and aspects
	// Retrieve aspects for the engine
	onChainBeforeInitAspects := aspects.GetOnChainBeforeInitAspects()
	for _, aspect := range onChainBeforeInitAspects {
		if err := aspect.OnChainBeforeInit(config, chainDef); err != nil {
			return nil, err
		}
	}
	var chainCtx = &ChainCtx{
		config:         config,
		selfDefinition: chainDef,
		nodes:          make(map[string]types.NodeCtx),
		nodeRoutes:     make(map[string][]types.RuleNodeRelation),
		aspects:        aspects,
	}

	chainCtx.beforeAspects, chainCtx.afterAspects = aspects.GetNodeAspects()

	// Load all node information
	for _, item := range chainDef.Metadata.Nodes {
		ruleNodeCtx, err := InitRuleNodeCtx(config, chainCtx, aspects, item)
		if err != nil {
			return nil, err
		}
		chainCtx.nodes[item.Id] = ruleNodeCtx

		if item.Type == types.RuleSubTypeStart {
			chainCtx.rootNodeId = item.Id
		}
	}

	// Load node relationship information
	for _, item := range chainDef.Metadata.Connections {
		inNodeId := item.FromId
		outNodeId := item.ToId
		ruleNodeRelation := types.RuleNodeRelation{
			InId:         inNodeId,
			OutId:        outNodeId,
			RelationType: item.Type,
		}
		nodeRelations, ok := chainCtx.nodeRoutes[inNodeId]

		if ok {
			nodeRelations = append(nodeRelations, ruleNodeRelation)
		} else {
			nodeRelations = []types.RuleNodeRelation{ruleNodeRelation}
		}
		chainCtx.nodeRoutes[inNodeId] = nodeRelations
	}

	return chainCtx, nil
}

// Config returns the configuration of the rule chain context
func (rc *ChainCtx) Config() types.Config {
	return rc.config
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainCtx) Id() string {
	return rc.selfDefinition.Id
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainCtx) Name() string {
	return rc.selfDefinition.Name
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainCtx) TerminalOnErr() bool {
	return rc.selfDefinition.TerminalOnErr
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainCtx) GetNodeById(id string) (types.NodeCtx, bool) {
	ruleNodeCtx, ok := rc.nodes[id]
	return ruleNodeCtx, ok
}

// GetNodeRoutes retrieves the routes for a given node ID
func (rc *ChainCtx) GetNodeRoutes(id string) ([]types.RuleNodeRelation, bool) {
	relations, ok := rc.nodeRoutes[id]
	return relations, ok
}

func (rc *ChainCtx) getNextNode(id string, relationType string) (types.NodeCtx, bool) {
	relations, ok := rc.GetNodeRoutes(id)
	if ok {
		for _, item := range relations {
			if item.RelationType == relationType {
				if nodeCtx, nodeCtxOk := rc.GetNodeById(item.OutId); nodeCtxOk {
					return nodeCtx, true
				}
			}
		}
		// 父节点存在但是子分支不存在
		return nil, true
	}
	// 父节点不存在子分支也不存在
	return nil, false
}

// Type returns the component type
func (rc *ChainCtx) Type() types.NodeType {
	return rc.selfDefinition.Type
}

// New creates a new instance (not supported for RuleChainCtx)
func (rc *ChainCtx) New() types.Node {
	panic("not support New method")
}

// Init initializes the rule chain context
func (rc *ChainCtx) Init(_ types.Config, configuration types.Configuration) error {
	return fmt.Errorf("RuleChainCtx cant not init")
}

// OnMsg processes incoming messages
func (rc *ChainCtx) OnMsg(ctx context.Context, msg types.RuleMsg) (string, error) {
	return "", rc.execute(ctx, msg)
}

// Destroy cleans up resources and executes destroy aspects
func (rc *ChainCtx) Destroy() {
	// Execute destroy aspects without holding locks
	// Note: We avoid calling methods that need locks within OnDestroy by pre-fetching data
	for _, node := range rc.nodes {
		node.Destroy()
	}
}

// DSL returns the rule chain definition as a byte slice
func (rc *ChainCtx) DSL() []byte {
	v, _ := rc.config.Parser.EncodeChain(rc.selfDefinition)
	return v
}

func (rc *ChainCtx) execute(ctx context.Context, msg types.RuleMsg) error {
	currentNode, found := rc.GetNodeById(rc.rootNodeId)
	if !found {
		return errors.New("not found rootNode")
	}
	for currentNode != nil {
		fmt.Printf("执行节点: %s (%s)\n", currentNode.Id(), currentNode.Type())

		_, err := rc.onBefore(currentNode, msg, "")
		if err != nil {
			return err
		}
		relationType, err := currentNode.OnMsg(ctx, msg)
		if err != nil {
			return err
		}
		_, err = rc.onAfter(currentNode, msg, relationType)
		if err != nil {
			return err
		}

		if len(relationType) == 0 {
			break
		}
		nodeCtx, found := rc.getNextNode(currentNode.Id(), relationType)
		if !found {
			return fmt.Errorf("node for id:%s not found", currentNode.Id())
		}
		if nodeCtx == nil {
			return fmt.Errorf("node for id:%s branch: %s node not found", currentNode.Id(), relationType)
		}
		currentNode = nodeCtx
	}
	return nil
}

// 执行After aop
func (rc *ChainCtx) onBefore(nodeCtx types.NodeCtx, msg types.RuleMsg, relationType string) (types.RuleMsg, error) {
	// after aop
	var err error
	for _, aop := range rc.beforeAspects {
		if aop.PointCut(nodeCtx, msg, relationType) {
			msg, err = aop.Before(nodeCtx, msg, relationType)
			if err != nil {
				return msg, err
			}
		}
	}
	return msg, err
}

// 执行After aop
func (rc *ChainCtx) onAfter(nodeCtx types.NodeCtx, msg types.RuleMsg, relationType string) (types.RuleMsg, error) {
	// after aop
	var err error
	for _, aop := range rc.afterAspects {
		if aop.PointCut(nodeCtx, msg, relationType) {
			msg, err = aop.After(nodeCtx, msg, relationType)
			if err != nil {
				return msg, err
			}
		}
	}
	return msg, err
}
