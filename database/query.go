package database

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/magento-mcom/inventory-prototype/util"
	"log"
	"time"
)

func InsertStock(sourceStockHistory []util.SourceStockHistory) {
	if len(sourceStockHistory) == 0 {
		return
	}

	db, err := GetConnection()

	util.Check(err)

	start := time.Now()
	defer func() {
		ops := len(sourceStockHistory)
		duration := time.Since(start)
		log.Printf("Insert of %v records took %v (%v per operation)", ops, duration, duration/time.Duration(ops))
	}()

	for _, history := range sourceStockHistory {
		for _, stock := range history.Stock {
			_, err = db.Exec(
				"INSERT INTO source_stock_history(client, source_id, sku, qty, unlimited_stock, event, created_on) VALUES (?,?,?,?,?,?,?)",
				history.Client,
				history.SourceId,
				stock.Sku,
				stock.Quantity,
				stock.UnlimitedStock,
				history.Event,
				history.CreatedOn,
			)
			//log.Printf("Inserted into source_stock_history, %#v \n", history)

			util.Check(err)
		}
	}
}
