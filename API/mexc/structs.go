package main

import (
	"net/http"
)

type mexcContext struct {
	Client    *http.Client
	ApiSecret []byte
	ApiKey    string
}

type mexcKline struct {
	OpenTime         uint64
	Open             float64
	High             float64
	Low              float64
	Close            float64
	Volume           float64
	CloseTime        uint64
	QuoteAssetVolume float64
}

type mexcOrder struct {
	Symbol        string `json:"symbol"`
	OrderId       string `json:"orderId"`
	OrderListId   int    `json:"orderListId"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	Type          string `json:"type"`
	Side          string `json:"side"`
	TransactTime  uint64 `json:"transactTime"`
}

type mexcOrderStatus struct {
	Symbol             string `json:"symbol"`
	OrderId            string `json:"orderId"`
	OrderListId        int    `json:"orderListId"`
	ClientOrderId      string `json:"clientOrderId"`
	Price              string `json:"price"`
	OrigQty            string `json:"origQty"`
	ExecutedQty        string `json:"executedQty"`
	CumulativeQuoteQty string `json:"cummulativeQuoteQty"`
	Status             string `json:"status"`
	TimeInForce        string `json:"timeInForce"`
	Type               string `json:"type"`
	Side               string `json:"side"`
	StopPrice          string `json:"stopPrice"`
	Time               uint64 `json:"time"`
	UpdateTime         uint64 `json:"updateTime"`
	IsWorking          bool   `json:"isWorking"`
	OrigQuoteOrderQty  string `json:"origQuoteOrderQty"`
}


type mecxAccountBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type mexcAccount struct {
	MakerCommission    int      `json:"makerCommission"`
	TakerCommission    int      `json:"takerCommission"`
	BuyerCommission    int      `json:"buyerCommission"`
	SellerCommission   int      `json:"sellerCommission"`
	CanTrade           bool     `json:"canTrade"`
	CanWithdraw        bool     `json:"canWithdraw"`
	CanDeposit         bool     `json:"canDeposit"`
	UpdateTime         *int64   `json:"updateTime"`
	AccountType        string   `json:"accountType"`
	Balances           []mecxAccountBalance `json:"balances"`
	Permissions        []string `json:"permissions"`
}

type balance struct {
	Free   float64
	Locked float64
}

type mexcBalance struct {
	Map map[string]balance
}
