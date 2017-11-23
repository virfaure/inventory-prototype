package consumer

import (
	"github.com/magento-mcom/inventory-prototype/util"
)

type Consumer interface {
	PollStockMessages() []util.StockMovement
	SendReindexRequests([]util.StockMovement)
}
