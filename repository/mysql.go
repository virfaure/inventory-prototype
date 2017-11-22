package repository

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"log"
	"time"
	"strings"
	"github.com/magento-mcom/inventory-prototype/util"
	"github.com/magento-mcom/inventory-prototype/configuration"
)

var instance *sql.DB

func NewMysqlRepository(config configuration.Config) Repository {
	return &mysqlRepository{config}
}

type mysqlRepository struct {
	config configuration.Config
}

func (repository *mysqlRepository) InsertStock(stockMovement []util.StockMovement) {
	if len(stockMovement) == 0 {
		return
	}

	db, err := repository.getConnection()

	util.Check(err)

	start := time.Now()
	defer func() {
		ops := len(stockMovement)
		duration := time.Since(start)
		log.Printf("Insert of %v records took %v (%v per operation)", ops, duration, duration/time.Duration(ops))
	}()

	//query := "INSERT INTO source_stock_history(client, source_id, sku, qty, unlimited_stock, event, created_on) VALUES (?,?,?,?,?,?,?)" + strings.Repeat(", (?,?,?,?,?,?,?)", len(sourceStockHistory)-1)
	//args := make([]interface{}, 0, 6*len(sourceStockHistory))
	query := "INSERT INTO stock_movement(sku, quantity, type, client, source, date, reason) VALUES (?,?,?,?,?, NOW() - INTERVAL FLOOR(RAND() * 14) DAY,?)" + strings.Repeat(", (?,?,?,?,?, NOW() - INTERVAL FLOOR(RAND() * 14) DAY,?)", len(stockMovement)-1)

	//for _, history := range sourceStockHistory {
	//	for _, stock := range history.Stock {
	//		args = append(args, history.Client, history.SourceId, stock.Sku, stock.Quantity, stock.UnlimitedStock, history.Event, history.CreatedOn)
	//	}
	//}

	args := make([]interface{}, 0, 6*len(stockMovement))
	for _, movement := range stockMovement {
		args = append(args, movement.Sku, movement.Quantity, movement.Type, movement.Client, movement.Source, movement.Reason)
	}

	_, err = db.Exec(query, args...)

	if err != nil {
		log.Printf("Stock insert failed: %v", err)
	}
}

func (repository *mysqlRepository) getConnection() (db *sql.DB, err error) {
	if instance == nil {
		x, err := sql.Open(repository.config.Database.Engine, repository.config.Database.DSN)
		if err != nil {
			return nil, err
		}

		instance = x
	}

	return instance, nil
}
