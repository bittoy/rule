package transform

import "strings"

type Case struct {
	Case string `json:"case"`
	Then string `json:"then"`
}

func genExprScriptByCases(cases []Case) string {
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
