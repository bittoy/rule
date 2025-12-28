package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/expr-lang/expr"
)

func main() {
	env := map[string]any{
		"score": 50,
		"test":  2,
		"level": "B",
	}

	program, err := expr.Compile("student==\"3\" ? \"1\" : score > 75 && level == \"B\" || student == \"C\" ? \"2\" : score > 60 ? \"3\" : \"default\"")
	if err != nil {
		log.Fatalln("Compile", err)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		log.Fatalln("Run", err)
	}

	fmt.Println(output) // 输出: 60

	program, err = expr.Compile(`
		{"score": (score + 10)*test,"level": score > 60 ? "A" : "B","tag": level}`, expr.AsKind(reflect.Map))
	if err != nil {
		log.Fatalln("Compile", err)
	}

	output, err = expr.Run(program, env)
	if err != nil {
		log.Fatalln("Run", err)
	}
	fmt.Println(output)

	program, err = expr.Compile(`3>1`, expr.AsKind(reflect.Bool))
	if err != nil {
		log.Fatalln("Compile", err)
	}

	output, err = expr.Run(program, env)
	if err != nil {
		log.Fatalln("Run", err)
	}
	fmt.Println(output)

	// terminate 是否终止，不执行后面策略，白名单，黑名单，强规则使用
	program, err = expr.Compile(`{"terminate":true, "score":35, "action":"REJECT", "reason":"High risk ip", tags:["IP","GEO"]}`, expr.AsKind(reflect.Map))
	if err != nil {
		log.Fatalln("Compile", err)
	}

	output, err = expr.Run(program, env)
	if err != nil {
		log.Fatalln("Run", err)
	}
	fmt.Println(output)
}
