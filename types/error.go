package types

import (
	"fmt"
)

type EngineError struct {
	nodeCtx NodeCtx
	msg     RuleMsg
	err     error
}

func (e *EngineError) Error() string {
	return fmt.Sprintf("EngineError: %s, input:%+v, nodeDSL: %s", e.err.Error(), e.msg.GetInput(), e.nodeCtx.DSL())
}

func NewEngineError(nodeCtx NodeCtx, msg RuleMsg, err error) *EngineError {
	return &EngineError{
		nodeCtx: nodeCtx,
		msg:     msg,
		err:     err,
	}
}
