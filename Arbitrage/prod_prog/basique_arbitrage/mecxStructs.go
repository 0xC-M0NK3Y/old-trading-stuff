package main

const MECX_BASE_URL           = "https://api.mexc.com"
const MECX_FETCH_ORDERBOOK_ED = "/api/v3/depth"
const MECX_FETCH_BALANCE_ED   = "/api/v3/account"
const MECX_FETCH_TIME_ED      = "/api/v3/time"
const MECX_ORDER_ED           = "/api/v3/order"

type mecxFetchOrderbookResp struct {
	LastUpdateId	int `json:"lastUpdateId"`
	Bids			[][]string `json:"bids"`
	Asks			[][]string `json:"asks"`
}

type mecxOrder struct {
	Symbol			string `json:"symbol"`
	Side			string `json:"side"`
	Type			string `json:"type"`
	Quantity		string `json:"quantity"`
	Price			string `json:"price"`
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

type mecxFetchTimeResp struct {
	ServerTime int64 `json:"serverTime"`
}
