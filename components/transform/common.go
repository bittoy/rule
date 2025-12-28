package transform

import (
	"errors"
	"strings"
)

import (
	"github.com/bittoy/rule/types"
)

func genExprScriptByCases(cases []types.Case) (string, error) {
	var script = strings.Builder{}

	for _, v := range cases {
		v.Case = strings.TrimSpace(v.Case)
		v.Then = strings.TrimSpace(v.Then)
		if len(v.Case) == 0 || len(v.Then) == 0 {
			return "", errors.New("case must not be empty")
		}
		if v.Case == "other" {
			script.WriteString(v.Then)
		} else {
			script.WriteString(v.Case)
			script.WriteString(" ? ")
			script.WriteString(v.Then)
			script.WriteString(" : ")
		}
	}
	return script.String(), nil
}
