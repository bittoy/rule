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

var (
	shareKey       = "shareKey"
	shareValue     = "shareValue"
	addShareKey    = "addShareKey"
	addShareValue  = "addShareValue"
	testdataFolder = "../testdata/rule/"
)
var ruleChainFile0 = `{
                "id": "chain05",
                "type": "end"
            }`
var ruleChainFile1 = `{
            "id": "test01",
            "name": "testRuleChain01",
            "privateParameter": [
            {
                "key": "pjws",
                "type": "STRING",
                "name": "裁判文书",
                "desc": ""
            }],
          "disabled": false,
          "metadata": {
            "nodes": [
			  {
                "id": "s1",
                "type": "start",
                "name": "开始",
                "configuration": {
                  "jsScript": "return msg.temperature>10;"
                }
              },
              {
                "id": "s2",
                "type": "exprSwitch",
                "name": "expr过滤",
                "configuration": {
				  "cases": [
                        {
                            "case": "student==\"3\"",
                            "then": "\"1\""
                        },
                        {
                            "case": "score > 75 && level == \"B\" || student == \"C\"",
                            "then": "\"2\""
                        },
                        {
                            "case": "score > 60",
                            "then": "\"3\""
                        },
						{
                            "case": "other",
                            "then": "\"default\""
                        }
				  ],
				  "vars": {
                    "student": "student",
                    "score": 80,
                    "level": "11"
                  }
				}
              },
              {
                "id": "s3",
                "type": "jsSwitch",
                "name": "转换",
                "configuration": {
                  "script": "return msg.student === '3' && msg.level === 'B' ? '1' : msg.level === 'B' ? '2' : msg.score > 60 ? '3' : 'default';",
				   "cases": [
                        {
                            "case": "msg.student === '3' && msg.level === 'B'",
                            "then": "1"
                        },
                        {
                            "case": "msg.level === 'B'",
                            "then": "2"
                        },
                        {
                            "case": "msg.score > 60",
                            "then": "3"
                        },
						{
                            "case": "other",
                            "then": "default"
                        }
				  ]
                }
              },
              {
                "id": "s4",
                "type": "jsSwitch",
                "name": "转换",
                "configuration": {
                  "script": "return msg.student === '3' ? '1' : '2';"
                }
              },
              {
                "id": "s5",
                "type": "exprAssign",
                "name": "赋值",
                "configuration": {
                  "script": "{\"score\": (score + 10)*score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
                }
              },
			  {
                "id": "s6",
                "type": "exprFilter",
                "name": "赋值",
                "configuration": {
                  "script": "score > 10"
                }
              },
			  {
				"id": "s7",
				"type": "end",
				"name": "赋值",
				"configuration": {
					"script": "{\"score\": (score + 10)*priVars.score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
				}
              },
			  {
				"id": "s8",
				"type": "end",
				"name": "赋值",
				"configuration": {
					"script": "{\"score\": 2*priVars.score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
				}
              }
            ],
            "connections": [
              {
                "fromId": "s1",
                "toId": "s2",
                "type": "default"
              },
			  {
                "fromId": "s2",
                "toId": "s3",
                "type": "1"
              },
			  {
                "fromId": "s2",
                "toId": "s4",
                "type": "2"
              },
			  {
                "fromId": "s2",
                "toId": "s5",
                "type": "default"
              },
			  {
                "fromId": "s3",
                "toId": "s5",
                "type": "1"
              },
			   {
                "fromId": "s3",
                "toId": "s5",
                "type": "default"
              },
			  {
                "fromId": "s4",
                "toId": "s6",
                "type": "default"
              },
			  {
                "fromId": "s5",
                "toId": "s6",
                "type": "default"
              },
			  {
                "fromId": "s6",
                "toId": "s7",
                "type": "false"
              },
			  {
                "fromId": "s6",
                "toId": "s8",
                "type": "true"
              }
            ]
          }
        }`

// 修改metadata和msg 节点
var modifyMetadataAndMsgNode = `
	  {
			"id":"s2",
			"type": "jsTransform",
			"name": "转换",
			"debugMode": true,
			"configuration": {
			  "jsScript": "metadata['test']='test02';\n metadata['index']=50;\n msgType='TEST_MSG_TYPE_MODIFY';\n  msg['aa']=66;\n return {'msg':msg,'metadata':metadata,'msgType':msgType};"
			}
		  }
`

func testRuleEngine() {
	var properties = make(types.Properties)
	properties["env"] = "dev"
	config := engine.NewConfig(
		types.WithProperties(properties),
	)
	now := time.Now()
	ruleEngine, err := engine.NewChainEngine([]byte(ruleChainFile1), engine.WithConfig(config), engine.WithAspects(&aspect.NodeDebug{}, &aspect.ChainDebug{}, &aspect.ChainValidator{}, &aspect.ChainAggregationValidator{}))
	if err != nil {
		log.Fatalf("Failed to create rule engine: %v", err)
	}
	fmt.Println("NewRuleEngine cost:", time.Since(now))
	defer ruleEngine.Stop()

	shardData := map[string]any{
		"student": "3",
		"level":   "B",
		"score":   80,
	}
	if err != nil {
		log.Fatalf("Failed to NewMapSharedData: %v", err)
	}

	msg := types.NewRuleMsg("TELEMETRY_MSG", 0, shardData)

	err = ruleEngine.OnMsg(context.Background(), msg)
	if err != nil {
		log.Fatalf("Failed to OnMsg: %v", err)
	}
	fmt.Println("OnMsg1 data input:", msg.GetInput(), "output:", msg.GetChainOutput())
	fmt.Println("OnMsg1 cost:", time.Since(now))

	if err := ruleEngine.ReloadSelf([]byte(ruleChainFile2)); err != nil {
		log.Fatalf("Failed to create rule engine: %v", err)
	}
	err = ruleEngine.OnMsg(context.Background(), msg)
	if err != nil {
		log.Fatalf("Failed to OnMsg: %v", err)
	}
	fmt.Println("OnMsg2 data input:", msg.GetInput(), "output:", msg.GetChainOutput())
	fmt.Println("OnMsg2 cost:", time.Since(now))
}

func main() {
	testRuleEngine()
}

var ruleChainFile2 = `{
            "id": "test02",
            "name": "testRuleChain01",
            "privateParameter": [
            {
                "key": "pjws",
                "type": "STRING",
                "name": "裁判文书",
                "desc": ""
            }],
          "disabled": false,
          "metadata": {
            "nodes": [
			  {
                "id": "s11",
                "type": "start",
                "name": "开始",
                "configuration": {
                  "jsScript": "return msg.temperature>10;"
                }
              },
              {
                "id": "s12",
                "type": "exprSwitch",
                "name": "expr过滤",
                "configuration": {
                  "script": "student==\"3\" ? \"1\" : score > 75 && level == \"B\" || student == \"C\" ? \"2\" : score > 60 ? \"3\" : \"default\"",
				  "cases": [
                        {
                            "case": "student==\"3\"",
                            "then": "1"
                        },
                        {
                            "case": "score > 75 && level == \"B\"|| student == \"C\"",
                            "then": "2"
                        },
                        {
                            "case": "score > 60",
                            "then": "3"
                        },
						{
                            "case": "other",
                            "then": "default"
                        }
				  ],
				  "vars": {
                    "student": "student",
                    "score": 80,
                    "level": "11"
                  }
				}
              },
              {
                "id": "s13",
                "type": "jsSwitch",
                "name": "转换",
                "configuration": {
                  "script": "return msg.student === '3' && msg.level === 'B' ? '1' : msg.level === 'B' ? '2' : msg.score > 60 ? '3' : 'default';",
				   "cases": [
                        {
                            "case": "msg.student === '3' && msg.level === 'B'",
                            "then": "1"
                        },
                        {
                            "case": "msg.level === 'B'",
                            "then": "2"
                        },
                        {
                            "case": "msg.score > 60",
                            "then": "3"
                        },
						{
                            "case": "other",
                            "then": "default"
                        }
				  ]
                }
              },
              {
                "id": "s14",
                "type": "jsSwitch",
                "name": "转换",
                "configuration": {
                  "script": "return msg.student === '3' ? '1' : '2';"
                }
              },
              {
                "id": "s15",
                "type": "exprAssign",
                "name": "赋值",
                "configuration": {
                  "script": "{\"score\": (score + 10)*score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
                }
              },
			  {
                "id": "s16",
                "type": "exprAssign",
                "name": "赋值",
                "configuration": {
                  "script": "{\"score\": (score + 10)*priVars.score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
                }
              },
			  {
				"id": "s17",
				"type": "end",
				"name": "赋值",
				"configuration": {
					"script": "{\"score\": (score + 10)*priVars.score,\"level\": score > 60 ? \"A\" : \"B\",\"tag\": student}"
				}
              }
            ],
            "connections": [
              {
                "fromId": "s11",
                "toId": "s12",
                "type": "default"
              },
			  {
                "fromId": "s12",
                "toId": "s13",
                "type": "1"
              },
			  {
                "fromId": "s12",
                "toId": "s14",
                "type": "2"
              },
			  {
                "fromId": "s12",
                "toId": "s15",
                "type": "default"
              },
			  {
                "fromId": "s13",
                "toId": "s15",
                "type": "1"
              },
			   {
                "fromId": "s13",
                "toId": "s15",
                "type": "default"
              },
			  {
                "fromId": "s14",
                "toId": "s16",
                "type": "default"
              },
			  {
                "fromId": "s15",
                "toId": "s16",
                "type": "default"
              },
			  {
                "fromId": "s16",
                "toId": "s17",
                "type": "default"
              }
            ]
          }
        }`
