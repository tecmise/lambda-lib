package listener_lambda

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
	"github.com/tecmise/lambda-lib/pkg/ports/input"
	"github.com/tecmise/lambda-lib/pkg/ports/input/queue"
)

func NewListenerLambda[T queue.QObject](exec func(context.Context, events.SQSMessage, T) error) input.TecmiseLambda {
	return &listener[T]{
		execution: exec,
	}
}

type listener[T queue.QObject] struct {
	execution func(context.Context, events.SQSMessage, T) error
}

func (m listener[T]) Handler(ctx context.Context, event events.SQSEvent) error {
	if event.Records != nil {
		logrus.Debugf("Received %d records", len(event.Records))
		for _, record := range event.Records {
			logrus.Debugf("Processing record with MessageID %s", record.MessageId)
			var result T
			err := json.Unmarshal([]byte(record.Body), &result)
			if err != nil {
				logrus.Errorf("Error unmarshalling record with MessageID %s: %v", record.MessageId, err)
				return err
			}
			logrus.Debugf("Record unmarshalled: %+v", result)

			if errValidation := result.Validate(); errValidation != nil {
				logrus.Errorf("Error validating record with MessageID %s: %v", record.MessageId, errValidation)
				return err
			}
			logrus.Debug("Record validated successfully")

			if m.execution == nil {
				logrus.Errorf("No execution function defined for record with MessageID %s", record.MessageId)
				return errors.New("there's no exection function defined")
			}

			return m.execution(ctx, record, result)
		}
	}
	logrus.Debug("All records processed successfully")
	return nil
}
