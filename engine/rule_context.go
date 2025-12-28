/*
 * Copyright 2025 The RuleGo Authors.
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

// DefaultRuleContext is the default context for message processing in the rule engine.
type DefaultRuleContext struct {
	// Context of the root rule chain.
	ruleChainCtx *ChainCtx
	// Context of the previous node.
	from types.NodeCtx
	// Context of the current node.
	self types.NodeCtx
	// first node relationType
	relationType string
}

// NewRuleContext creates a new instance of the default rule engine message processing context.
func NewRuleContext(ruleChainCtx *ChainCtx, from types.NodeCtx, self types.NodeCtx) *DefaultRuleContext {
	// Return a new DefaultRuleContext populated with the provided parameters and aspects.
	return &DefaultRuleContext{
		ruleChainCtx: ruleChainCtx,
		from:         from,
		self:         self,
	}
}

// NewNextNodeRuleContext 创建下一个节点的规则上下文
// NewNextNodeRuleContext creates a rule context for the next node
func (ctx *DefaultRuleContext) NewNextNodeRuleContext(nextNode types.NodeCtx) *DefaultRuleContext {
	// Create a new context directly instead of using object pool to avoid data races
	// 但是复用不可变的共享状态以减少内存开销
	nextCtx := &DefaultRuleContext{
		ruleChainCtx: ctx.ruleChainCtx, // 共享规则链上下文
		from:         ctx.self,
		self:         nextNode,
	}

	return nextCtx
}

func (rCtx *DefaultRuleContext) Tell(ctx context.Context, msg types.RuleMsg, relationType string) error {
	return rCtx.tell(ctx, msg, rCtx.self, relationType)
}

func (rCtx *DefaultRuleContext) TellNext(ctx context.Context, msg types.RuleMsg, relationType string) error {
	return rCtx.tellNext(ctx, msg, relationType)
}

func (ctx *DefaultRuleContext) Self() types.NodeCtx {
	return ctx.self
}

func (ctx *DefaultRuleContext) From() types.NodeCtx {
	return ctx.from
}

// getNextNodes 获取当前节点指定关系的子节点
// 如果找到任何子节点则为 true
func (ctx *DefaultRuleContext) getNextNode(relationType string) (types.NodeCtx, bool) {
	return ctx.ruleChainCtx.GetNextNode(ctx.self.Id(), relationType)
}

// tellNext 通知执行子节点，如果是当前第一个节点则执行当前节点
// 如果找不到relationTypes对应的节点，而且defaultRelationType非默认值，则通过defaultRelationType查找节点
func (rCtx *DefaultRuleContext) tellNext(ctx context.Context, msg types.RuleMsg, relationType string) error {
	//执行After aop
	msg, err := rCtx.onAfter(msg, relationType)
	if err != nil {
		return err
	}
	//根据relationType查找子节点列表
	node, ok := rCtx.getNextNode(relationType)
	if !ok {
		return fmt.Errorf("no next node for id:%s with relationType:%s", rCtx.self.Id(), relationType)
	}
	if node != nil {
		return rCtx.tell(ctx, msg, node, relationType)
	}
	return nil
}

// 执行After aop
func (rCtx *DefaultRuleContext) onBefore(msg types.RuleMsg, relationType string) (types.RuleMsg, error) {
	// after aop
	var err error
	for _, aop := range rCtx.ruleChainCtx.beforeAspects {
		if aop.PointCut(rCtx, msg, relationType) {
			msg, err = aop.Before(rCtx, msg, relationType)
			if err != nil {
				return msg, err
			}
		}
	}
	return msg, err
}

// 执行After aop
func (rCtx *DefaultRuleContext) onAfter(msg types.RuleMsg, relationType string) (types.RuleMsg, error) {
	// after aop
	var err error
	for _, aop := range rCtx.ruleChainCtx.afterAspects {
		if aop.PointCut(rCtx, msg, relationType) {
			msg, err = aop.After(rCtx, msg, relationType)
			if err != nil {
				return msg, err
			}
		}
	}
	return msg, err
}

// 执行下一个节点
func (rCtx *DefaultRuleContext) tell(ctx context.Context, msg types.RuleMsg, nextNode types.NodeCtx, relationType string) error {
	nextCtx := rCtx.NewNextNodeRuleContext(nextNode)

	//aop
	msg, err := nextCtx.onBefore(msg, relationType)
	if err != nil {
		return err
	}

	// 已经执行节点OnMsg逻辑，不在执行下面的逻辑
	nextCtx.relationType = relationType
	return nextNode.OnMsg(ctx, nextCtx, msg)
}
