package listener_lambda

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
	"github.com/tecmise/lambda-lib/pkg/ports/input"
	"reflect"
)

func NewListenerLambda[T any](exec func(context.Context, events.SQSMessage, T) error) input.TecmiseLambda {
	return &listener[T]{
		execution: exec,
	}
}

type listener[T any] struct {
	execution func(context.Context, events.SQSMessage, T) error
}

func (m listener[T]) Handler(ctx context.Context, event events.SQSEvent) error {
	if event.Records != nil {
		for _, record := range event.Records {
			logrus.Debugf("Processing record with MessageID %s", record.MessageId)
			logrus.Debugf("Time: %s", record.Attributes["ApproximateReceiveCount"])
			var result T
			err := json.Unmarshal([]byte(record.Body), &result)
			if err != nil {
				logrus.Errorf("Error unmarshalling record with MessageID %s: %v", record.MessageId, err)
				return err
			}
			logrus.Debugf("Record unmarshalled: %+v", result)

			method := reflect.ValueOf(result).MethodByName("Validate")
			if method.IsValid() {
				out := method.Call(nil)
				if len(out) == 1 && !out[0].IsNil() {
					errValidation := out[0].Interface().(error)
					logrus.Errorf("Error validating record with MessageID %s: %v", record.MessageId, errValidation)
					return errValidation
				}
				logrus.Debug("Record validated successfully")
			} else {
				logrus.Warnf("No Validate method found for record with MessageID %s", record.MessageId)
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
