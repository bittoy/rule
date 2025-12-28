package main

import (
	"fmt"
	"github.com/dop251/goja"
)

var jsScript1 = `
	function jsSwitch(msg, out) { out.a = "c"; } jsSwitch;
`
var jsScript2 = `
	function calc(a, b) { return a * b + 10; } calc;
`

func main() {
	vm := goja.New()

	// 预编译脚本
	prog, _ := goja.Compile("func.js", jsScript1, true)

	fnVal, _ := vm.RunProgram(prog)
	calcFunc, _ := goja.AssertFunction(fnVal)

	// 多次复用
	for i := 1; i <= 3; i++ {
		res, _ := calcFunc(goja.Undefined(), vm.ToValue(i), vm.ToValue(5))
		fmt.Println("res =", res)
	}

	// 多次复用
	for i := 1; i <= 3; i++ {
		res, _ := calcFunc(goja.Undefined(), vm.ToValue(i), vm.ToValue(5))
		fmt.Println("res =", res)
	}

	var vars = map[string]any{"a": "b"}

}
