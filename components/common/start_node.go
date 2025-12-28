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

package common

import (
	"context"

	"github.com/bittoy/rule/types"
)

// init registers the EndNode component with the default registry.
func init() {
	Registry.Add(&StartNode{})
}

// StartNode 开始节点组件，用于触发规则链的结束回调。如果规则链设置了结束节点组件，则会替代默认的分支结束行为，只有运行到结束节点组件时，才会触发结束回调
// StartNode is an end node component that triggers the end callback of the rule chain. If the rule chain has an end node component set, it will replace the default branch ending behavior.
//
// 功能说明：
// Function Description:
// 1. 接收消息并触发DoOnEnd回调 - Receives messages and triggers DoOnEnd callback
// 2. 使用上一个节点传入的关系类型 - Uses the relation type passed from the previous node
// 3. 不会继续传递消息到下一个节点 - Does not continue passing messages to next nodes
//
// 使用场景：
// Use Cases:
// - 规则链的明确结束点 - Explicit end point of rule chains
// - 触发特定的结束处理逻辑 - Trigger specific end processing logic
// - 替代默认的分支结束行为 - Replace default branch ending behavior
type StartNode struct {
}

// Type 返回组件类型
// Type returns the component type identifier.
func (x *StartNode) Type() types.NodeType {
	return types.RuleSubTypeStart
}

// New creates a new instance.
func (x *StartNode) New() types.Node {
	return &StartNode{}
}

// Init initializes the component.
func (x *StartNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	// No configuration needed
	return nil
}

// OnMsg processes the incoming message and triggers the end callback.
func (x *StartNode) OnMsg(ctx context.Context, rCtx types.RuleContext, msg types.RuleMsg) error {
	return rCtx.TellNext(ctx, msg, types.DefaultRelationType)
}

func (x *StartNode) Destroy() {
}
