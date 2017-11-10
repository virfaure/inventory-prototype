package main

import (
	"github.com/magento-mcom/inventory-prototype/util"
	"github.com/magento-mcom/inventory-prototype/sqs"
	"encoding/json"
	"log"
	"sync"
	"time"
	"os"
	"os/signal"
	"github.com/magento-mcom/inventory-prototype/database"
)

const (
	amazonQueueUrl = "https://sqs.us-west-2.amazonaws.com/277100466574/prototype2-inventory-api-gateway-queue"
)

func main() {
	updates := make(chan util.SourceStockHistory, 1000)
	done := make(chan struct{})

	// Consumers
	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			poll(done, updates)
		}()
	}

	// Inserters
	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			inserter(updates)
		}()
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	<-signals

	close(done)
	close(updates)
	wg.Wait()
}

func poll(done chan struct{}, updates chan util.SourceStockHistory) {
	for {
		select {
		case <-done:
			return
		default:
		}

		messages := sqs.PollMessages(amazonQueueUrl)

		for _, message := range messages {

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

			//log.Printf("Processing message number %#v \n", message)
			var sourceStockHistory util.SourceStockHistory

			switch {
			case sourceStockUpdate.Params.Snapshot != nil:
				sourceStockHistory = util.SourceStockHistory{
					Client:    message.MessageAttributes["client"].String(),
					SourceId:  sourceStockUpdate.Params.Snapshot.SourceId,
					Event:     "snapshot",
					CreatedOn: sourceStockUpdate.Params.Snapshot.CreatedOn,
					Stock:     sourceStockUpdate.Params.Snapshot.Stock,
				}
			case sourceStockUpdate.Params.Adjustment != nil:
				sourceStockHistory = util.SourceStockHistory{
					Client:    message.MessageAttributes["client"].String(),
					SourceId:  sourceStockUpdate.Params.Adjustment.SourceId,
					Event:     "adjustment",
					CreatedOn: sourceStockUpdate.Params.Adjustment.CreatedOn,
					Stock:     sourceStockUpdate.Params.Adjustment.StockAdjustment,
				}
			default:
				log.Println("Message is not a stock update nor an adjustment")
				continue
			}

			select {
			case <-done:
				return
			default:
			}

			updates <- sourceStockHistory
			sqs.DeleteMessage(amazonQueueUrl, message)
		}
	}
}

func inserter(updates chan util.SourceStockHistory) {

	var buffer []util.SourceStockHistory

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				return
			}

			buffer = append(buffer, update)

			if len(buffer) > 10 {
				database.InsertStock(buffer)
				buffer = nil
			}
		case <-time.After(time.Second):
			database.InsertStock(buffer)
			buffer = nil
		}

	}

	database.InsertStock(buffer)
	buffer = nil
}
