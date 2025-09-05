// main.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tecmise/lambda-lib/pkg/adapters/inbound"
	"github.com/tecmise/lambda-lib/pkg/ports/input"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type (
	dlqLambda struct {
		serviceName string
	}
)

func NewDlq(serviceName string) inbound.ListenerLambda {
	return &dlqLambda{
		serviceName: serviceName,
	}
}

func (a *dlqLambda) Handler(_ context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		logrus.Debugf("Processing record with MessageID %s", record.MessageId)

		webhookURL := os.Getenv("DISCORD_NOTIFIER_URL")
		if webhookURL == "" {
			fmt.Println("Webhook URL não definida")
		}

		payload := input.Payload{
			Embeds: []input.Embed{
				{
					Title:       "🚨 Erro de processamento de fila",
					Description: fmt.Sprintf("Houve um erro ao processar a fila %s", record.EventSourceARN),
					Color:       16711680,
					Fields: []input.Field{
						{Name: "MessageId:", Value: record.MessageId, Inline: true},
						{Name: "Event Source", Value: record.EventSource, Inline: true},
						{Name: "AWS Region", Value: record.AWSRegion, Inline: true},
						{Name: "Body", Value: record.Body, Inline: false},
						{Name: "ApproximateReceiveCount", Value: record.Attributes["ApproximateReceiveCount"], Inline: true},
					},
					Footer: input.Footer{
						Text: fmt.Sprintf("Notificao enviada pelo service %s • %s", a.serviceName, time.Now().UTC().Format("2006-01-02 15:04 UTC")),
					},
				},
			},
		}

		body, err := json.Marshal(payload)
		if err != nil {
			fmt.Println("Erro ao serializar JSON:", err)
			return err
		}

		resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println("Erro ao enviar para o Discord:", err)
			return err
		}
		defer resp.Body.Close()

		fmt.Println("Status:", resp.Status)

		logrus.Debugf("Email sent successfully for MessageID %s", record.MessageId)
	}
	return nil // Retornando nil, o Lambda entende que a mensagem foi processada
}
