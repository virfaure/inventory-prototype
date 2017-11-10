package util

type SourceStockHistory struct {
	Client string
	SourceId string
	Event string
	CreatedOn string
	Stock []Stock
}

type StockHistory struct {
	Qty int
	Sku string
	UnlimitedStock int
}

type SourceStockUpdateJsonRpc struct {
	Params *SourceStockUpdate `json:"params"`
}

type SourceStockUpdate struct {
	Snapshot   *SourceStockSnapshot `json:"snapshot"`
	Adjustment *SourceStockSnapshot `json:"adjustment"`
}

type SourceStockSnapshot struct {
	CreatedOn string `json:"created_on"`
	Mode string `json:"mode"`
	SourceId string `json:"source_id"`
	Stock []Stock `json:"stock"`
	StockAdjustment []Stock `json:"adjustments"`
}

type Stock struct {
	Quantity int `json:"quantity"`
	Sku string `json:"sku"`
	UnlimitedStock int `json:"unlimited_stock"`
}
