package listener_dead_letters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tecmise/lambda-lib/pkg/adapters/inbound"
	"github.com/tecmise/lambda-lib/pkg/ports/output/discord"
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

func NewLambdaDlq(serviceName string) inbound.ListenerLambda {
	return &dlqLambda{
		serviceName: serviceName,
	}
}

func (a *dlqLambda) Handler(_ context.Context, event events.SQSEvent) error {
	logrus.Debugf("Content: %v", event)
	for _, record := range event.Records {
		logrus.Debugf("Processing record with MessageID %s", record.MessageId)

		webhookURL := os.Getenv("DISCORD_NOTIFIER_URL")
		if webhookURL == "" {
			logrus.Errorf("Webhook URL não definida")
			return errors.New("discord notifier url is null")
		}

		payload := discord.Payload{
			Embeds: []discord.Embed{
				{
					Title:       "🚨 Erro de processamento de fila",
					Description: fmt.Sprintf("Houve um erro ao processar a fila %s", record.EventSourceARN),
					Color:       16711680,
					Fields: []discord.Field{
						{Name: "Service", Value: a.serviceName, Inline: false},
						{Name: "QueueSource", Value: record.Attributes["DeadLetterQueueSourceArn"], Inline: false},
						{Name: "Publisher", Value: record.Attributes["SenderId"], Inline: false},
						{Name: "MessageId", Value: record.MessageId, Inline: true},
						{Name: "AWS Region", Value: record.AWSRegion, Inline: true},
						{Name: "Recebido em", Value: record.Attributes["ApproximateFirstReceiveTimestamp"], Inline: true},
						{Name: "Content", Value: record.Body, Inline: false},
						{Name: "ApproximateReceiveCount", Value: record.Attributes["ApproximateReceiveCount"], Inline: true},
					},
					Footer: discord.Footer{
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

		logrus.Debugf("Status:", resp.Status)

		logrus.Debugf("Email sent successfully for MessageID %s", record.MessageId)
	}
	return nil
}
