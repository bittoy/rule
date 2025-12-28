package main

import (
	"fmt"
)

type Node struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type Conn struct {
	From string `json:"fromId"`
	To   string `json:"toId"`
	Type string `json:"type"`
}

func validateRuleChain(nodes []Node, conns []Conn) error {

	// 收集 outgoing edges
	out := map[string][]Conn{}
	for _, c := range conns {
		out[c.From] = append(out[c.From], c)
	}

	// 需要严格限制为“只能 default”
	strictNodes := map[string]bool{
		"start":      true,
		"exprAssign": true,
	}

	for _, n := range nodes {

		if !strictNodes[n.Type] {
			continue
		}

		outs := out[n.Id]
		if len(outs) == 0 {
			return fmt.Errorf("节点 %s(%s) 必须有且仅有一个 default 连接，但当前没有任何连接", n.Id, n.Type)
		}

		// 必须只有 default
		for _, c := range outs {
			if c.Type != "default" {
				return fmt.Errorf("节点 %s(%s) 只能包含 type=default 的连接，但发现了: %s", n.Id, n.Type, c.Type)
			}
		}
	}

	return nil
}

func main() {

	nodes := []Node{
		{"s1", "start", "开始"},
		{"s2", "exprSwitch", "expr过滤"},
		{"s3", "jsSwitch", "转换"},
		{"s4", "jsSwitch", "转换"},
		{"s5", "exprAssign", "赋值"},
		{"s6", "exprAssign", "赋值"},
		{"s7", "end", "结束"},
	}

	conns := []Conn{
		{"s1", "s2", "default"}, // OK
		{"s1", "s3", "default"}, // OK
		{"s2", "s3", "1"},
		{"s2", "s4", "2"},
		{"s3", "s5", "1"},
		{"s5", "s6", "default"}, // OK
		{"s6", "s7", "default"}, // OK
		{"s6", "s1", "default"}, // OK
	}

	err := validateRuleChain(nodes, conns)
	if err != nil {
		fmt.Println("校验失败:", err)
	} else {
		fmt.Println("校验通过")
	}
}
