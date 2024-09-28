package main

import (
	_ "fmt"
	"io"
	"time"
	"errors"
	"strconv"
	"net/http"
	"encoding/json"
)

func mexcInit(
	apiKey string,
	apiSecret string) (mexcContext, error) {
	var ret mexcContext

	// TODO: verify if api key and secret is working ?
	
	ret.Client    = &http.Client{}
	ret.ApiKey    = apiKey
	ret.ApiSecret = []byte(apiSecret)
	
	return ret, nil
}

func mexcFetchKlines(
	ctx mexcContext,
	symbol string,
	interval string,
	limit int64) ([]mexcKline, error) {

	var data [][]interface{}
	var ret []mexcKline
	
	body := "symbol="+symbol
	body += "&interval="+interval
	body += "&limit="+strconv.FormatInt(limit, 10)

	r, err := makeRequest(ctx, "GET", "/api/v3/klines?"+body, nil, false)
	if err != nil {
		return nil, err
	}
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(data); i++ {
		var tmp mexcKline
		var e1,e2,e3,e4,e5,e6,e7,e8 error
		
		tmp.OpenTime             = uint64(data[i][0].(float64))
		tmp.Open, e2             = strconv.ParseFloat(data[i][1].(string), 64)
		tmp.High, e3             = strconv.ParseFloat(data[i][2].(string), 64)
		tmp.Low, e4              = strconv.ParseFloat(data[i][3].(string), 64)
		tmp.Close, e5            = strconv.ParseFloat(data[i][4].(string), 64)
		tmp.Volume, e6           = strconv.ParseFloat(data[i][5].(string), 64)
		tmp.CloseTime            = uint64(data[i][6].(float64))
		tmp.QuoteAssetVolume, e8 = strconv.ParseFloat(data[i][7].(string), 64)

		if (e1 != nil || e2 != nil || e3 != nil || e4 != nil ||
			e5 != nil || e6 != nil || e7 != nil || e8 != nil) {
			return nil, errors.New("erreur dans le parsing")
		}
		ret = append(ret, tmp)
	}
	return ret, nil
}


func mexcNewOrder(
	ctx mexcContext,
	symbol string,
	side string,
	typee string,
	quantity float64,
	quoteOrderQty float64,
	price float64) (mexcOrder, error) {

	var ret mexcOrder
	
	if (quantity != 0 && quoteOrderQty != 0) || (quantity == 0 && quoteOrderQty == 0) {
		return ret, errors.New("Need only quantity or quoteOrderQty")
	}
	if typee == MEXC_ORDER_TYPE_MARKET && price != 0 {
		return ret, errors.New("Can't have price with market order")
	}
	if typee == MEXC_ORDER_TYPE_LIMIT && price == 0 {
		return ret, errors.New("Need price with limit order")
	}
	if typee == MEXC_ORDER_TYPE_FILL_OR_KILL && price == 0 {
		return ret, errors.New("Need price with fill or kill order")
	}
	if typee == MEXC_ORDER_TYPE_FILL_OR_KILL && (quantity == 0 || quoteOrderQty != 0) {
		return ret, errors.New("Need quantity with fill or kill order")
	}
	if typee == MEXC_ORDER_TYPE_MARKET && side == MEXC_SIDE_BUY && quoteOrderQty == 0 {
		return ret, errors.New("Need quoteOrderQty for market side buy order")
	} else if typee == MEXC_ORDER_TYPE_MARKET && side == MEXC_SIDE_SELL && quantity == 0 {
		return ret, errors.New("Need quantity for market side sell order")
	}
	body := "symbol="+symbol
	body += "&side="+side
	body += "&type="+typee
	if quantity != 0 {
		body += "&quantity="+strconv.FormatFloat(quantity, 'f', -1, 64)
	} else {
		body += "&quoteOrderQty="+strconv.FormatFloat(quoteOrderQty, 'f', -1, 64)
	}
	if price != 0 {
		body += "&price="+strconv.FormatFloat(price, 'f', -1, 64)
	}
	body += "&timestamp="+strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	body += "&signature="+getSignature(ctx, body)
	
	r, err := makeRequest(ctx, "POST", "/api/v3/order", []byte(body), true)
	if err != nil {
		return ret, err
	}
	err = json.NewDecoder(r.Body).Decode(&ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func mexcFetchAccount(
	ctx mexcContext) (mexcAccount ,error) {
	var ret mexcAccount

	body      := "timestamp="+strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	body      += "&signature="+getSignature(ctx, body)

	r, err := makeRequest(ctx, "GET", "/api/v3/account?"+body, nil, true)
	if err != nil {
		return ret, err
	}
	err = json.NewDecoder(r.Body).Decode(&ret)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func mexcFetchBalance(
	ctx mexcContext) (mexcBalance, error) {
	var ret mexcBalance

	ret.Map = make(map[string]balance)
	
	account, err := mexcFetchAccount(ctx)
	if err != nil {
		return ret, err
	}
	for i := 0; i < len(account.Balances); i++ {
		var tmp balance

		tmp.Free, err = strconv.ParseFloat(account.Balances[i].Free, 64)
		if err != nil {
			return ret, err
		}
		tmp.Locked, err = strconv.ParseFloat(account.Balances[i].Locked, 64)
		if err != nil {
			return ret, err
		}
		ret.Map[account.Balances[i].Asset] = tmp
	}
	return ret, nil
}

func mexcQueryOrder(
	ctx mexcContext,
	order mexcOrder) (mexcOrderStatus, error) {
	var ret mexcOrderStatus

	body := "symbol="+order.Symbol
	body += "&orderId="+order.OrderId
	body += "&timestamp="+strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	body += "&signature="+getSignature(ctx, body)

	r, err := makeRequest(ctx, "GET", "/api/v3/order?"+body, nil, true)
	if err != nil {
		return ret, err
	}
	err = json.NewDecoder(r.Body).Decode(&ret)
	if err != nil {
		return ret, err
	}
	return ret, nil	
}
