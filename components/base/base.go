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

// Package base provides foundational components and utilities for the RuleGo rule engine.
package base

import (
	"errors"
	"reflect"
	"strings"

	"github.com/bittoy/rule/types"
)

var (
	ErrNodePoolNil   = errors.New("node pool is nil")
	ErrClientNotInit = errors.New("client not init")
)

var NodeUtils = &nodeUtils{}

type nodeUtils struct {
}

func (n *nodeUtils) GetVars(configuration types.Configuration) (map[string]any, error) {
	if v, ok := configuration[types.Vars]; ok {
		if !IsMap(v) {
			return nil, errors.New("vars is not map")
		} else {
			return v.(map[string]any), nil
		}
	}
	return nil, nil

}

// IsMap 判断任意变量是否是 map
func IsMap(v any) bool {
	return reflect.TypeOf(v).Kind() == reflect.Map
}

func (n *nodeUtils) IsNodePool(config types.Config, server string) bool {
	return strings.HasPrefix(server, types.NodeConfigurationPrefixInstanceId)
}

func (n *nodeUtils) GetInstanceId(config types.Config, server string) string {
	if n.IsNodePool(config, server) {
		//截取资源ID
		return server[len(types.NodeConfigurationPrefixInstanceId):]
	}
	return ""
}

func (n *nodeUtils) IsInitNetResource(_ types.Config, configuration types.Configuration) bool {
	_, ok := configuration[types.NodeConfigurationKeyIsInitNetResource]
	return ok
}

// TrimStrings 去除配置中所有字符串值的前后空格
// 遍历 Configuration 中的所有值，如果是字符串类型则去除前后空格
func (n *nodeUtils) TrimStrings(config types.Configuration) {
	for key, value := range config {
		if strValue, ok := value.(string); ok {
			config[key] = strings.TrimSpace(strValue)
		}
	}
}

// isZeroValue 检查值是否为零值
// 使用反射来安全地比较值，避免在不可比较类型上出现运行时恐慌
func isZeroValue[T any](v T) bool {
	// 使用反射来安全地检查零值
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}
	return rv.IsZero()
}

// zeroValue 函数用于返回 T 类型的零值
func zeroValue[T any]() T {
	var zero T
	return zero
}
