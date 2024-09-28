package main

const MEXC_BASE_URL           = "https://api.mexc.com"
const MEXC_FETCH_KLINES_ED    = "/api/v3/klines"
const MEXC_FETCH_ORDERBOOK_ED = "/api/v3/depth"
const MEXC_FETCH_BALANCE_ED   = "/api/v3/account"
const MEXC_ORDER_ED           = "/api/v3/order" // POST
const MEXC_FETCH_ORDER_ED     = "/api/v3/order" // GET

type mexcBook struct {
	bid		float64
	bidQnt	float64
	ask		float64
	askQnt	float64
}

type mecxOrder struct {
	Symbol			string `json:"symbol"`
	Side			string `json:"side"`
	Type			string `json:"type"`
	Quantity		string `json:"quantity"`
	Timestamp		string `json:"timestamp"`
	Signature		string `json:"signature"`
}

type mecxOrderResp struct {
	Symbol        string `json:"symbol"`
	OrderId       string `json:"orderId"`
	OrderListId   int64  `json:"orderListId"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	Type          string `json:"type"`
	Side          string `json:"side"`
	TransactTime  int64  `json:"transactTime"`
}

type mecxFetchBalanceData struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type mecxFetchBalanceResp struct {
	MakerCommission    int      `json:"makerCommission"`
	TakerCommission    int      `json:"takerCommission"`
	BuyerCommission    int      `json:"buyerCommission"`
	SellerCommission   int      `json:"sellerCommission"`
	CanTrade           bool     `json:"canTrade"`
	CanWithdraw        bool     `json:"canWithdraw"`
	CanDeposit         bool     `json:"canDeposit"`
	UpdateTime         *int64   `json:"updateTime"`
	AccountType        string   `json:"accountType"`
	Balances           []mecxFetchBalanceData `json:"balances"`
	Permissions        []string `json:"permissions"`
}

type mexcFetchOrderbookResp struct {
	LastUpdateId	int `json:"lastUpdateId"`
	Bids			[][]string `json:"bids"`
	Asks			[][]string `json:"asks"`
}

type mexcFetchOrderResp struct {
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
	Time               int64  `json:"time"`
	UpdateTime         int64  `json:"updateTime"`
	IsWorking          bool   `json:"isWorking"`
	OrigQuoteOrderQty  string `json:"origQuoteOrderQty"`
}
