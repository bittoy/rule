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
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/bittoy/rule/types"
)

// Ensuring RuleEngine implements types.RuleEngine interface.
var _ types.Engine = (*ChainEngine)(nil)

type ChainEngine struct {
	// Config is the configuration for the rule engine containing
	// global settings, component registry, and execution parameters
	// Config 是规则引擎的配置，包含全局设置、组件注册表和执行参数
	config types.Config

	// rootRuleChainCtx is the context of the root rule chain containing
	// all nodes and their relationships
	// rootRuleChainCtx 是根规则链的上下文，包含所有节点及其关系
	ruleChainCtx *ChainCtx

	// initialized indicates whether the rule engine has been properly initialized
	// Use atomic operations to prevent data races during concurrent access
	// initialized 指示规则引擎是否已正确初始化
	// 使用原子操作防止并发访问时的数据竞态
	initialized int32

	// Aspects is a list of AOP (Aspect-Oriented Programming) aspects
	// that provide cross-cutting concerns like logging, validation, and metrics
	// Aspects 是面向切面编程（AOP）切面列表，提供如日志、验证和指标等横切关注点
	aspects types.AspectList

	beforeAspects []types.ChainBeforeAspect

	afterAspects []types.ChainAfterAspect

	// Callbacks provides hooks for rule engine lifecycle events,
	// enabling custom handling of creation, updates, and deletion.
	// Callbacks 为规则引擎生命周期事件提供钩子，
	// 支持创建、更新和删除的自定义处理。
	callbacks types.Callbacks
}

func NewChainEngine(def []byte, opts ...types.EngineOption) (types.Engine, error) {
	if len(def) == 0 {
		return nil, errors.New("def can not nil")
	}

	// 使用 ID 创建新的 RuleEngine
	ruleEngine := &ChainEngine{
		config: NewConfig(),
	}
	ruleEngine.callbacks = types.NewCallbacks(
		types.WithOnNew(ruleEngine.onNew),
		types.WithOnDeleted(ruleEngine.onDelete),
		types.WithOnUpdated(ruleEngine.onUpdate),
	)

	err := ruleEngine.reloadSelf(def, opts...)
	return ruleEngine, err
}

// Id returns the unique identifier of the rule engine instance.
// Id 返回规则引擎实例的唯一标识符。
func (e *ChainEngine) Id() string {
	return e.ruleChainCtx.Id()
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainEngine) Name() string {
	return rc.ruleChainCtx.Name()
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainEngine) TerminalOnErr() bool {
	return rc.ruleChainCtx.TerminalOnErr()
}

// SetConfig updates the configuration of the rule engine.
// This should be called before initialization for best results.
// SetConfig 更新规则引擎的配置。
// 为了获得最佳效果，应在初始化前调用。
func (e *ChainEngine) SetConfig(config types.Config) {
	e.config = config
}

// SetAspects updates the list of aspects used by the rule engine.
// Aspects provide cross-cutting functionality like logging and validation.
// SetAspects 更新规则引擎使用的切面列表。
// 切面提供如日志和验证等横切功能。
func (e *ChainEngine) SetAspects(aspects ...types.Aspect) {
	e.aspects = aspects
}

// GetAspects returns a copy of the current aspects list to avoid data races.
// GetAspects 返回当前切面列表的副本以避免数据竞争。
func (e *ChainEngine) GetAspects() types.AspectList {
	return e.aspects
}

// initBuiltinsAspects initializes the built-in aspects if no custom aspects are provided.
// It ensures that essential aspects like validation and debugging are always available.
// initBuiltinsAspects 如果没有提供自定义切面，则初始化内置切面。
// 它确保验证和调试等基本切面始终可用。
func (e *ChainEngine) initBuiltinsAspects() {
	//初始化内置切面
	for _, builtinsAspect := range BuiltinsAspects {
		e.aspects = append(e.aspects, builtinsAspect.New())
	}
	e.beforeAspects, e.afterAspects = e.aspects.GetChainAspects()
}

// initChain initializes the rule chain with the provided definition.
// It sets up all nodes, relationships, and executes creation aspects.
// initChain 使用提供的定义初始化规则链。
// 它设置所有节点、关系并执行创建切面。
func (e *ChainEngine) init(def types.Chain) error {
	if def.Disabled {
		return types.ErrEngineDisabled
	}
	ctx, err := InitChainCtx(e.config, e.aspects, &def)
	if err != nil {
		return err
	}

	unsafepL := (*unsafe.Pointer)(unsafe.Pointer(&e.ruleChainCtx))
	atomic.StorePointer(unsafepL, unsafe.Pointer(ctx))

	return nil
}

// ReloadSelf reloads the rule chain with new definition and options.
// This method supports hot reloading of rule configurations without stopping the engine.
// It implements a two-phase graceful reload process:
//
// Phase 1: Preparation (设置阶段)
// - Apply configuration options  应用配置选项
// - Wait for any ongoing reload to complete  等待任何正在进行的重载完成
// - Set reloading state to block new messages  设置重载状态以阻塞新消息
// - Wait for active messages to complete  等待活跃消息完成
//
// Phase 2: Reload (重载阶段)
// - Parse new rule chain definition  解析新的规则链定义
// - Update or create rule chain context  更新或创建规则链上下文
// - Update atomic aspect pointers  更新原子切面指针
// - Resume normal operation  恢复正常运行
//
// ReloadSelf 使用新定义和选项重新加载规则链。
// 此方法支持在不停止引擎的情况下热重载规则配置。
// 它实现了两阶段优雅重载过程：
//
// Parameters:
// 参数：
//   - dsl: Rule chain definition in byte format  字节格式的规则链定义
//   - opts: Optional configuration functions  可选的配置函数
//
// Returns:
// 返回：
//   - error: Reload error if any  如果有的话，重载错误
func (e *ChainEngine) ReloadSelf(dsl []byte) error {
	return e.reloadSelf(dsl)
}

func (e *ChainEngine) reloadSelf(dsl []byte, opts ...types.EngineOption) error {
	// Apply the options to the RuleEngine.
	// 将选项应用于 RuleEngine。
	for _, opt := range opts {
		_ = opt(e)
	}

	//初始化
	chainDef, err := e.config.Parser.DecodeChain(dsl)
	if err != nil {
		return err
	}

	err = e.init(chainDef)
	if err != nil {
		return err
	}

	if e.isInitialized() {
		//执行创建切面逻辑
		if e.callbacks.OnUpdated != nil {
			e.callbacks.OnUpdated(e.Id(), e.DSL())
		}
	} else {
		e.initBuiltinsAspects()
		e.setInitialized()
		//执行创建切面逻辑
		if e.callbacks.OnNew != nil {
			e.callbacks.OnNew(e.Id(), e.DSL())
		}
	}

	return nil
}

// DSL returns the current rule chain configuration in its original format.
// DSL 返回原始格式的当前规则链配置。
func (e *ChainEngine) DSL() []byte {
	return e.ruleChainCtx.DSL()
}

// Initialized returns whether the rule engine has been properly initialized.
// Initialized 返回规则引擎是否已正确初始化。
func (e *ChainEngine) isInitialized() bool {
	return atomic.LoadInt32(&e.initialized) == 1
}

// Initialized returns whether the rule engine has been properly initialized.
// Initialized 返回规则引擎是否已正确初始化。
func (e *ChainEngine) setInitialized() {
	atomic.StoreInt32(&e.initialized, 1)
}

// Initialized returns whether the rule engine has been properly initialized.
// Initialized 返回规则引擎是否已正确初始化。
func (e *ChainEngine) unSetInitialized() {
	atomic.StoreInt32(&e.initialized, 0)
}

// Stop 关闭规则引擎并释放所有资源。
// 实现两阶段优雅停机策略：
func (e *ChainEngine) Stop() {
	// Clean up resources
	// 清理资源
	e.forceStop()
	//e.callbacks.OnDeleted()
}

// forceStop performs immediate cleanup of all rule engine resources.
// This method is called during shutdown to ensure complete resource cleanup,
// regardless of whether graceful shutdown completed successfully.
//
// Cleanup operations (with panic recovery):
// 清理操作（带恐慌恢复）：
// 1. Force cancellation of graceful shutdown context  强制取消优雅停机上下文
// 2. Destroy rule chain context and all nodes  销毁规则链上下文和所有节点
// 3. Clear instance cache entries  清理实例缓存条目
// 4. Reset initialization state  重置初始化状态
//
// Each cleanup operation is wrapped with panic recovery to ensure
// that failures in one cleanup step don't prevent others from executing.
//
// forceStop 执行规则引擎资源的立即清理。
// 此方法在停机期间调用以确保完整的资源清理，无论优雅停机是否成功完成。
func (e *ChainEngine) forceStop() {
	if e.callbacks.OnDeleted != nil {
		e.callbacks.OnDeleted(e.Id())
	}

	// Destroy rule chain context and all nodes
	// 销毁规则链上下文和所有节点
	if e.ruleChainCtx != nil {
		e.ruleChainCtx.Destroy()
		e.ruleChainCtx = nil
	}

	e.unSetInitialized()
}

// OnMsg asynchronously processes a message using the rule engine.
// It accepts optional RuleContextOption parameters to customize the execution context.
//
// OnMsg 使用规则引擎异步处理消息。
// 它接受可选的 RuleContextOption 参数来自定义执行上下文。
func (e *ChainEngine) OnMsg(ctx context.Context, msg types.RuleMsg) error {
	return e.onMsg(ctx, msg)
}

func (e *ChainEngine) onMsg(ctx context.Context, msg types.RuleMsg) error {
	var err error
	start := time.Now()
	defer func() {
		var status int
		if err != nil {
			status = 100
		}
		duration := time.Since(start).Seconds()
		// 统计
		enginRequestsTotal.WithLabelValues(
			e.Name(),
			strconv.Itoa(status),
		).Inc()

		enginRequestDuration.WithLabelValues(
			e.Name(),
		).Observe(duration)
	}()

	// Execute start aspects
	// 执行开始切面
	msg, err = e.onBefore(msg)
	if err != nil {
		return err
	}

	// Process message with or without waiting
	// 处理消息，可选择是否等待
	if err := e.ruleChainCtx.OnMsg(ctx, NewChainContext(e.ruleChainCtx), msg); err != nil {
		return err
	}

	// Execute start aspects
	// 执行开始切面
	_, err = e.onAfter(msg)
	return err
}

func (e *ChainEngine) onBefore(msg types.RuleMsg) (types.RuleMsg, error) {
	var err error
	for _, aop := range e.beforeAspects {
		if aop.PointCut(NewChainContext(e.ruleChainCtx), msg) {
			msg, err = aop.Before(NewChainContext(e.ruleChainCtx), msg)
		}
	}
	return msg, err
}

// onEnd executes the list of end aspects when a branch of the rule chain ends.
// onEnd 在规则链分支结束时执行结束切面列表。
func (e *ChainEngine) onAfter(msg types.RuleMsg) (types.RuleMsg, error) {
	var err error
	for _, aop := range e.afterAspects {
		if aop.PointCut(NewChainContext(e.ruleChainCtx), msg) {
			msg, err = aop.After(NewChainContext(e.ruleChainCtx), msg)
		}
	}
	return msg, err
}

func (e *ChainEngine) onNew(chainId string, dsl []byte) {
	e.config.Logger.Printf("ChainEngine OnNew: chainId=%s", chainId)
}

func (e *ChainEngine) onUpdate(chainId string, dsl []byte) {
	e.config.Logger.Printf("ChainEngine onUpdate: chainId=%s", chainId)
}

func (e *ChainEngine) onDelete(id string) {
	e.config.Logger.Printf("ChainEngine onDel: chainId=%s", id)
}
