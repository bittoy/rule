package engine

///*
// * Copyright 2025 The RuleGo Authors.
// *
// * Licensed under the Apache License, Version 2.0 (the "License");
// * you may not use this file except in compliance with the License.
// * You may obtain a copy of the License at
// *
// *     http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing, software
// * distributed under the License is distributed on an "AS IS" BASIS,
// * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// * See the License for the specific language governing permissions and
// * limitations under the License.
// */
//
//package engine
//
//import (
//	"context"
//
//	"github.com/bittoy/rule/types"
//)
//
//// DefaultRuleContext is the default context for message processing in the rule engine.
//type DefaultChainAggregationContext struct {
//	// Context of the current node.
//	self types.ChainAggregationCtx
//}
//
//// NewRuleContext creates a new instance of the default rule engine message processing context.
//func NewDefaultChainAggregationContext(self types.ChainAggregationCtx) *DefaultChainAggregationContext {
//	// Return a new DefaultRuleContext populated with the provided parameters and aspects.
//	return &DefaultChainAggregationContext{
//		self: self,
//	}
//}
//
//func (rCtx *DefaultChainAggregationContext) Tell(ctx context.Context, msg types.RuleMsg, relationType string) error {
//	return nil
//}
//
//func (rCtx *DefaultChainAggregationContext) TellNext(ctx context.Context, msg types.RuleMsg, relationType string) error {
//	return nil
//}
//
//func (ctx *DefaultChainAggregationContext) Self() types.NodeCtx {
//	return ctx.self
//}
//
//func (ctx *DefaultChainAggregationContext) From() types.NodeCtx {
//	return nil
//}
