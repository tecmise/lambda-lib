package input

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
)

type TecmiseLambda interface {
	Handler(ctx context.Context, event events.SQSEvent) error
}

type SNSMessage struct {
	Type           string `json:"Type"`
	MessageId      string `json:"MessageId"`
	SequenceNumber string `json:"SequenceNumber"`
	TopicArn       string `json:"TopicArn"`
	Message        string `json:"Message"`
	Timestamp      string `json:"Timestamp"`
	UnsubscribeURL string `json:"UnsubscribeURL"`
}
