package main

import (
	"os"
	"io"
	"fmt"
	"time"
	"bytes"
	"errors"
	"strings"
	"strconv"
	"net/http"
	"encoding/json"
	"encoding/hex"
	"crypto/hmac"
	"crypto/sha256"
)

const DIR_UP   = 0
const DIR_DOWN = 1

func verifConfig() bool {
	if MA1 >= MA2 {
		return false
	}
	// TODO: d'autre verif, flemme
	return true
}

func main() {

	if verifConfig() != true {
		fmt.Println("Error bad config")
		return
	}
/*
	id, err := mexcPlaceOrder("BUY", 0.01, 1623.01)
	if err != nil {
		fmt.Println("err = ", err)
	}
	fmt.Println("id = ", id)

	return
*/

	client := &http.Client{}
	endpoint := MEXC_FETCH_KLINES_ED+"?symbol="+SYMBOL+"&interval="+INTERVAL+"&limit="+strconv.FormatInt(MA2+1, 10)
	req, err := http.NewRequest("GET", MEXC_BASE_URL+endpoint, nil)
	if err != nil {
		fmt.Printf("Err %s\n", err)
		os.Exit(1)
	}
	req.Header.Add("Content-Type", "application/json")

	fmt.Println("Started mecx fetcher")

	var balance map[string]float64

	need_balance := true
	dir          := DIR_UP

	for {
		var ma1 float64
		var ma2 float64

		if need_balance == true {
			balance, err = mexcFetchBalance()
			if err != nil {
				fmt.Printf("Err fetch balance %s\n", err)
				time.Sleep(500*time.Millisecond)
				continue
			}
			fmt.Println("Balance : ", balance)
			need_balance = false
		}

		ma1 = 0
		ma2 = 0
		time.Sleep(500*time.Millisecond)
		r, err := client.Do(req)
		if err != nil {
			fmt.Printf("Err %s\n", err)
			continue
		}
		var data [][]interface{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Err %s\n", err)
			continue
		}
		if err := json.Unmarshal(body, &data); err != nil {
			fmt.Printf("Err %s\n", err)
			continue
		}
		n1 := 0
		n2 := 0
		for i := len(data)-1; i >= 0; i-- {
			if len(data[i]) != 8 {
				fmt.Printf("Err received something weird ! EXIT")
				os.Exit(1)
			}
			if n1 < MA1 {
				tmp, err := strconv.ParseFloat(data[i][4].(string), 64)
				if err != nil {
					fmt.Printf("Err converting floats ! EXIT")
					os.Exit(1)
				}
				ma1 += tmp
				n1 += 1
			}
			if n2 < MA2 {
				tmp, err := strconv.ParseFloat(data[i][4].(string), 64)
				if err != nil {
					fmt.Printf("Err converting floats ! EXIT")
					os.Exit(1)
				}
				ma2 += tmp
				n2 += 1
			}
		}
		if n1 != MA1 || n2 != MA2 {
			fmt.Printf("Err in computing MA ! EXIT\n")
			os.Exit(1)
		}
		ma1 /= float64(MA1)
		ma2 /= float64(MA2)

		if dir == DIR_UP && ma1 < ma2 {
			fmt.Printf("Cross sell\n")

			book, err := mexcFetchOrderbook()
			if err != nil {
				fmt.Printf("Err fetching orderbook")
				continue
			}
			book.bid -= 0.01
			fmt.Println("book = ", book)
			//TODO: "ETH" et "USDT" en fonction de SYMBOL
			id, err := mexcPlaceOrder("SELL", balance["ETH"], book.bid)
			if err != nil {
				fmt.Printf("Err place order %s\n", err)
				continue
			}
			dir = DIR_DOWN
			need_balance = true
			fmt.Println("Success sell ", id)
		} else if dir == DIR_DOWN && ma1 > ma2 {
			fmt.Printf("Cross buy\n")

			book, err := mexcFetchOrderbook()
			if err != nil {
				fmt.Printf("Err fetching orderbook")
				continue
			}
			book.ask += 0.01
			fmt.Println("book = ", book)
			id, err := mexcPlaceOrder("BUY", balance["USDC"]/book.ask, book.ask)
			if err != nil {
				fmt.Printf("Err place order %s\n", err)
				continue
			}
			dir = DIR_UP
			need_balance = true
			fmt.Println("Success buy ", id)
		}
	}
}

func mexcPlaceOrder(side string, qnt float64, price float64) (string, error) {

	price_str := strconv.FormatFloat(price, 'f', -1, 64)
	qnt_str   := strconv.FormatFloat(qnt, 'f', -1, 64)
	timestamp := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)

	body := "symbol="+SYMBOL+"&side="+side+"&type=IMMEDIATE_OR_CANCEL&quantity="+qnt_str+"&price="+price_str+"&timestamp="+timestamp
	sign := hmac256Hex(body, []byte(MEXC_API_SECRET))
	body += "&signature="+sign

	client := &http.Client{}
	req, err := http.NewRequest("POST", MEXC_BASE_URL+MEXC_ORDER_ED, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MEXC_ORDER_ED, err))
	}
	mexcAddHeaders(req)

	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200 {
		tmp, _ := io.ReadAll(r.Body)
		return "", errors.New(fmt.Sprintf("Err %s %d %s %s", MEXC_ORDER_ED, r.StatusCode, string(tmp), err))
	}
	var resp mecxOrderResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MEXC_ORDER_ED, err))
	}

	fmt.Println("resp = ", resp)

	return resp.OrderId, nil
}

func mexcFetchBalance() (map[string]float64, error) {

	ret := make(map[string]float64)

	timestamp := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)

	client := &http.Client{}
	req, err := http.NewRequest("GET", MEXC_BASE_URL+MEXC_FETCH_BALANCE_ED+"?timestamp="+timestamp+"&signature="+strings.ToLower(hmac256Hex("timestamp="+timestamp, []byte(MEXC_API_SECRET))), nil)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Err %s %s", MEXC_FETCH_BALANCE_ED, err))
	}
	mexcAddHeaders(req)
	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Err %s %d %s", MEXC_FETCH_BALANCE_ED, r.StatusCode, err))
	}
	var resp mecxFetchBalanceResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Err %s %s", MEXC_FETCH_BALANCE_ED, err))
	}
	r.Body.Close()

	for i := 0; i < len(resp.Balances); i++ {
		ret[resp.Balances[i].Asset], err = strconv.ParseFloat(resp.Balances[i].Free, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Err %s converting floats", MEXC_FETCH_BALANCE_ED))
		}
	}
	return ret, nil
}

func mexcFetchOrderbook() (mexcBook, error) {

	var ret mexcBook

	client := &http.Client{}
	req, err := http.NewRequest("GET", MEXC_BASE_URL+MEXC_FETCH_ORDERBOOK_ED+"?symbol="+SYMBOL, nil)
	if err != nil {
		return ret, errors.New(fmt.Sprintf("Err %s %s\n", MEXC_FETCH_ORDERBOOK_ED, err))
	}
	req.Header.Add("Content-Type", "application/json")

	r, err := client.Do(req)
	if err != nil {
		return ret, errors.New(fmt.Sprintf("Err %s %s\n", MEXC_FETCH_ORDERBOOK_ED, err))
	}
	var resp mexcFetchOrderbookResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil {
		return ret, errors.New(fmt.Sprintf("Err %s %s\n", MEXC_FETCH_ORDERBOOK_ED, err))
	}
	r.Body.Close()

	var err1 error
	var err2 error
	var err3 error
	var err4 error

	ret.bid, err1     = strconv.ParseFloat(resp.Bids[0][0], 64)
	ret.bidQnt, err2  = strconv.ParseFloat(resp.Bids[0][1], 64)
	ret.ask, err3     = strconv.ParseFloat(resp.Asks[0][0], 64)
	ret.askQnt, err4  = strconv.ParseFloat(resp.Asks[0][1], 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return ret, errors.New(fmt.Sprintf("Err %s converting floats\n", MEXC_FETCH_ORDERBOOK_ED))
	}
	return ret, nil
}

func mexcAddHeaders(req *http.Request) {
	req.Header["X-MEXC-APIKEY"] =  []string{MEXC_API_KEY}
	req.Header.Add("Content-Type", "application/json")
}

func hmac256Hex(str string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(str))
	sign := mac.Sum(nil)
	ret := hex.EncodeToString(sign)
	return ret
}
