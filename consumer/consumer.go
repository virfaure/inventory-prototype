package consumer

import (
	"github.com/magento-mcom/inventory-prototype/util"
)

type Consumer interface {
	PollMessages() []util.StockMovement
}
