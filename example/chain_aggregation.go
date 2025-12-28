package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bittoy/rule/builtin/aspect"
	"github.com/bittoy/rule/engine"
	"github.com/bittoy/rule/types"
)

var chainAggregationFile = `
{
    "id": "chainAggregation01",
    "name": "testRuleChain01",
    "type": "shortCircuit",
    "disabled": false,
    "metadata": {
        "chains": [
            {
                "id": "chain02",
                "type": "chain",
                "name": "testRuleChain01",
                "errOnTermination": false,
                "priority": 10,
                "privateParameter": [
                    {
                        "key": "pjws",
                        "type": "STRING",
                        "name": "裁判文书",
                        "desc": ""
                    }
                ],
                "disabled": false,
                "metadata": {
                    "nodes": [
                        {
                            "id": "s21",
                            "type": "rule",
                            "type": "start",
                            "name": "开始",
                            "configuration": {
                                "jsScript": "return msg.temperature>10;"
                            }
                        },
                        {
                            "id": "s22",
                            "type": "exprSwitch",
                            "name": "expr过滤",
                            "configuration": {
                                "script": "student==\"3\" ? \"1\" : \"default\""
                            }
                        },
                        {
                            "id": "s25",
                            "type": "exprAssign",
                            "name": "赋值",
                            "configuration": {
                                "script": "{\"score\": score*3,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
                            }
                        },
                        {
                            "id": "s27",
                            "type": "end",
                            "name": "赋值",
                            "configuration": {
                                "script": "{\"score\": (score)*priVars.score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student+3}"
                            }
                        }
                    ],
                    "connections": [
                        {
                            "fromId": "s21",
                            "toId": "s22",
                            "type": "default"
                        },
                        {
                            "fromId": "s22",
                            "toId": "s25",
                            "type": "1"
                        },
                        {
                            "fromId": "s22",
                            "toId": "s25",
                            "type": "default"
                        },
                        {
                            "fromId": "s25",
                            "toId": "s27",
                            "type": "default"
                        }
                    ]
                }
            },
            {
                "id": "chain03",
                "type": "chain",
                "name": "testRuleChain01",
                "errOnTermination": false,
                "priority": 20,
                "privateParameter": [
                    {
                        "key": "pjws",
                        "type": "STRING",
                        "name": "裁判文书",
                        "desc": ""
                    }
                ],
                "disabled": false,
                "metadata": {
                    "nodes": [
                        {
                            "id": "s31",
                            "type": "start",
                            "name": "开始",
                            "configuration": {
                                "jsScript": "return msg.temperature>10;"
                            }
                        },
                        {
                            "id": "s32",
                            "type": "exprSwitch",
                            "name": "expr过滤",
                            "configuration": {
                                "script": "student==\"3\" ? \"1\" : score > 75 && level == \"B\" || student == \"C\" ? \"2\" : score > 60 ? \"3\" : \"default\""
                            }
                        },
                        {
                            "id": "s35",
                            "type": "exprAssign",
                            "name": "赋值",
                            "configuration": {
                                "script": "{\"score\": (score + 2)*score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
                            }
                        },
                        {
                            "id": "s37",
                            "type": "end",
                            "name": "赋值",
                            "configuration": {
                                "script": "{\"score\": (score + 5)*priVars.score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
                            }
                        }
                    ],
                    "connections": [
                        {
                            "fromId": "s31",
                            "toId": "s32",
                            "type": "default"
                        },
                        {
                            "fromId": "s32",
                            "toId": "s35",
                            "type": "1"
                        },
                        {
                            "fromId": "s32",
                            "toId": "s35",
                            "type": "default"
                        },
                        {
                            "fromId": "s35",
                            "toId": "s37",
                            "type": "default"
                        }
                    ]
                }
            },
            {
                "id": "chain04",
                "type": "chain",
                "name": "testRuleChain01",
                "errOnTermination": false,
                "priority": 30,
                "privateParameter": [
                    {
                        "key": "pjws",
                        "type": "STRING",
                        "name": "裁判文书",
                        "desc": ""
                    }
                ],
                "disabled": false,
                "metadata": {
                    "nodes": [
                        {
                            "id": "s41",
                            "type": "start",
                            "name": "开始",
                            "configuration": {
                                "jsScript": "return msg.temperature>10;"
                            }
                        },
                        {
                            "id": "s42",
                            "type": "exprSwitch",
                            "name": "expr过滤",
                            "configuration": {
                                "script": "student==\"3\" ? \"1\" : score > 75 && level == \"B\" || student == \"C\" ? \"2\" : score > 60 ? \"3\" : \"default\""
                            }
                        },
                        {
                            "id": "s45",
                            "type": "exprAssign",
                            "name": "赋值",
                            "configuration": {
                                "script": "{\"score\": score*score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
                            }
                        },
                        {
                            "id": "s46",
                            "type": "exprAssign",
                            "name": "赋值",
                            "configuration": {
                                "script": "{\"score\": score*3,\"level\": score > 60 ? \"C\" : \"B\",\"tag\": student}"
                            }
                        },
                        {
                            "id": "s47",
                            "type": "end",
                            "name": "赋值",
                            "configuration": {
                                "script": "{\"score\": (score + 3)*priVars.score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
                            }
                        }
                    ],
                    "connections": [
                        {
                            "fromId": "s41",
                            "toId": "s42",
                            "type": "default"
                        },
                        {
                            "fromId": "s42",
                            "toId": "s45",
                            "type": "2"
                        },
                        {
                            "fromId": "s42",
                            "toId": "s45",
                            "type": "default"
                        },
                        {
                            "fromId": "s46",
                            "toId": "s47",
                            "type": "default"
                        },
                        {
                            "fromId": "s45",
                            "toId": "s46",
                            "type": "default"
                        }
                    ]
                }
            }
        ]
    }
}
`

func chainAggregationEngine(ruleChainFile string) {
	var properties = make(types.Properties)
	properties["env"] = "dev"
	config := engine.NewConfig(
		types.WithProperties(properties),
	)
	now := time.Now()
	ruleEngine, err := engine.NewChainAggregationEngine([]byte(chainAggregationFile), engine.WithConfig(config), engine.WithAspects(&aspect.NodeDebug{}, &aspect.ChainDebug{}, &aspect.ChainValidator{}, &aspect.ChainAggregationValidator{}))
	if err != nil {
		log.Fatalf("Failed to create rule engine: %v", err)
		return
	}
	fmt.Println("NewRuleEngine cost:", time.Since(now))
	defer ruleEngine.Stop()

	shardData := map[string]any{
		"student": 3,
		"level":   "B",
		"score":   80,
	}
	if err != nil {
		log.Fatalf("Failed to NewMapSharedData: %v", err)
		return
	}

	msg := types.NewRuleMsg("TELEMETRY_MSG", 0, shardData)

	err = ruleEngine.OnMsg(context.Background(), msg)
	if err != nil {
		log.Fatalf("Failed to OnMsg: %v", err)
	}
	fmt.Println("OnMsg data input:", msg.GetInput(), "OnMsg data output:", msg.GetChainAggregationOutput(), "OnMsg data chainPriority:", msg.GetChainAggregationPriority())
	fmt.Println("OnMsg cost:", time.Since(now))
}

func main() {
	chainAggregationEngine(chainAggregationFile)
}
