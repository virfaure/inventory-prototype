package consumer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"log"
	"encoding/json"
	"github.com/magento-mcom/inventory-prototype/util"
	"github.com/magento-mcom/inventory-prototype/configuration"
	"fmt"
	"bytes"
	"text/template"
	"io/ioutil"
)

var sqsSession *sqs.SQS

func NewSQSConsumer(config configuration.Config) Consumer {
	return &sqsConsumer{config}
}

type sqsConsumer struct {
	config configuration.Config
}

type Message struct{}

func (consumer *sqsConsumer) getSqsSession() (*sqs.SQS) {
	if sqsSession == nil {
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(consumer.config.Consumer.Region),
			Credentials: credentials.NewSharedCredentials("", consumer.config.Consumer.Profile),
		})

		if err != nil {
			log.Fatal(err)
		}

		sqsSession = sqs.New(sess)
	}

	return sqsSession
}

func (consumer *sqsConsumer) PollStockMessages() []util.StockMovement {
	svc := consumer.getSqsSession()

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              &consumer.config.Consumer.Queuestock,
		MessageAttributeNames: aws.StringSlice([]string{"client"}),
		MaxNumberOfMessages:   aws.Int64(10),
		VisibilityTimeout:     aws.Int64(36000), // 10 hours
		WaitTimeSeconds:       aws.Int64(10),
	})

	util.Check(err)

	if len(result.Messages) == 0 {
		log.Println("No message in the queue")
		return nil
	}

	var buffer []util.StockMovement

	for _, message := range result.Messages {
		if message == nil {
			continue
		}

		if message.MessageAttributes["client"] == nil {
			log.Println("No Attribute client in message")
			continue
		}

		var sourceStockUpdate util.SourceStockUpdateJsonRpc
		if err := json.Unmarshal([]byte(*message.Body), &sourceStockUpdate); err != nil {
			log.Printf("error unmarshalling: %v", err)
			continue
		}

		var stockMovement util.StockMovement

		switch {
		case sourceStockUpdate.Params.Snapshot != nil:
			stockMovement = util.StockMovement{
				Sku:      sourceStockUpdate.Params.Snapshot.Stock[0].Sku,
				Quantity: sourceStockUpdate.Params.Snapshot.Stock[0].Quantity,
				Type:     "snapshot",
				Client:   *message.MessageAttributes["client"].StringValue,
				Source:   sourceStockUpdate.Params.Snapshot.SourceId,
				Date:     sourceStockUpdate.Params.Snapshot.CreatedOn,
				Reason:   "stock update",
			}
		case sourceStockUpdate.Params.Adjustment != nil:
			stockMovement = util.StockMovement{
				Sku:      sourceStockUpdate.Params.Adjustment.StockAdjustment[0].Sku,
				Quantity: sourceStockUpdate.Params.Adjustment.StockAdjustment[0].Quantity,
				Type:     "adjustment",
				Client:   *message.MessageAttributes["client"].StringValue,
				Source:   sourceStockUpdate.Params.Adjustment.SourceId,
				Date:     sourceStockUpdate.Params.Adjustment.CreatedOn,
				Reason:   "sales",
			}
		default:
			log.Println("Message is not a stock update nor an adjustment")
			continue
		}

		buffer = append(buffer, stockMovement)

		consumer.deleteMessage(svc, message)
	}

	return buffer
}

func (consumer *sqsConsumer) SendReindexRequests(stockMovements []util.StockMovement) {
	svc := consumer.getSqsSession()

	var file = "template/reindex-request.json"
	body, err := ioutil.ReadFile(file)

	if err != nil {
		log.Printf("Cannot read %s /n", file)
	}

	for _, stockMovement := range stockMovements {
		reindexRequest := map[string]interface{}{
			"Source": stockMovement.Source,
			"Sku":    stockMovement.Sku,
			"Client": stockMovement.Client,
		}

		tmpl, err := template.New("reindex-request").Parse(string(body))

		if err != nil {
			panic(err)
		}

		buffer := bytes.NewBuffer(nil)
		err = tmpl.Execute(buffer, reindexRequest)

		params := &sqs.SendMessageInput{
			MessageBody:  aws.String(buffer.String()),
			QueueUrl:     aws.String(consumer.config.Consumer.Queuereindex),
			DelaySeconds: aws.Int64(3),
		}

		_, err = svc.SendMessage(params)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("Reindex Request for source %s, sku %s and client %s sent \n", stockMovement.Source, stockMovement.Sku, stockMovement.Client)
	}
}

func (consumer *sqsConsumer) deleteMessage(svc *sqs.SQS, message *sqs.Message) {
	resultDelete := &sqs.DeleteMessageInput{
		QueueUrl:      &consumer.config.Consumer.Queuestock,
		ReceiptHandle: message.ReceiptHandle,
	}
	_, err := svc.DeleteMessage(resultDelete)

	if err != nil {
		log.Fatal(err)
	}
}
