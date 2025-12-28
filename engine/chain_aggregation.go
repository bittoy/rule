package engine

import (
	"context"
	"fmt"
	"sort"

	"rule/types"
	"rule/utils/maps"
)

type Configuration struct {
	Aggregation Aggregation `json:"aggregation"`
}

type Aggregation struct {
	Type       string `json:"type"`
	Method     string `json:"method"`
	Thresholds struct {
		Action    string `json:"action"`
		ScoreExpr string `json:"scoreExpr"`
	} `json:"thresholds"`
}

type ChainAggregationCtx struct {
	// SelfDefinition contains the complete rule chain definition including
	// metadata, nodes, connections, and configuration
	// SelfDefinition 包含完整的规则链定义，包括元数据、节点、连接和配置
	selfDefinition *types.ChainAggregation

	// config contains the rule engine configuration including
	// component registry, parser, and global settings
	// config 包含规则引擎配置，包括组件注册表、解析器和全局设置
	config types.Config

	// nodes maps node identifiers to their corresponding node contexts,
	// providing O(1) lookup time for node access operations
	// nodes 将节点标识符映射到其对应的节点上下文，为节点访问操作提供 O(1) 查找时间
	chains []types.ChainCtx

	// nodeRoutes maps each node to its outgoing relationships,
	// defining the flow of messages through the rule chain
	// nodeRoutes 将每个节点映射到其传出关系，定义消息通过规则链的流动
	chainRoutes map[string]types.ChainCtx

	// aspects contains the list of AOP aspects applied to this rule chain,
	// providing cross-cutting concerns like logging, validation, and metrics
	// aspects 包含应用于此规则链的 AOP 切面列表，提供如日志、验证和指标等横切关注点
	aspects types.AspectList

	beforeAspects []types.ChainBeforeAspect
	afterAspects  []types.ChainAfterAspect

	configuration Configuration
}

func InitChainAggregationCtx(config types.Config, aspects types.AspectList, chainAggregationDef *types.ChainAggregation) (*ChainAggregationCtx, error) {
	// Initialize a new RuleChainCtx with the provided configuration and aspects
	// Retrieve aspects for the engine
	onChainAggregationBeforeInitAspects := aspects.GetOnChainAggregationBeforeInitAspects()
	for _, aspect := range onChainAggregationBeforeInitAspects {
		if err := aspect.OnChainAggregationBeforeInit(config, chainAggregationDef); err != nil {
			return nil, err
		}
	}

	var chainAggregationCtx = &ChainAggregationCtx{
		config:         config,
		selfDefinition: chainAggregationDef,
		chainRoutes:    map[string]types.ChainCtx{},
		aspects:        aspects,
	}

	sort.Slice(chainAggregationDef.Metadata.Chains, func(i, j int) bool {
		return chainAggregationDef.Metadata.Chains[i].Priority > chainAggregationDef.Metadata.Chains[j].Priority
	})

	for _, chain := range chainAggregationDef.Metadata.Chains {
		chainCtx, err := InitChainCtx(config, aspects, chain)
		if err != nil {
			return nil, err
		}
		chainAggregationCtx.chainRoutes[chain.Id] = chainCtx
		chainAggregationCtx.chains = append(chainAggregationCtx.chains, chainCtx)
	}

	chainAggregationCtx.beforeAspects, chainAggregationCtx.afterAspects = aspects.GetChainAspects()

	err := maps.Map2Struct(chainAggregationDef.Configuration, &chainAggregationCtx.configuration)
	if err != nil {
		return nil, err
	}

	return chainAggregationCtx, nil
}

// Config returns the configuration of the rule chain context
func (rc *ChainAggregationCtx) Config() types.Config {
	return rc.config
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainAggregationCtx) Id() string {
	return rc.selfDefinition.Id
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainAggregationCtx) Name() string {
	return rc.selfDefinition.Name
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainAggregationCtx) TerminalOnErr() bool {
	return rc.selfDefinition.TerminalOnErr
}

// GetNodeById retrieves a node context by its ID
func (rc *ChainAggregationCtx) GetChainById(id string) (types.ChainCtx, bool) {
	chainCtx, ok := rc.chainRoutes[id]
	return chainCtx, ok
}

// Type returns the component type
func (rc *ChainAggregationCtx) Type() types.NodeType {
	return rc.selfDefinition.Type
}

// New creates a new instance (not supported for RuleChainCtx)
func (rc *ChainAggregationCtx) New() types.Node {
	panic("not support New method")
}

// Init initializes the rule chain context
func (rc *ChainAggregationCtx) Init(_ types.Config, configuration types.Configuration) error {
	return fmt.Errorf("RuleChainCtx cant not init")
}

// OnMsg processes incoming messages
func (rc *ChainAggregationCtx) OnMsg(ctx context.Context, rCtx types.RuleContext, msg types.RuleMsg) error {
	var output = map[string]map[string]any{}
	for _, chain := range rc.chains {
		msg, err := rc.onBefore(chain, msg)
		if err != nil {
			return err
		}
		if err = chain.OnMsg(ctx, rCtx, msg); err != nil {
			if chain.TerminalOnErr() {
				return err
			} else {
				fmt.Printf("chain:%s, err%v\n", chain.Id(), err)
			}
		}
		msg, err = rc.onAfter(chain, msg)
		if err != nil {
			return err
		}
		output[chain.Id()] = msg.GetChainOutput()
	}
	msg.SetChainOutput(nil)
	msg.SetChainAggregationOutput(output)
	return nil
}

// Destroy cleans up resources and executes destroy aspects
func (rc *ChainAggregationCtx) Destroy() {
	// Execute destroy aspects without holding locks
	// Note: We avoid calling methods that need locks within OnDestroy by pre-fetching data
	for _, chain := range rc.chains {
		chain.Destroy()
	}
}

// DSL returns the rule chain definition as a byte slice
func (rc *ChainAggregationCtx) DSL() []byte {
	v, _ := rc.config.Parser.EncodeChainAggregation(rc.selfDefinition)
	return v
}

func (e *ChainAggregationCtx) onBefore(chain types.ChainCtx, msg types.RuleMsg) (types.RuleMsg, error) {
	var err error
	for _, aop := range e.beforeAspects {
		if aop.PointCut(NewChainContext(chain), msg) {
			msg, err = aop.Before(NewChainContext(chain), msg)
		}
	}
	return msg, err
}

// onEnd executes the list of end aspects when a branch of the rule chain ends.
// onEnd 在规则链分支结束时执行结束切面列表。
func (e *ChainAggregationCtx) onAfter(chain types.ChainCtx, msg types.RuleMsg) (types.RuleMsg, error) {
	var err error
	for _, aop := range e.afterAspects {
		if aop.PointCut(NewChainContext(chain), msg) {
			msg, err = aop.After(NewChainContext(chain), msg)
		}
	}
	return msg, err
}
