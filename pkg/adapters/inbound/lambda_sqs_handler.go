package inbound

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
)

type (
	ListenerLambda interface {
		Handler(ctx context.Context, event events.SQSEvent) error
	}
)
