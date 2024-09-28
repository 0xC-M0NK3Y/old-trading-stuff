package main

import (
	"fmt"
	"time"
	"bytes"
	"errors"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
)

func mecxFetcher(symbol string, c chan fetchData) {

	// TODO: faire un GET /status avant (et tous les x temps ?)

	dummy_data := fetchData{
		bid:     0,
		bidQnt: 0,
		ask:     999999,
		askQnt: 0}

	client := &http.Client{}
	req, err := http.NewRequest("GET", MECX_BASE_URL+MECX_FETCH_ORDERBOOK_ED+"?symbol="+symbol, nil)
	if err != nil {
		fmt.Printf("Err %s %s\n", MECX_FETCH_ORDERBOOK_ED, err)
		c <- dummy_data
		return
	}
	req.Header.Add("Content-Type", "application/json")

	fmt.Println("Started mecx fetcher ", symbol)

	for {

		time.Sleep(time.Millisecond)

		r, err := client.Do(req)
		if err != nil {
			fmt.Printf("Err %s %s\n", MECX_FETCH_ORDERBOOK_ED, err)
			c <- dummy_data
			continue
		}
		var resp mecxFetchOrderbookResp

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&resp)
		if err != nil {
			fmt.Printf("Err %s %s\n", MECX_FETCH_ORDERBOOK_ED, err)
			c <- dummy_data
			continue
		}
		r.Body.Close()
		var tmp fetchData
		var err1 error
		var err2 error
		var err3 error
		var err4 error

		tmp.bid, err1     = strconv.ParseFloat(resp.Bids[0][0], 64)
		tmp.bidQnt, err2  = strconv.ParseFloat(resp.Bids[0][1], 64)
		tmp.ask, err3     = strconv.ParseFloat(resp.Asks[0][0], 64)
		tmp.askQnt, err4  = strconv.ParseFloat(resp.Asks[0][1], 64)

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			fmt.Printf("Err %s converting floats\n", MECX_FETCH_ORDERBOOK_ED)
			c <- dummy_data
			continue
		}
		c <- tmp
	}
}

func mecxPlaceOrder(t orderData) (string, error) {

	var trade mecxOrder

	if t.side == SIDE_BUY {
		trade.Side = "BUY"
	} else if t.side == SIDE_SELL {
		trade.Side = "SELL"
	} else {
		return "", errors.New(fmt.Sprintf("Err %s orderData error", MECX_ORDER_ED))
	}
	trade.Symbol    = t.symbol
	trade.Type      = "FILL_OR_KILL"
	trade.Price     = strconv.FormatFloat(t.price, 'f', -1, 64)
	trade.Quantity  = strconv.FormatFloat(t.qnt, 'f', -1, 64)
	trade.Timestamp = strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	trade.Signature = hmac256Hex(mecxOrderSignData(trade), []byte(MECX_API_SECRET))

	data, err := json.Marshal(trade)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MECX_ORDER_ED, err))
	}

	fmt.Println(string(data))
	return "", nil

	client := &http.Client{}
	req, err := http.NewRequest("POST", MECX_BASE_URL+MECX_ORDER_ED, bytes.NewBuffer(data))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MECX_ORDER_ED, err))
	}
	mecxAddHeaders(req)

	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200{
		return "", errors.New(fmt.Sprintf("Err %s %d %s", MECX_ORDER_ED, r.StatusCode, err))
	}
	var resp mecxOrderResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MECX_ORDER_ED, err))
	}
	return resp.OrderId, nil
}

func mecxFetchBalance() (map[string]float64, error) {

	ret := make(map[string]float64)

	timestamp := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)

	client := &http.Client{}
	req, err := http.NewRequest("GET", MECX_BASE_URL+MECX_FETCH_BALANCE_ED+"?timestamp="+timestamp+"&signature="+strings.ToLower(hmac256Hex("timestamp="+timestamp, []byte(MECX_API_SECRET))), nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Err %s %s", MECX_FETCH_BALANCE_ED, err))
	}
	mecxAddHeaders(req)
	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200{
		return nil, errors.New(fmt.Sprintf("Err %s %d %s", MECX_FETCH_BALANCE_ED, r.StatusCode, err))
	}
	var resp mecxFetchBalanceResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Err %s %s", MECX_FETCH_BALANCE_ED, err))
	}
	r.Body.Close()

	for i := 0; i < len(resp.Balances); i++ {
		ret[resp.Balances[i].Asset], err = strconv.ParseFloat(resp.Balances[i].Free, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Err %s converting floats", MECX_FETCH_BALANCE_ED))
		}
	}
	return ret, nil
}

func mecxAddHeaders(req *http.Request) {
	req.Header["X-MEXC-APIKEY"] =  []string{MECX_API_KEY}
	req.Header.Add("Content-Type", "application/json")
}

func mecxFetchTime() (string, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", MECX_BASE_URL+MECX_FETCH_TIME_ED, nil)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MECX_FETCH_TIME_ED, err))
	}
	req.Header.Add("Content-Type", "application/json")
	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("Err %s %d %s", MECX_FETCH_TIME_ED, r.StatusCode, err))
	}
	var resp mecxFetchTimeResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MECX_FETCH_TIME_ED, err))
	}
	r.Body.Close()

	return strconv.FormatInt(resp.ServerTime, 10), nil
}

func mecxOrderSignData(t mecxOrder) string {
	return "symbol="+t.Symbol+"&side="+t.Side+"&type="+t.Type+"&quantity="+t.Quantity+"&price="+t.Price+"&timestamp="+t.Timestamp
}
