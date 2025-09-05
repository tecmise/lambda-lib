package input

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
)

type TecmiseLambda interface {
	Handler(ctx context.Context, event events.SQSEvent) error
}
