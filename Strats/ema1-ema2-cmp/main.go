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

	"log"
	"net/http/httputil"
)

const DIR_UP   = 0
const DIR_DOWN = 1

func verifConfig() bool {
	if EMA1 >= EMA2 {
		return false
	}
	// TODO: d'autre verif, flemme
	// ema max, symbol tradable ?

	return true
}

func main() {

	status, err := mexcFetchOrder("C02__396046499349299200029")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(status)

	return
	
	if verifConfig() != true {
		fmt.Println("Error bad config")
		return
	}

	client := &http.Client{}
	endpoint := MEXC_FETCH_KLINES_ED+"?symbol="+SYMBOL+"&interval="+INTERVAL+"&limit="+strconv.FormatInt(EMA2+1, 10)
	req, err := http.NewRequest("GET", MEXC_BASE_URL+endpoint, nil)
	if err != nil {
		fmt.Printf("Err %s\n", err)
		os.Exit(1)
	}
	req.Header.Add("Content-Type", "application/json")

	fmt.Printf("EMA%d | EMA%d | %s| %s\n\n", EMA1, EMA2, INTERVAL, SYMBOL)

	var balance map[string]float64

	need_balance := true
	dir          := DIR_DOWN
	block        := 0
	start        := true
	tmp_start    := 0

	for {
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
		if len(data) < EMA2 {
			fmt.Printf("Err received not enought klines ! EXIT\n")
			os.Exit(1)
		}
		n1     := 1
		n2     := 1
		ema1, err := strconv.ParseFloat(data[len(data)-1][4].(string), 64)
		if err != nil {
			fmt.Printf("Err converting floats ! EXIT")
			os.Exit(1)
		}
		ema2, err := strconv.ParseFloat(data[len(data)-1][4].(string), 64)
		if err != nil {
			fmt.Printf("Err converting floats ! EXIT")
			os.Exit(1)
		}
		alpha1 := 2.0 / (float64(EMA1) + 1.0)
		alpha2 := 2.0 / (float64(EMA2) + 1.0)
		for i := len(data)-2; i >= 0; i-- {
			if len(data[i]) != 8 {
				fmt.Printf("Err received something weird ! EXIT")
				os.Exit(1)
			}
			if n1 < EMA1 {
				tmp, err := strconv.ParseFloat(data[i][4].(string), 64)
				if err != nil {
					fmt.Printf("Err converting floats ! EXIT")
					os.Exit(1)
				}
				ema1 = ema1 + (tmp-ema1)*alpha1
				n1 += 1
			}
			if n2 < EMA2 {
				tmp, err := strconv.ParseFloat(data[i][4].(string), 64)
				if err != nil {
					fmt.Printf("Err converting floats ! EXIT")
					os.Exit(1)
				}
				ema2 = ema2 + (tmp-ema2)*alpha2
				n2 += 1
			}
		}
		if n1 != EMA1 || n2 != EMA2 {
			fmt.Printf("Err in computing MA ! EXIT\n")
			os.Exit(1)
		}
		if time.Now().Unix() % 900 == 0 {
			if block == 0 {
				fmt.Printf("%s | ema%d = %f | ema%d = %f\n", time.Now(), EMA1, ema1, EMA2, ema2)
				block = 1
			}
		} else {
			block = 0
		}
		if start == true {
			if ema1 > ema2 && tmp_start == 0 { tmp_start = 1 }
			if tmp_start == 1 && ema1 < ema2 { tmp_start = 2 }
			if ema1 < ema2 && tmp_start == 0 { tmp_start = 2 }
			if tmp_start == 2 && ema1 > ema2 { start = false }
		}
		if dir == DIR_UP && ema1 < ema2 && start == false {
			fmt.Printf("%s | Cross sell | ema%d=%f | ema%d=%f\n", time.Now(), EMA1, ema1, EMA2, ema2)
			err = tryPlaceOrder(balance, "SELL")
			if err != nil {
				need_balance = true
				fmt.Printf("Err try placing sell order : %s\n", err)
				continue
			}
			dir = DIR_DOWN
			need_balance = true
			fmt.Println("Success sell")
		} else if dir == DIR_DOWN && ema1 > ema2 && start == false {
			fmt.Printf("Cross buy ema%d=%f | ema%d=%f\n", EMA1, ema1, EMA2, ema2)
			err = tryPlaceOrder(balance, "BUY")
			if err != nil {
				need_balance = true
				fmt.Printf("Err try placing buy order : %s\n", err)
				continue
			}
			dir = DIR_UP
			need_balance = true
			fmt.Println("Success buy")
		}
	}
}

func tryPlaceOrder(balance map[string]float64, side string) error {
	var qnt float64
	var price float64

	book, err := mexcFetchOrderbook()
	if err != nil {
		return err
	}
	book.ask += 0.01
	book.bid -= 0.01
	fmt.Println(book)
	if side == "BUY" {
		qnt = balance["USDC"]/book.ask
		price = book.ask
	} else if side == "SELL" {
		qnt = balance["ETH"]
		price = book.bid
	} else {
		return errors.New(fmt.Sprintf("Bad side %s", side))
	}
	id, err := mexcPlaceOrder(side, qnt, price)
	if err != nil {
		return err
	}
	for {
		time.Sleep(100*time.Millisecond)
		status, err := mexcFetchOrder(id)
		if err != nil {
			fmt.Printf("critical error by fetching order status, don't know if was canceled or not : %s\n", err)
			os.Exit(1) // on sait pas si l'ordre est passé ou pas
		}
		if status == "NEW" {
			continue
		}
		if status != "FILLED" {
			return errors.New("Order has been canceled")
		} else {
			break
		}
	}
	return nil
}

func mexcPlaceOrder(side string, qnt float64, price float64) (string, error) {

	price_str := strconv.FormatFloat(price, 'f', -1, 64)
	qnt_str   := strconv.FormatFloat(qnt, 'f', -1, 64)
	timestamp := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)

	body := "symbol="+SYMBOL+"&side="+side+"&type=IMMEDIATE_OR_CANCEL&quantity="+qnt_str+"&price="+price_str+"&timestamp="+timestamp
	fmt.Println(body)
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

	fmt.Println("id=",resp.OrderId)

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

func mexcFetchOrder(order_id string) (string, error) {

	timestamp := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)
	endpoint  := "symbol="+SYMBOL+"&orderId="+order_id+"&timestamp="+timestamp

	fmt.Println("url = ", MEXC_BASE_URL+MEXC_FETCH_ORDER_ED+"?"+endpoint)
	
	client := &http.Client{}
	req, err := http.NewRequest("GET", MEXC_BASE_URL+MEXC_FETCH_ORDER_ED+"?"+endpoint+"&signature="+strings.ToLower(hmac256Hex(endpoint, []byte(MEXC_API_SECRET))), nil)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MEXC_FETCH_ORDER_ED, err))
	}
	mexcAddHeaders(req)

    reqDump, err := httputil.DumpRequestOut(req, true)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("REQUEST:\n%s", string(reqDump))

	r, err := client.Do(req)
	if err != nil || r.StatusCode != 200 {
		tmp, _ := io.ReadAll(r.Body)
		return "", errors.New(fmt.Sprintf("Err %s %d %s %s", MEXC_FETCH_ORDER_ED, r.StatusCode, string(tmp), err))
	}
	var resp mexcFetchOrderResp

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&resp)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s\n", MEXC_FETCH_ORDER_ED, err))
	}
	r.Body.Close()

	fmt.Println(resp)

	return resp.Status, nil
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
