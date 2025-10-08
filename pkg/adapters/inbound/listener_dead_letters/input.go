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
					Title:       fmt.Sprintf("🚨 [%s] Erro de processamento de fila (%s)", a.serviceName, record.Attributes["DeadLetterQueueSourceArn"]),
					Description: fmt.Sprintf("Houve um erro ao preocessa mensagem %s na fila", record.MessageId),
					Color:       16711680,
					Fields: []discord.Field{
						{Name: "Publisher", Value: record.Attributes["SenderId"], Inline: false},
						{Name: "Recebido", Value: record.Attributes["ApproximateFirstReceiveTimestamp"], Inline: false},
						{Name: "Tentativas", Value: record.Attributes["ApproximateReceiveCount"], Inline: false},
						{Name: "Content", Value: record.Body, Inline: false},
					},
					Footer: discord.Footer{
						Text: fmt.Sprintf("Data do processo • %s", time.Now().UTC().Format("2006-01-02 15:04 UTC")),
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

		logrus.Debugf("Status: %s", resp.Status)

		logrus.Debugf("MessageID %s", record.MessageId)
	}
	return nil
}
