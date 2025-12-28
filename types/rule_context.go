package types

import (
	"context"
)

// RuleContext is the interface for message processing context within the rule engine.
// It handles the transfer of messages to the next or multiple nodes and triggers their business logic.
// It also controls and orchestrates the node flow of the current execution instance.
type RuleContext interface {
	// TellNext sends the message to the next node using the specified relationTypes.
	Tell(ctx context.Context, msg RuleMsg, relationType string) error
	// TellNext sends the message to the next node using the specified relationTypes.
	TellNext(ctx context.Context, msg RuleMsg, relationType string) error
	// Self retrieves the current node instance.
	Self() NodeCtx
	// From retrieves the node instance from which the message entered the current node.
	From() NodeCtx
}
