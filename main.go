package main

import (
	"flag"
	"fmt"
	"github.com/magento-mcom/inventory-prototype/app"
	"github.com/magento-mcom/inventory-prototype/configuration"
	"github.com/magento-mcom/inventory-prototype/util"
	"sync"
)

func main() {
	filename := flag.String("config", "config.yml", "Configuration file")
	flag.Parse()

	config, err := configuration.Load(*filename)

	if err != nil {
		panic(fmt.Errorf("Failed to load configuration: %v", err))
	}

	l := app.NewLoader(config)
	var messages []util.StockMovement

	for {
		// Consumers
		wg := sync.WaitGroup{}
		for i := 0; i < 20; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				messages = l.Consumer().PollMessages()
			}()
		}

		// Inserters
		for i := 0; i < 20; i++ {
			wg.Add(1)

			go func() {
				defer wg.Done()
				l.Repository().InsertStock(messages)
			}()
		}

		wg.Wait()
	}
}
