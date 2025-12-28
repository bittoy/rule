package types

import (
	"fmt"
)

type EngineError struct {
	ruleContext RuleContext
	msg         RuleMsg
	err         error
}

func (e *EngineError) Error() string {
	return fmt.Sprintf("EngineError: %s, relationType: %s", e.err.Error(), e.ruleContext.Self().DSL())
}

func NewEngineError(ruleContext RuleContext, msg RuleMsg, err error) *EngineError {
	return &EngineError{
		msg:         msg,
		err:         err,
		ruleContext: ruleContext,
	}
}
