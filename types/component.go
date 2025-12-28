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
	"sync"
)

// Component kind constants define the different types of components in the RuleGo ecosystem.
// 组件类型常量定义了 RuleGo 生态系统中不同类型的组件。
const (
	// ComponentKindDynamic represents a dynamic component that can be loaded at runtime
	// ComponentKindDynamic 表示可以在运行时加载的动态组件
	ComponentKindDynamic string = "dc"

	// ComponentKindNative represents a native component that is built into the system
	// ComponentKindNative 表示内置在系统中的原生组件
	ComponentKindNative string = "nc"

	// ComponentKindEndpoint represents an endpoint component for input/output operations
	// ComponentKindEndpoint 表示用于输入/输出操作的端点组件
	ComponentKindEndpoint string = "ec"
)

// CategoryGetter is an optional interface that components can implement to provide
// category information for organizing components in visual tools.
//
// CategoryGetter 是组件可以实现的可选接口，用于提供分类信息，
// 在可视化工具中组织组件。
type CategoryGetter interface {
	// Category returns the category name for this component
	// Category 返回此组件的类别名称
	Category() string
}

// DescGetter is an optional interface that components can implement to provide
// a description of the component's functionality.
//
// DescGetter 是组件可以实现的可选接口，用于提供组件功能的描述。
type DescGetter interface {
	// Desc returns a description of the component
	// Desc 返回组件的描述
	Desc() string
}

// SafeComponentSlice provides a thread-safe slice for storing Node components.
// It uses mutex synchronization to ensure safe concurrent access.
//
// SafeComponentSlice 提供了用于存储 Node 组件的线程安全切片。
// 它使用互斥锁同步来确保安全的并发访问。
type SafeComponentSlice struct {
	// components holds the list of Node components
	// components 保存 Node 组件列表
	components []Node
	sync.Mutex
}

// Add safely appends one or more Node components to the slice.
// This method is thread-safe and can be called concurrently.
//
// Add 安全地将一个或多个 Node 组件追加到切片中。
// 此方法是线程安全的，可以并发调用。
func (p *SafeComponentSlice) Add(nodes ...Node) {
	p.Lock()
	defer p.Unlock()
	for _, node := range nodes {
		p.components = append(p.components, node)
	}
}

// Components returns a copy of the current component list.
// This method is thread-safe and returns a snapshot of the components.
//
// Components 返回当前组件列表的副本。
// 此方法是线程安全的，返回组件的快照。
func (p *SafeComponentSlice) Components() []Node {
	p.Lock()
	defer p.Unlock()
	return p.components
}
