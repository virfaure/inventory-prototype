package repository

import "github.com/magento-mcom/inventory-prototype/util"

type Repository interface {
	InsertStock(stockMovement []util.StockMovement)
}