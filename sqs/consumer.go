package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/magento-mcom/inventory-prototype/util"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"log"
	"time"
)

var sqsSession *sqs.SQS

const (
	regionAws = "us-west-2"
	profileAws = "training"
	clientMessageAttribute = "client"
)

func getSqsSession() (*sqs.SQS) {
	if sqsSession == nil {
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(regionAws),
			Credentials: credentials.NewSharedCredentials("", profileAws),
		})

		util.Check(err)

		sqsSession = sqs.New(sess)
	}

	return sqsSession
}

func PollOneMessage(queueUrl string) *sqs.Message {
	svc := getSqsSession()

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              &queueUrl,
		MessageAttributeNames: aws.StringSlice([]string{clientMessageAttribute}),
		MaxNumberOfMessages:   aws.Int64(10),
		VisibilityTimeout:     aws.Int64(36000), // 10 hours
		WaitTimeSeconds:       aws.Int64(10),
	})

	util.Check(err)

	if len(result.Messages) == 0 {
		log.Println("No message in the queue")
		return nil
	}

	resultDelete := &sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: result.Messages[0].ReceiptHandle,
	}
	_, err = svc.DeleteMessage(resultDelete)

	util.Check(err)

	return result.Messages[0]
}

func PollMessages(queueUrl string) []*sqs.Message {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		log.Printf("Receiving 10 messages took %v", duration)
	}()

	svc := getSqsSession()

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              &queueUrl,
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

	return result.Messages
}

func DeleteMessage(queueUrl string, message *sqs.Message) {
	svc := getSqsSession()

	resultDelete := &sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: message.ReceiptHandle,
	}
	_, err := svc.DeleteMessage(resultDelete)

	util.Check(err)
}