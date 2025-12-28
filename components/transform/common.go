package transform

import "strings"

import (
	"github.com/bittoy/rule/types"
)

func genExprScriptByCases(cases []types.Case) string {
	var script = strings.Builder{}

	for _, v := range cases {
		if v.Case == "other" {
			script.WriteString(v.Then)
		} else {
			script.WriteString(v.Case)
			script.WriteString(" ? ")
			script.WriteString(v.Then)
			script.WriteString(" : ")
		}
	}
	return script.String()
}
