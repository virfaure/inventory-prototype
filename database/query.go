package database

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/magento-mcom/inventory-prototype/util"
	"log"
	"time"
	"strings"
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

	query := "INSERT INTO source_stock_history(client, source_id, sku, qty, unlimited_stock, event, created_on) VALUES (?,?,?,?,?,?,?)" + strings.Repeat(", (?,?,?,?,?,?,?)", len(sourceStockHistory)-1)
	args := make([]interface{}, 0, 6*len(sourceStockHistory))

	for _, history := range sourceStockHistory {
		for _, stock := range history.Stock {
			args = append(args, history.Client, history.SourceId, stock.Sku, stock.Quantity, stock.UnlimitedStock, history.Event, history.CreatedOn)
		}
	}

	_, err = db.Exec(query, args...)

	if err != nil {
		log.Printf("Stock insert failed: %v", err)
	}
}
