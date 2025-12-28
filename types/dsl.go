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

type NodeType string

const (
	// ChainAggregationType
	AggShortCircuit NodeType = "shortCircuit" // 触发即返回
	AggParallel     NodeType = "parallel"     // 并行执行
	AggTable        NodeType = "policyTable"  // 表驱动

	// rule
	RuleSubTypeStart      NodeType = "start"
	RuleSubTypeEnd        NodeType = "end"
	RuleSubTypeJsSwitch   NodeType = "jsSwitch"
	RuleSubTypeExprSwitch NodeType = "exprSwitch"
	RuleSubTypeJsFilter   NodeType = "jsFilter"
	RuleSubTypeExprFilter NodeType = "exprFilter"
	RuleSubTypeExprAssign NodeType = "exprAssign"
)

type ChainAggregation struct {
	BaseInfo
	Metadata ChainMetadata `json:"metadata"`
}

type ChainMetadata struct {
	Chains []*Chain `json:"chains"`
	// Connections define the connections between two nodes in the rule chain.
	// Connections 定义规则链中两个节点之间的连接。
	//
	// Connections establish the message flow topology by specifying how messages
	// move from one node to another based on processing results and relationship types.
	// 连接通过指定消息如何基于处理结果和关系类型从一个节点移动到另一个节点来建立消息流拓扑。
	Connections []NodeConnection `json:"connections"`
}

type Chain struct {
	BaseInfo
	Metadata RuleMetadata `json:"metadata"`
}

type BaseInfo struct {
	// ID is the unique identifier of the rule chain.
	// ID 是规则链的唯一标识符。
	//
	// The ID must be unique within the rule engine context and is used for
	// chain references, sub-chain invocation, and management operations.
	// ID 在规则引擎上下文中必须是唯一的，用于链引用、子链调用和管理操作。
	Id string `json:"id"`

	// Name is the name of the rule chain.
	// Name 是规则链的名称。
	//
	// The name provides a human-readable identifier for the chain, useful
	// for UI display, logging, and administrative purposes.
	// 名称为链提供人类可读的标识符，对 UI 显示、日志记录和管理目的很有用。
	Name string `json:"name"`

	Type NodeType `json:"type"`

	Version string `json:"version"` // 字典序

	Timestamp string `json:"timestamp"` // 字典序
	// Disabled indicates whether the rule chain is disabled.
	// Disabled 表示规则链是否被禁用。
	//
	// When disabled, the rule chain will not process messages and can be used
	// for maintenance, testing, or gradual rollout scenarios.
	// 禁用时，规则链不会处理消息，可用于维护、测试或渐进式推出场景。
	Disabled bool `json:"disabled"`

	// 策略组优先级，按照优先级大小排序依次执行
	Priority int `json:"priority"`

	// 出错终止
	TerminalOnErr bool `json:"terminalOnErr"`

	Configuration Configuration `json:"configuration,omitempty"`
}

// RuleMetadata defines the metadata of a rule chain, including information about nodes and connections.
// RuleMetadata 定义规则链的元数据，包括节点和连接的信息。
//
// This structure contains the detailed topology and routing information that defines
// how messages flow through the rule chain. It includes node definitions, connections
// between nodes, endpoint configurations, and legacy sub-chain connections.
// 此结构包含定义消息如何通过规则链流动的详细拓扑和路由信息。
// 它包括节点定义、节点间连接、端点配置和传统子链连接。
//
// Structural Components:
// 结构组件：
//   - FirstNodeIndex: Entry point identification
//     FirstNodeIndex：入口点标识
//   - Endpoints: External connectivity configuration
//     Endpoints：外部连接配置
//   - Nodes: Processing component definitions
//     Nodes：处理组件定义
//   - Connections: Inter-node message routing
//     Connections：节点间消息路由
//   - RuleChainConnections: Legacy sub-chain integration
//     RuleChainConnections：传统子链集成
type RuleMetadata struct {
	// Nodes are the component definitions of the nodes.
	// Each object represents a rule node within the rule chain.
	// Nodes 是节点的组件定义。
	// 每个对象表示规则链内的一个规则节点。
	//
	// Nodes define the processing components that transform, filter, route, and
	// act upon messages as they flow through the rule chain. Each node encapsulates
	// specific business logic or integration functionality.
	// 节点定义在消息通过规则链流动时对消息进行转换、过滤、路由和操作的处理组件。
	// 每个节点封装特定的业务逻辑或集成功能。
	Nodes []*BaseInfo `json:"nodes"`

	// Connections define the connections between two nodes in the rule chain.
	// Connections 定义规则链中两个节点之间的连接。
	//
	// Connections establish the message flow topology by specifying how messages
	// move from one node to another based on processing results and relationship types.
	// 连接通过指定消息如何基于处理结果和关系类型从一个节点移动到另一个节点来建立消息流拓扑。
	Connections []NodeConnection `json:"connections"`
}

// NodeAdditionalInfo is used for visualization position information (reserved field).
// NodeAdditionalInfo 用于可视化位置信息（保留字段）。
//
// This structure defines the standard additional information fields used by
// visual rule chain editors for node positioning and documentation.
// 此结构定义了可视化规则链编辑器用于节点定位和文档的标准附加信息字段。
type NodeAdditionalInfo struct {
	// Description provides detailed documentation for the node
	// Description 为节点提供详细文档
	Description string `json:"description"`
	// LayoutX represents the horizontal position in the visual editor
	// LayoutX 表示可视化编辑器中的水平位置
	LayoutX int `json:"layoutX"`
	// LayoutY represents the vertical position in the visual editor
	// LayoutY 表示可视化编辑器中的垂直位置
	LayoutY int `json:"layoutY"`
}

// NodeConnection defines the connection between two nodes in a rule chain.
// NodeConnection 定义规则链中两个节点之间的连接。
//
// NodeConnection establishes the message flow topology by specifying how messages
// move from one processing node to another based on the results of message processing.
// The connection type determines the conditions under which messages flow.
// NodeConnection 通过指定消息如何基于消息处理结果从一个处理节点移动到另一个节点来建立消息流拓扑。
// 连接类型决定消息流动的条件。
//
// Connection Flow Logic:
// 连接流逻辑：
//  1. Source node processes message
//     源节点处理消息
//  2. Processing result determines relationship type
//     处理结果决定关系类型
//  3. Message is routed to target node if relationship matches
//     如果关系匹配，消息路由到目标节点
//  4. Multiple connections enable parallel or conditional flows
//     多个连接支持并行或条件流
//
// Common Connection Types:
// 常见连接类型：
//   - Success/Failure: General processing outcomes
//     Success/Failure：一般处理结果
//   - True/False: Boolean logic for filters and conditions
//     True/False：过滤器和条件的布尔逻辑
//   - Custom types: Domain-specific routing logic
//     自定义类型：领域特定的路由逻辑
type NodeConnection struct {
	// FromId is the id of the source node, which should match the id of a node in the nodes array.
	// FromId 是源节点的 id，应匹配节点数组中节点的 id。
	//
	// This field establishes the starting point of the message flow connection.
	// The referenced node must exist in the rule chain's node list.
	// 此字段建立消息流连接的起始点。
	// 引用的节点必须存在于规则链的节点列表中。
	FromId string `json:"fromId"`

	// ToId is the id of the target node, which should match the id of a node in the nodes array.
	// ToId 是目标节点的 id，应匹配节点数组中节点的 id。
	//
	// This field establishes the destination of the message flow connection.
	// The referenced node must exist in the rule chain's node list.
	// 此字段建立消息流连接的目标。
	// 引用的节点必须存在于规则链的节点列表中。
	ToId string `json:"toId"`

	// Type is the type of connection, which determines when and how messages are sent from one node to another. It should match one of the connection types supported by the source node type.
	// For example, a JS filter node might support two connection types: "True" and "False," indicating whether the message passes or fails the filter condition.
	// Type 是连接类型，决定何时以及如何将消息从一个节点发送到另一个节点。它应匹配源节点类型支持的连接类型之一。
	// 例如，JS 过滤器节点可能支持两种连接类型："True" 和 "False"，表示消息是通过还是未通过过滤条件。
	//
	// The type acts as a conditional gate that controls message flow based on
	// processing results. Each node type defines its own set of supported relationship types.
	// 类型作为基于处理结果控制消息流的条件门。
	// 每种节点类型定义自己支持的关系类型集。
	Type string `json:"type"`

	// Label is the label of the connection, used for display.
	// Label 是连接的标签，用于显示。
	//
	// The label provides a human-readable description of the connection,
	// useful for visual editors and documentation purposes.
	// 标签提供连接的人类可读描述，
	// 对于可视化编辑器和文档目的很有用。
	Label string `json:"label,omitempty"`
}

// RuleChainConnection defines the connection between a node and a sub-rule chain.
// RuleChainConnection 定义节点和子规则链之间的连接。
//
// This structure represents the legacy way of connecting to sub-rule chains
// directly. Modern implementations should use Flow nodes instead for better
// flexibility and consistency.
// 此结构表示直接连接到子规则链的传统方式。
// 现代实现应使用 Flow 节点以获得更好的灵活性和一致性。
type RuleChainConnection struct {
	// FromId is the id of the source node, which should match the id of a node in the nodes array.
	// FromId 是源节点的 id，应匹配节点数组中节点的 id。
	FromId string `json:"fromId"`
	// ToId is the id of the target sub-rule chain, which should match one of the sub-rule chains registered in the rule engine.
	// ToId 是目标子规则链的 id，应匹配规则引擎中注册的子规则链之一。
	ToId string `json:"toId"`
	// Type is the type of connection, which determines when and how messages are sent from one node to another. It should match one of the connection types supported by the source node type.
	// Type 是连接类型，决定何时以及如何将消息从一个节点发送到另一个节点。它应匹配源节点类型支持的连接类型之一。
	Type string `json:"type"`
}
