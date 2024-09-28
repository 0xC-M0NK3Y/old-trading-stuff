package main

/* Endpoints */

const OKX_BASE_URL           = "https://www.okx.com"
const OKX_FETCH_ORDERBOOK_ED = "/api/v5/market/books"
const OKX_ORDER_ED           = "/api/v5/trade/order" // POST
const OKX_FETCH_BALANCE_ED   = "/api/v5/account/balance"
const OKX_FETCH_ORDER_ED     = "/api/v5/trade/order" // GET
const OKX_CANCEL_ORDER_ED    = "/api/v5/trade/cancel-order"

/* Structs */

//////////////////////////////////////////////

type okxFetchOrderbookResp struct {
	Code	string `json:"code"`
	Msg		string `json:"msg"`
	Data	[]struct {
		Asks [][]string `json:"asks"`
		Bids [][]string `json:"bids"`
		Ts string `json:"ts"`
	} `json:"data"`
}

//////////////////////////////////////////////

type okxOrderBody struct {
	InstId	string `json:"instId"`
	TdMode	string `json:"tdMode"`
	Side	string `json:"side"`
	OrdType	string `json:"ordType"`
	Px		string `json:"px"`
	Sz		string `json:"sz"`
	TgtCcy	string `json:"tgtCcy"`
}
type okxOrderResp struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	Data    []struct {
		ClOrdId string `json:"clOrdId"`
		OrdId   string `json:"ordId"`
		Tag     string `json:"tag"`
		SCode   string `json:"sCode"`
		SMsg    string `json:"sMsg"`
	} `json:"data"`
	InTime  string `json:"inTime"`
	OutTime string `json:"outTime"`
}

///////////////////////////////////////////////

type okxFetchBalanceRespDetails struct {
	AvailBal       string `json:"availBal"`
	AvailEq        string `json:"availEq"`
	CashBal        string `json:"cashBal"`
	Ccy            string `json:"ccy"`
	CrossLiab      string `json:"crossLiab"`
	DisEq          string `json:"disEq"`
	Eq             string `json:"eq"`
	EqUsd          string `json:"eqUsd"`
	FixedBal       string `json:"fixedBal"`
	FrozenBal      string `json:"frozenBal"`
	Interest       string `json:"interest"`
	IsoEq          string `json:"isoEq"`
	IsoLiab        string `json:"isoLiab"`
	IsoUpl         string `json:"isoUpl"`
	Liab           string `json:"liab"`
	MaxLoan        string `json:"maxLoan"`
	MgnRatio       string `json:"mgnRatio"`
	NotionalLever  string `json:"notionalLever"`
	OrdFrozen      string `json:"ordFrozen"`
	Twap           string `json:"twap"`
	UTime          string `json:"uTime"`
	Upl            string `json:"upl"`
	UplLiab        string `json:"uplLiab"`
	StgyEq         string `json:"stgyEq"`
	SpotInUseAmt   string `json:"spotInUseAmt"`
	BorrowFroz     string `json:"borrowFroz"`
}
type okxFetchBalanceRespData struct {
    AdjEq        string    `json:"adjEq"`
    BorrowFroz   string    `json:"borrowFroz"`
    Details      []okxFetchBalanceRespDetails `json:"details"`
    Imr          string    `json:"imr"`
    IsoEq        string    `json:"isoEq"`
    MgnRatio     string    `json:"mgnRatio"`
    Mmr          string    `json:"mmr"`
    NotionalUsd  string    `json:"notionalUsd"`
    OrdFroz      string    `json:"ordFroz"`
    TotalEq      string    `json:"totalEq"`
    UTime        string    `json:"uTime"`
}
type okxFetchBalanceResp struct {
    Code string `json:"code"`
    Data []okxFetchBalanceRespData `json:"data"`
    Msg  string `json:"msg"`
}

////////////////////////////////////////////////////////

type okxFetchOrderRespData struct {
	InstType          string `json:"instType"`
	InstID            string `json:"instId"`
	Ccy               string `json:"ccy"`
	OrdID             string `json:"ordId"`
	ClOrdID           string `json:"clOrdId"`
	Tag               string `json:"tag"`
	Px                string `json:"px"`
	PxUsd             string `json:"pxUsd"`
	PxVol             string `json:"pxVol"`
	PxType            string `json:"pxType"`
	Sz                string `json:"sz"`
	Pnl               string `json:"pnl"`
	OrdType           string `json:"ordType"`
	Side              string `json:"side"`
	PosSide           string `json:"posSide"`
	TDMode            string `json:"tdMode"`
	AccFillSz         string `json:"accFillSz"`
	FillPx            string `json:"fillPx"`
	TradeID           string `json:"tradeId"`
	FillSz            string `json:"fillSz"`
	FillTime          string `json:"fillTime"`
	State             string `json:"state"`
	AvgPx             string `json:"avgPx"`
	Lever             string `json:"lever"`
	AttachAlgoClOrdID string `json:"attachAlgoClOrdId"`
	TpTriggerPx       string `json:"tpTriggerPx"`
	TpTriggerPxType   string `json:"tpTriggerPxType"`
	TpOrdPx           string `json:"tpOrdPx"`
	SlTriggerPx       string `json:"slTriggerPx"`
	SlTriggerPxType   string `json:"slTriggerPxType"`
	SlOrdPx           string `json:"slOrdPx"`
	StpID             string `json:"stpId"`
	StpMode           string `json:"stpMode"`
	FeeCcy            string `json:"feeCcy"`
	Fee               string `json:"fee"`
	RebateCcy         string `json:"rebateCcy"`
	Rebate            string `json:"rebate"`
	TgtCcy            string `json:"tgtCcy"`
	Category          string `json:"category"`
	ReduceOnly        string `json:"reduceOnly"`
	CancelSource      string `json:"cancelSource"`
	CancelSourceReason string `json:"cancelSourceReason"`
	QuickMgnType      string `json:"quickMgnType"`
	AlgoClOrdID       string `json:"algoClOrdId"`
	AlgoID            string `json:"algoId"`
	UTime             string `json:"uTime"`
	CTime             string `json:"cTime"`
}
type okxFetchOrderResp struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data []okxFetchOrderRespData `json:"data"`
}

////////////////////////////////////////////////////////

type okxCancelOrderBody struct {
	OrdId  string `json:"ordId"`
	InstId string `json:"instId"`
}
type okxCancelOrderRespData struct {
	ClOrdID string `json:"clOrdId"`
	OrdID   string `json:"ordId"`
	SCode   string `json:"sCode"`
	SMsg    string `json:"sMsg"`
}
type okxCancelOrderResp struct {
	Code    string          `json:"code"`
	Msg     string          `json:"msg"`
	Data    []okxCancelOrderRespData     `json:"data"`
	InTime  string          `json:"inTime"`
	OutTime string          `json:"outTime"`
}

//////////////////////////////////////////////////////////
