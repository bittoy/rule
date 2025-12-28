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
	"time"

	"github.com/bittoy/rule/utils/maps"
	"github.com/gofrs/uuid/v5"
)

// Constants for keys used in message handling and metadata operations.
// These standardized keys ensure consistency across the rule engine.
// 用于消息处理和元数据操作的键常量。
// 这些标准化的键确保规则引擎的一致性。
const (
	IdKey       = "id"       // Key for the message unique identifier  消息唯一标识符的键
	TsKey       = "ts"       // Key for the message timestamp  消息时间戳的键
	DataKey     = "data"     // Key for the message content  消息内容的键
	MsgKey      = "msg"      // Key for the message content object  消息内容对象的键
	MetadataKey = "metadata" // Key for the message metadata  消息元数据的键
	MsgTypeKey  = "msgType"  // Key for the message type  消息类型的键
	DataTypeKey = "dataType" // Key for the data type of the message  消息数据类型的键
)

// Properties is a simple map type for storing key-value pairs as metadata.
// It provides basic operations for metadata management without Copy-on-Write optimization.
// This type is suitable for scenarios where performance is not critical or when
// metadata sharing between multiple instances is not required.
//
// Properties 是用于存储键值对作为元数据的简单映射类型。
// 它提供基本的元数据管理操作，但不包含写时复制优化。
// 此类型适用于性能不关键或不需要在多个实例间共享元数据的场景。
type Properties map[string]any

// NewProperties creates a new empty Properties instance.
// Returns an initialized Properties map ready for use.
//
// NewProperties 创建一个新的空 Properties 实例。
// 返回一个已初始化的 Properties 映射，可立即使用。
func NewProperties() Properties {
	return make(Properties)
}

// BuildProperties creates a new Properties instance from existing data.
// If the input data is nil, returns an empty Properties instance.
// The function creates a deep copy of the input data to ensure isolation.
//
// BuildProperties 从现有数据创建新的 Properties 实例。
// 如果输入数据为 nil，返回空的 Properties 实例。
// 该函数创建输入数据的深度副本以确保隔离。
func BuildProperties(data Properties) Properties {
	if data == nil {
		return make(Properties)
	}
	// Pre-allocate with known capacity to reduce map resizing
	// 预分配已知容量以减少映射重新调整大小
	metadata := make(Properties, len(data))
	for k, v := range data {
		metadata[k] = v
	}
	return metadata
}

// Copy creates a deep copy of the Properties.
// This ensures that modifications to the copy do not affect the original.
//
// Copy 创建 Properties 的深度副本。
// 这确保对副本的修改不会影响原始数据。
func (md Properties) Copy() Properties {
	return BuildProperties(md)
}

// Has checks if a key exists in the metadata.
// Returns true if the key exists, false otherwise.
//
// Has 检查元数据中是否存在键。
// 如果键存在返回 true，否则返回 false。
func (md Properties) Has(key string) bool {
	_, ok := md[key]
	return ok
}

// GetValue retrieves a value by key from the metadata.
// Returns the value if the key exists, or an empty string if not found.
//
// GetValue 通过键从元数据中检索值。
// 如果键存在返回值，如果未找到返回空字符串。
func (md Properties) GetValue(key string) any {
	v, _ := md[key]
	return v
}

// PutValue sets a value in the metadata.
// If the key is empty, the operation is ignored to prevent invalid entries.
//
// PutValue 在元数据中设置值。
// 如果键为空，操作将被忽略以防止无效条目。
func (md Properties) PutValue(key string, value any) {
	if key != "" {
		md[key] = value
	}
}

// Values returns the underlying map containing all key-value pairs.
// Note: This returns a direct reference to the internal map, so modifications
// will affect the original Properties instance.
//
// Values 返回包含所有键值对的底层映射。
// 注意：这返回内部映射的直接引用，因此修改会影响原始 Properties 实例。
func (md Properties) Values() map[string]any {
	return md
}

type RuleMsg struct {
	// Ts is the message timestamp in milliseconds since Unix epoch.
	// This field is automatically set when creating a new message if not provided.
	// Ts 是自 Unix 纪元以来的消息时间戳（毫秒）。
	// 如果未提供，创建新消息时会自动设置此字段。
	ts int64

	// Id is the unique identifier for the message as it flows through the rule engine.
	// Each message gets a UUID when created, ensuring uniqueness across the system.
	// Id 是消息在规则引擎中流转时的唯一标识符。
	// 每个消息在创建时都会获得一个 UUID，确保在系统中的唯一性。
	id string

	// Data contains the actual message payload using Copy-on-Write optimization.
	// The format of this data should match the DataType field.
	// Data 包含使用写时复制优化的实际消息负载。
	// 此数据的格式应与 DataType 字段匹配。
	data *RuleData
}

// SharedData represents a thread-safe copy-on-write data structure for message payload.
// This improved version addresses potential race conditions in the original implementation.
type RuleData struct {
	input                  map[string]any
	chainOutput            map[string]any
	chainAggregationOutput map[string]map[string]any
	aggregationOutput      map[string]any
}

// NewMsgWithJsonDataFromBytes creates a new message instance with JSON data from []byte.
func NewRuleMsg(id string, ts int64, data map[string]any) RuleMsg {
	return newRuleMsg(id, ts, data)
}

// newMsg is a helper function to create a new RuleMsg from []byte data.
func newRuleMsg(id string, ts int64, input map[string]any) RuleMsg {
	if ts <= 0 {
		ts = time.Now().UnixMilli()
	}
	if id == "" {
		uuId, _ := uuid.NewV4()
		id = uuId.String()
	}
	input["priVars"] = map[string]any{}
	// Create the message
	return RuleMsg{
		ts:   ts,
		id:   id,
		data: &RuleData{input: input},
	}
}

// IsEmpty checks if the data is empty.
func (sd *RuleMsg) GetInput() map[string]any {
	return sd.data.input
}

// IsEmpty checks if the data is empty.
func (sd *RuleMsg) CopyInnerData(priVars map[string]any) {
	maps.Copy(sd.data.input["priVars"].(map[string]any), priVars)
}

func (sd *RuleMsg) ClearInnerData() {
	sd.data.input["priVars"] = map[string]any{}
}

// IsEmpty checks if the data is empty.
func (sd *RuleMsg) SetChainOutput(output map[string]any) {
	sd.data.chainOutput = output
}

// IsEmpty checks if the data is empty.
func (sd *RuleMsg) GetChainOutput() map[string]any {
	return sd.data.chainOutput
}

// IsEmpty checks if the data is empty.
func (sd *RuleMsg) SetChainAggregationOutput(output map[string]map[string]any) {
	sd.data.chainAggregationOutput = output
}

// IsEmpty checks if the data is empty.
func (sd *RuleMsg) GetChainAggregationOutput() map[string]map[string]any {
	return sd.data.chainAggregationOutput
}

// IsEmpty checks if the data is empty.
func (sd *RuleMsg) SetAggregationOutput(aggregationOutput map[string]any) {
	sd.data.aggregationOutput = aggregationOutput
}

// IsEmpty checks if the data is empty.
func (sd *RuleMsg) GetAggregationOutput() map[string]any {
	return sd.data.aggregationOutput
}
