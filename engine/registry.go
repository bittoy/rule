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

package engine

import (
	"fmt"
	"sync"

	"rule/components/common"
	"rule/components/transform"
	"rule/types"
)

// PluginsSymbol is the symbol used to identify plugins in a Go plugin file.
const PluginsSymbol = "Plugins"

// Registry is the default registry for rule engine components.
var Registry = new(RuleComponentRegistry)

// init registers default components to the default component registry.
func init() {
	var components []types.Node
	// Append components from various packages to the components slice.
	components = append(components, common.Registry.Components()...)
	components = append(components, transform.Registry.Components()...)

	// Register all components to the default component registry.
	for _, node := range components {
		_ = Registry.Register(node)
	}
}

// RuleComponentRegistry is a registry for rule engine components.
type RuleComponentRegistry struct {
	// components is a map of rule engine node components.
	components map[types.NodeType]types.Node
	// RWMutex is a read/write mutex lock.
	sync.RWMutex
}

// Register adds a rule engine node component to the registry.
func (r *RuleComponentRegistry) Register(node types.Node) error {
	r.Lock()
	defer r.Unlock()
	if r.components == nil {
		r.components = make(map[types.NodeType]types.Node)
	}
	if _, ok := r.components[node.Type()]; ok {
		return fmt.Errorf("the component already exists. componentType=%s", node.Type())
	}
	r.components[node.Type()] = node

	return nil
}

// Unregister removes a component from the registry by its type or plugin name.
func (r *RuleComponentRegistry) Unregister(componentType types.NodeType) error {
	r.Lock()
	defer r.Unlock()
	var removed = false

	// Check if it's a component type
	if _, ok := r.components[componentType]; ok {
		// Delete the component
		delete(r.components, componentType)
		removed = true
	}

	if !removed {
		return fmt.Errorf("component not found. componentType=%s", componentType)
	} else {
		return nil
	}
}

// NewNode creates a new instance of a rule engine node component by its type.
func (r *RuleComponentRegistry) NewNode(componentType types.NodeType) (types.Node, error) {
	r.RLock()
	defer r.RUnlock()

	if node, ok := r.components[componentType]; !ok {
		return nil, fmt.Errorf("component not found. componentType=%s", componentType)
	} else {
		return node.New(), nil
	}
}

// GetComponents returns a map of all registered components.
func (r *RuleComponentRegistry) GetComponents() map[types.NodeType]types.Node {
	r.RLock()
	defer r.RUnlock()
	var components = map[types.NodeType]types.Node{}
	for k, v := range r.components {
		components[k] = v
	}
	return components
}
