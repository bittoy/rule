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

package aspect

import (
	"fmt"
	"sync"

	"github.com/bittoy/rule/types"
)

var (
	// Compile-time check Validator implements types.OnChainBeforeInitAspect.
	_ types.OnChainBeforeInitAspect = (*ChainValidator)(nil)
)

// Validator is a rule chain initialization validation aspect that performs
// comprehensive validation checks before rule chain creation. It ensures
// rule chain integrity and prevents invalid configurations from being deployed.
//
// Validator 是规则链初始化验证切面，在规则链创建之前执行全面的验证检查。
// 它确保规则链完整性并防止部署无效配置。
//
// Features:
// 功能特性：
//   - Pre-initialization validation  初始化前验证
//   - Cycle detection in rule chains  规则链中的环检测
//   - Endpoint node restrictions for sub-chains  子链的端点节点限制
//   - Extensible validation rule system  可扩展的验证规则系统
//   - Configurable validation behavior  可配置的验证行为
//
// Built-in Validation Rules:
// 内置验证规则：
//   - Sub-chains cannot contain endpoint nodes  子链不能包含端点节点
//   - Cycle detection (unless explicitly allowed)  环检测（除非明确允许）
//   - Node existence validation  节点存在性验证
//   - Connection integrity checks  连接完整性检查
//
// Usage:
// 使用方法：
//
//	// Apply validator to rule engine
//	// 为规则引擎应用验证器
//	config := types.NewConfig().WithAspects(&Validator{})
//	engine := rulego.NewRuleEngine(config)
//
//	// Add custom validation rules
//	// 添加自定义验证规则
//	Rules.AddRule(func(config types.Config, def *types.RuleChain) error {
//		// Custom validation logic
//		return nil
//	})
type ChainValidator struct {
}

// Order returns the execution order of this aspect. Lower values execute earlier.
// Validator has order 10, ensuring validation occurs before other aspects.
//
// Order 返回此切面的执行顺序。值越低，执行越早。
// Validator 的顺序为 10，确保验证在其他切面之前进行。
func (aspect *ChainValidator) Order() int {
	return 10
}

// New creates a new instance of the validation aspect.
// Each rule engine gets its own validator instance.
//
// New 创建验证切面的新实例。
// 每个规则引擎都获得自己的验证器实例。
func (aspect *ChainValidator) New() types.Aspect {
	return &ChainValidator{}
}

// Type returns the unique identifier for this aspect type.
//
// Type 返回此切面类型的唯一标识符。
func (aspect *ChainValidator) Type() string {
	return "chainValidator"
}

func (aspect *ChainValidator) PointCut(ctx types.RuleContext, msg types.RuleMsg, relationType string) bool {
	return true
}

// OnChainBeforeInit is called before rule chain initialization. It executes
// all registered validation rules and returns an error if any validation fails.
// This prevents invalid rule chains from being created.
//
// OnChainBeforeInit 在规则链初始化之前调用。它执行所有注册的验证规则，
// 如果任何验证失败则返回错误。这防止创建无效的规则链。
//
// Parameters:
// 参数：
//   - config: Rule engine configuration  规则引擎配置
//   - def: Rule chain definition to validate  要验证的规则链定义
//
// Returns:
// 返回：
//   - error: Validation error if any rule fails, nil if all pass
//     error：如果任何规则失败则返回验证错误，如果全部通过则为 nil
func (aspect *ChainValidator) OnChainBeforeInit(config types.Config, def *types.Chain) error {
	ruleList := ChainRules.Rules()
	for _, rule := range ruleList {
		if err := rule(config, def); err != nil {
			return err
		}
	}
	return nil
}

// Rules is the global validation rules registry that manages all validation
// functions applied during rule chain initialization.
//
// Rules 是全局验证规则注册表，管理规则链初始化期间应用的所有验证函数。
var ChainRules = NewChainRules()

// rules is a thread-safe container for validation rule functions.
// It provides methods to add new rules and retrieve existing ones safely.
//
// rules 是验证规则函数的线程安全容器。
// 它提供安全地添加新规则和检索现有规则的方法。
type chainRules struct {
	rules        []func(config types.Config, def *types.Chain) error // Validation rule functions  验证规则函数
	sync.RWMutex                                                     // Reader-writer mutex for thread safety  用于线程安全的读写互斥锁
}

// NewRules creates a new rules registry with default validation rules pre-configured.
// It includes built-in rules for endpoint node restrictions and cycle detection.
//
// NewRules 创建一个预配置默认验证规则的新规则注册表。
// 它包括端点节点限制和环检测的内置规则。
//
// Default Rules:
// 默认规则：
//  1. Sub-chains cannot contain endpoint nodes  子链不能包含端点节点
//  2. Cycle detection (when not explicitly allowed)  环检测（当未明确允许时）
//
// Returns:
// 返回：
//   - *rules: Configured rules registry  配置好的规则注册表
func NewChainRules() *chainRules {
	r := &chainRules{}
	//建环检测
	r.AddRule(func(config types.Config, def *types.Chain) error {
		if def != nil {
			return checkChainCycles(def)
		}
		return nil
	})
	r.AddRule(func(config types.Config, def *types.Chain) error {
		if def != nil {
			return validateChainNode(def)
		}
		return nil
	})
	return r
}

// AddRule adds one or more validation rule functions to the registry.
// New rules are appended to the existing list and will be executed
// in the order they were added.
//
// AddRule 向注册表添加一个或多个验证规则函数。
// 新规则会附加到现有列表中，并按添加顺序执行。
//
// Parameters:
// 参数：
//   - fn: Variable number of validation rule functions
//     fn：可变数量的验证规则函数
//
// Thread Safety:
// 线程安全：
// This method is thread-safe and uses a write lock to ensure
// concurrent modifications don't corrupt the rules list.
// 此方法是线程安全的，使用写锁确保并发修改不会破坏规则列表。
func (r *chainRules) AddRule(fn ...func(config types.Config, def *types.Chain) error) {
	r.Lock()
	defer r.Unlock()
	r.rules = append(r.rules, fn...)
}

// Rules returns a copy of all validation rule functions.
// This method provides thread-safe access to the rules without exposing
// the internal slice to modification.
//
// Rules 返回所有验证规则函数的副本。
// 此方法提供对规则的线程安全访问，而不会将内部切片暴露给修改。
//
// Returns:
// 返回：
//   - []func(...) error: Copy of validation rule functions  验证规则函数的副本
//
// Thread Safety:
// 线程安全：
// This method uses a read lock to allow concurrent reads while
// preventing reads during rule modifications.
// 此方法使用读锁允许并发读取，同时防止在规则修改期间读取。
func (r *chainRules) Rules() []func(config types.Config, def *types.Chain) error {
	r.RLock()
	defer r.RUnlock()
	return append([]func(config types.Config, def *types.Chain) error(nil), r.rules...)
}

// CheckCycles performs cycle detection in rule chains using topological sorting algorithm.
// It builds a directed graph from rule node connections and detects cycles that would
// cause infinite loops during rule execution.
//
// CheckCycles 使用拓扑排序算法在规则链中执行环检测。
// 它从规则节点连接构建有向图，并检测在规则执行期间会导致无限循环的环。
//
// Algorithm:
// 算法：
//  1. Build adjacency list and in-degree table  构建邻接表和入度表
//  2. Initialize queue with zero in-degree nodes  用零入度节点初始化队列
//  3. Process nodes in topological order  按拓扑顺序处理节点
//  4. If not all nodes processed, cycle exists  如果未处理所有节点，则存在环
//
// Parameters:
// 参数：
//   - metadata: Rule chain metadata containing nodes and connections
//     metadata：包含节点和连接的规则链元数据
//
// Returns:
// 返回：
//   - error: ErrCycleDetected if cycle found, nil if no cycles
//     error：如果发现环则返回 ErrCycleDetected，如果没有环则为 nil
//
// Time Complexity: O(V + E) where V is nodes and E is connections
// 时间复杂度：O(V + E)，其中 V 是节点数，E 是连接数
//
// Space Complexity: O(V + E) for adjacency list and degree tracking
// 空间复杂度：O(V + E) 用于邻接表和度数跟踪

func checkChainCycles(ruleChain *types.Chain) error {
	hasCycle, path := checkCycles(ruleChain.Metadata.Connections)
	if hasCycle {
		return fmt.Errorf("cycle detected in rule chain ruleId:%s path:%v", ruleChain.Id, path)
	}
	return nil
}

func checkCycles(edges []types.NodeConnection) (bool, []string) {
	// 创建邻接表和入度表{
	graph := map[string][]string{}
	for _, e := range edges {
		graph[e.FromId] = append(graph[e.FromId], e.ToId)
	}

	visited := map[string]bool{}
	stack := map[string]bool{}
	var path []string

	var dfs func(node string) bool
	dfs = func(node string) bool {
		if stack[node] {
			path = append(path, node)
			return true
		}
		if visited[node] {
			return false
		}

		visited[node] = true
		stack[node] = true
		path = append(path, node)

		for _, next := range graph[node] {
			if dfs(next) {
				return true
			}
		}

		stack[node] = false
		path = path[:len(path)-1]
		return false
	}

	for node := range graph {
		if !visited[node] {
			if dfs(node) {
				return true, path
			}
		}
	}
	return false, nil
}

func validateChainNode(chain *types.Chain) error {
	var nodeRoutes = make(map[string][]types.RuleNodeRelation)
	var nodes = make(map[string]struct{})
	var hasStart bool
	var hasEnd bool
	if len(chain.Metadata.Nodes) == 0 || len(chain.Metadata.Connections) == 0 {
		return fmt.Errorf("规则链中必须包含规则节点和消息拓扑节点")
	}

	for _, node := range chain.Metadata.Nodes {
		if node.Type == types.RuleSubTypeStart {
			hasStart = true
		}
		if node.Type == types.RuleSubTypeEnd {
			hasEnd = true
		}
		nodes[node.Id] = struct{}{}
	}
	if !hasStart {
		return fmt.Errorf("%s 规则链中必须包含一个开始节点", chain.Id)
	}
	if !hasEnd {
		return fmt.Errorf("%s 规则链中必须包含一个结束节点", chain.Id)
	}

	for _, item := range chain.Metadata.Connections {
		inNodeId := item.FromId
		outNodeId := item.ToId

		if _, ok := nodes[inNodeId]; !ok {
			return fmt.Errorf("节点 %s 不存在", inNodeId)
		}
		if _, ok := nodes[outNodeId]; !ok {
			return fmt.Errorf("节点 %s 不存在", outNodeId)
		}

		ruleNodeRelation := types.RuleNodeRelation{
			InId:         inNodeId,
			OutId:        outNodeId,
			RelationType: item.Type,
		}
		nodeRelations, ok := nodeRoutes[inNodeId]
		if ok {
			nodeRelations = append(nodeRelations, ruleNodeRelation)
		} else {
			nodeRelations = []types.RuleNodeRelation{ruleNodeRelation}
		}
		nodeRoutes[inNodeId] = nodeRelations
	}

	for _, node := range chain.Metadata.Nodes {
		if node.Type == types.RuleSubTypeStart || node.Type == types.RuleSubTypeExprAssign {
			if len(nodeRoutes[node.Id]) != 1 || nodeRoutes[node.Id][0].RelationType != types.DefaultRelationType {
				return fmt.Errorf("节点 %s(%s) 必须有且仅有一个 default 连接，但当前有 %d 个连接", node.Id, node.Type, len(nodeRoutes[node.Id]))
			}
		}
		if node.Type == types.RuleSubTypeEnd {
			if len(nodeRoutes[node.Id]) != 0 {
				return fmt.Errorf("节点 %s(%s) 不能有连接，但当前有 %d 个连接", node.Id, node.Type, len(nodeRoutes[node.Id]))
			}
		}
		if node.Type == types.RuleSubTypeJsFilter || node.Type == types.RuleSubTypeExprFilter {
			if len(nodeRoutes[node.Id]) != 2 {
				return fmt.Errorf("节点 %s(%s) 必须有两个连接", node.Id, node.Type)
			}
			var hasTrue bool
			var hasFalse bool
			for _, relation := range nodeRoutes[node.Id] {
				if relation.RelationType == types.TrueRelationType {
					hasTrue = true
				}
				if relation.RelationType == types.FalseRelationType {
					hasFalse = true
				}
			}
			if !hasFalse || !hasTrue {
				return fmt.Errorf("节点 %s(%s) 必须有ture和false两个连接", node.Id, node.Type)
			}
		}
		if node.Type == types.RuleSubTypeExprSwitch || node.Type == types.RuleSubTypeJsSwitch {
			if len(nodeRoutes[node.Id]) == 0 {
				return fmt.Errorf("节点 %s(%s) 必须有且仅有一个 default 连接，但当前没有任何连接", node.Id, node.Type)
			}
			var haveDefault bool
			for _, ruleNodeRelation := range nodeRoutes[node.Id] {
				if ruleNodeRelation.RelationType == types.DefaultRelationType {
					haveDefault = true
				}
			}
			if !haveDefault {
				return fmt.Errorf("节点 %s(%s) 必须有且仅有一个 default 连接，但当前没有任何 default 连接", node.Id, node.Type)
			}
		}
	}
	return nil
}

func endNodes(edges []types.NodeConnection) []string {
	fromSet := make(map[string]bool)
	toSet := make(map[string]bool)

	for _, e := range edges {
		fromSet[e.FromId] = true
		toSet[e.ToId] = true
	}

	var endNodes []string
	for node := range toSet {
		if !fromSet[node] {
			endNodes = append(endNodes, node)
		}
	}

	return endNodes
}
