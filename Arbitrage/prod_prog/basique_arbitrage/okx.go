package main

import (
	"fmt"
	"time"
	"bytes"
	"errors"
	"strings"
	"strconv"
	"encoding/json"
	"net/http"
)

func okxFetcher(symbol string, c chan fetchData) {

	// TODO: faire un GET /status avant (et tous les x temps ?)

	dummy_data := fetchData{
		bid:     0,
		bidQnt: 0,
		ask:     999999,
		askQnt: 0}

	client := &http.Client{}
	req, err := http.NewRequest("GET", OKX_BASE_URL+OKX_FETCH_ORDERBOOK_ED+"?sz=1&instId="+symbol, nil)
	if err != nil {
		fmt.Printf("Err creating fetch request okx %s\n", err)
		c <- dummy_data
		return
	}
	req.Header.Add("Content-Type", "application/json")

	fmt.Println("Started okx fetcher ", symbol)

	for {
		time.Sleep(time.Millisecond)

		r, err := client.Do(req)
		if err != nil || r.StatusCode != 200 {
			fmt.Printf("Err %s %d %s\n", OKX_FETCH_ORDERBOOK_ED, r.StatusCode, err)
			c <- dummy_data
			continue
		}
		var resp okxFetchOrderbookResp

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&resp)
		if err != nil || resp.Code != "0" {
			fmt.Printf("Err %s %s %s %s\n", OKX_FETCH_ORDERBOOK_ED, resp.Code, resp.Msg, err)
			c <- dummy_data
			continue
		}
		var tmp fetchData
		var err1 error
		var err2 error
		var err3 error
		var err4 error

		tmp.bid, err1     = strconv.ParseFloat(resp.Data[0].Bids[0][0], 64)
		tmp.bidQnt, err2  = strconv.ParseFloat(resp.Data[0].Bids[0][1], 64)
		tmp.ask, err3     = strconv.ParseFloat(resp.Data[0].Asks[0][0], 64)
		tmp.askQnt, err4  = strconv.ParseFloat(resp.Data[0].Asks[0][1], 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			fmt.Printf("Err %s converting floats\n", OKX_FETCH_ORDERBOOK_ED)
			c <- dummy_data
			continue
		}
		c <- tmp
	}
}

func okxPlaceOrder(t orderData) (string, error) {

	var order okxOrderBody

	if t.side == SIDE_BUY {
		order.Side = "buy"
	} else if t.side == SIDE_SELL {
		order.Side = "sell"
	} else {
		return "", errors.New(fmt.Sprintf("Err %s orderData error", OKX_ORDER_ED))
	}
	order.InstId  = t.symbol
	order.TdMode  = "cash"
	order.OrdType = "limit" // "post_only"
	order.Sz      = strconv.FormatFloat(t.qnt, 'f', -1, 64)
	order.Px      = strconv.FormatFloat(t.price, 'f', -1, 64)
	order.TgtCcy  = "base_ccy"

	data, err := json.Marshal(order)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", OKX_ORDER_ED, err))
	}

	client := http.Client{}
	req, err := http.NewRequest("POST", OKX_BASE_URL+OKX_ORDER_ED, bytes.NewBuffer(data))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", OKX_ORDER_ED, err))
	}
	okxAddHeaders(req, "POST", OKX_ORDER_ED, string(data))

	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Err %s %d %s", OKX_ORDER_ED, r.StatusCode, err))
	}
	var resp okxOrderResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil || resp.Code != "0" {
		return "", errors.New(fmt.Sprintf("Err %s %s %s %s", OKX_ORDER_ED, resp.Code, resp.Msg, err))
	}
	r.Body.Close()

	return resp.Data[0].OrdId, nil
}

func okxFetchBalance(symbol string) (map[string]float64, error) {

	ret := make(map[string]float64)
	client := http.Client{}

	symbol = strings.Replace(symbol, "-", ",", 1)

	req, err := http.NewRequest("GET", OKX_BASE_URL+OKX_FETCH_BALANCE_ED+"?ccy="+symbol, nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Err %s %s", OKX_FETCH_BALANCE_ED, err))
	}
	okxAddHeaders(req, "GET", OKX_FETCH_BALANCE_ED+"?ccy="+symbol, "")

	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Err %s %d %s", OKX_FETCH_BALANCE_ED, r.StatusCode, err))
	}
	var resp okxFetchBalanceResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil || resp.Code != "0" {
		return nil, errors.New(fmt.Sprintf("Err %s %s %s %s", OKX_FETCH_BALANCE_ED, resp.Code, resp.Msg, err))
	}
	r.Body.Close()

	for i := 0; i < len(resp.Data[0].Details); i++ {
		ret[resp.Data[0].Details[i].Ccy], err = strconv.ParseFloat(resp.Data[0].Details[i].AvailBal, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Err %s converting floats", OKX_FETCH_BALANCE_ED))
		}
	}
	return ret, nil
}

func okxFetchOrder(order_id string, symbol string) (orderStatus, error) {

	var ret orderStatus

	client := http.Client{}

	req, err := http.NewRequest("GET", OKX_BASE_URL+OKX_FETCH_ORDER_ED+"?ordId="+order_id+"&instId="+symbol, nil)
	if err != nil {
		return ret, errors.New(fmt.Sprintf("Err %s %s", OKX_FETCH_ORDER_ED, err))
	}
	okxAddHeaders(req, "GET", OKX_FETCH_ORDER_ED+"?ordId="+order_id+"&instId="+symbol, "")

	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200 {
		return ret, errors.New(fmt.Sprintf("Err %s %d %s", OKX_FETCH_ORDER_ED, r.StatusCode, err))
	}
	var resp okxFetchOrderResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil || resp.Code != "0" {
		return ret, errors.New(fmt.Sprintf("Err %s %s %s %s", OKX_FETCH_ORDER_ED, resp.Code, resp.Msg, err))
	}
	r.Body.Close()

	ret.filled, err = strconv.ParseFloat(resp.Data[0].FillSz, 64)
	ret.state       = resp.Data[0].State
	if err != nil {
		return ret, errors.New(fmt.Sprintf("Err %s converting floats", OKX_FETCH_ORDER_ED))
	}
	return ret, nil
}


func okxCancelOrder(order_id string, symbol string) error {

	var cancel okxCancelOrderBody

	cancel.OrdId  = order_id
	cancel.InstId = symbol

	data, err := json.Marshal(cancel)
	if err != nil {
		return errors.New(fmt.Sprintf("Err %s %s", OKX_CANCEL_ORDER_ED, err))
	}

	client := http.Client{}
	req, err := http.NewRequest("POST", OKX_BASE_URL+OKX_CANCEL_ORDER_ED, bytes.NewBuffer(data))
	if err != nil {
		return errors.New(fmt.Sprintf("Err %s %s", OKX_ORDER_ED, err))
	}
	okxAddHeaders(req, "POST", OKX_CANCEL_ORDER_ED, string(data))

	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Err %s %d %s", OKX_ORDER_ED, r.StatusCode, err))
	}
	var resp okxCancelOrderResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil || resp.Code != "0" {
		return errors.New(fmt.Sprintf("Err %s %s %s %s", OKX_ORDER_ED, resp.Code, resp.Msg, err))
	}
	r.Body.Close()

	return nil
}

func okxAddHeaders(req *http.Request, method string, endpoint string, body string) {
	req.Header.Add("OK-ACCESS-KEY", OKX_API_KEY)
	req.Header.Add("OK-ACCESS-PASSPHRASE", OKX_API_PASSPHRASE)
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	req.Header.Add("OK-ACCESS-SIGN", hmac256Base64(timestamp+method+endpoint+body, []byte(OKX_API_SECRET)))
	req.Header.Add("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Add("Content-Type", "application/json")
}
