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

func main() {

	client   := &http.Client{}
	endpoint := MEXC_FETCH_KLINES_ED+"?symbol="+SYMBOL+"&interval="+INTERVAL+"&limit="+strconv.FormatInt(EMA+1, 10)
	req, err := http.NewRequest("GET", MEXC_BASE_URL+endpoint, nil)
	if err != nil {
		fmt.Printf("Err %s\n", err)
		os.Exit(1)
	}
	req.Header.Add("Content-Type", "application/json")

	fmt.Printf("Random EMA%d | TP %f | SL %f | %s | %s\n\n", EMA, TP, SL, INTERVAL, SYMBOL)

	var balance map[string]float64
	var order_price float64
	var order_qnt float64

	order_waiting := false
	need_balance  := true
	dir           := DIR_DOWN
	block         := 0

	for {
		// fetch balance if needed
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
		// fetch klines
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
		if len(data) < EMA {
			fmt.Printf("Err received not enought klines ! EXIT\n")
			os.Exit(1)
		}
		n          := 1
		price, err := strconv.ParseFloat(data[len(data)-1][4].(string), 64)
		if err != nil {
			fmt.Printf("Err converting floats ! EXIT")
			os.Exit(1)
		}
		// compute EMA
		ema   := price
		alpha := 2.0 / (float64(EMA) + 1.0)
		for i := len(data)-2; i >= 0; i-- {
			if len(data[i]) != 8 {
				fmt.Printf("Err received something weird ! EXIT")
				os.Exit(1)
			}
			if n < EMA {
				tmp, err := strconv.ParseFloat(data[i][4].(string), 64)
				if err != nil {
					fmt.Printf("Err converting floats ! EXIT")
					os.Exit(1)
				}
				ema = ema + (tmp-ema)*alpha
				n += 1
			}

		}
		if n != EMA {
			fmt.Printf("Err in computing MA ! EXIT\n")
			os.Exit(1)
		}
		// just print price not every seconde
		if time.Now().Unix() % 300 == 0 {
			if block == 0 {
				fmt.Printf("%s |  price = %f\n", time.Now(), price)
				block = 1
			}
		} else {
			block = 0
		}
		// know what direction to take
		if price > ema {
			// TODO: verifie balance
			dir = DIR_UP
		} else {
			dir = DIR_DOWN
		}
		// if no order waiting, make one
		if dir == DIR_UP && order_waiting == false {
			fmt.Printf("Cross buy | [%s] | price %f | ema %f\n", time.Now(), price, ema)
			order_price, order_qnt, err = tryPlaceOrder(balance, "BUY")
			if err != nil {
				need_balance = true
				fmt.Printf("Err try placing sell order : %s\n", err)
				continue
			}
			fmt.Printf("Placed buy order price %f qnt %f\n", order_price, order_qnt)
			order_waiting = true
		} else if dir == DIR_DOWN && order_waiting == false {
			fmt.Printf("[%s] Cross sell | price %f | ema %f\n", time.Now(), price, ema)
			order_price, order_qnt, err = tryPlaceOrder(balance, "SELL")
			if err != nil {
				need_balance = true
				fmt.Printf("Err try placing buy order : %s\n", err)
				continue
			}
			fmt.Printf("Placed sell order price %f qnt %f\n", order_price, order_qnt)
			order_waiting = true
		}
		if order_waiting == true {
			if dir == DIR_UP {
				if price >= (order_price + (TP/100)*order_price) {
					gain_price, gain_qnt, err := tryPlaceOrder(balance, "SELL")
					if err != nil {
						need_balance = true
						fmt.Printf("Err try place gain order : %s\n", err)
						continue
					}
					fmt.Printf("gain price : %f gain qnt : %f\n", gain_price, gain_qnt)
					order_waiting = false
					need_balance = true
				}
				if price <= (order_price - (SL/100)*order_price) {
					lose_price, lose_qnt, err := tryPlaceOrder(balance, "SELL")
					if err != nil {
						need_balance = true
						fmt.Printf("Err try place lose order : %s\n", err)
						continue
					}
					fmt.Printf("lose price : %f lose qnt : %f\n", lose_price, lose_qnt)
					order_waiting = false
					need_balance = true
				}
			}
			if dir == DIR_DOWN {
				if price <= (order_price - (TP/100)*order_price) {
					gain_price, gain_qnt, err := tryPlaceOrder(balance, "BUY")
					if err != nil {
						need_balance = true
						fmt.Printf("Err try place gain order : %s\n", err)
						continue
					}
					fmt.Printf("gain price : %f gain qnt : %f\n", gain_price, gain_qnt)
					order_waiting = false
					need_balance = true
				}
				if price >= (order_price + (SL/100)*order_price) {
					lose_price, lose_qnt, err := tryPlaceOrder(balance, "BUY")
					if err != nil {
						need_balance = true
						fmt.Printf("Err try place lose order : %s\n", err)
						continue
					}
					fmt.Printf("lose price : %f lose qnt : %f\n", lose_price, lose_qnt)
					order_waiting = false
					need_balance = true
				}
			}
		}
	}
}

func tryPlaceOrder(balance map[string]float64, side string) (float64, float64, error) {
	var qnt float64
	var price float64

	book, err := mexcFetchOrderbook()
	if err != nil {
		return 0, 0, err
	}
	fmt.Println(book)
	if side == "BUY" {
		qnt = balance["USDC"]/book.ask
		price = book.ask
	} else if side == "SELL" {
		qnt = balance["ETH"]
		price = book.bid
	} else {
		return 0, 0, errors.New(fmt.Sprintf("Bad side %s", side))
	}
	fmt.Printf("Place order %s price %f qnt %f\n", side, price, qnt)
	id, err := mexcPlaceOrder(side, qnt, price, "FILL_OR_KILL")
	if err != nil {
		return 0, 0, err
	}
	time.Sleep(200*time.Millisecond)
	for {
		status, err := mexcFetchOrder(id)
		if err != nil {
			return 0, 0, err
		}
		fmt.Printf("Status = %s\n", status)
		if status == "CANCELED" {
			return 0, 0, errors.New("Order not filled")
		}
		if status == "FILLED" { 
			break
		}
		time.Sleep(500*time.Millisecond)
	}

	return price, qnt, nil
}

func mexcPlaceOrder(side string, qnt float64, price float64, typee string) (string, error) {

	price_str := strconv.FormatFloat(price, 'f', -1, 64)
	qnt_str   := strconv.FormatFloat(qnt, 'f', -1, 64)
	timestamp := strconv.FormatInt(time.Now().UTC().UnixMilli(), 10)

	body := "symbol="+SYMBOL+"&side="+side+"&type="+typee+"&quantity="+qnt_str+"&price="+price_str+"&timestamp="+timestamp
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

	client := &http.Client{}
	req, err := http.NewRequest("GET", MEXC_BASE_URL+MEXC_FETCH_ORDER_ED+"?"+endpoint+"&signature="+strings.ToLower(hmac256Hex(endpoint, []byte(MEXC_API_SECRET))), nil)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Err %s %s", MEXC_FETCH_ORDER_ED, err))
	}
	mexcAddHeaders(req)
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

	fmt.Println("order = ", resp)

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
