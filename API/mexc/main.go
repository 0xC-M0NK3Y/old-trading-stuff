package main

import (
	"fmt"
	"time"
)

const ASSET       = "ETHUSDC"
const BASE_ASSET  = "ETH"
const QUOTE_ASSET = "USDC"

func lastKlinesOver(klines []mexcKline, num int, val float64) bool {
	for i := 0; i < num; i++ {
		if klines[len(klines)-1-i].Close < val {
			return false
		}
	}
	return true
}
func lastKlinesUnder(klines []mexcKline, num int, val float64) bool {
	for i := 0; i < num; i++ {
		if klines[len(klines)-1-i].Close > val {
			return false
		}
	}
	return true
}


func ema200_600_strat() {
	var balance mexcBalance
	needBalance   := true
	waitingToSell := false
	ctx, _        := mexcInit(MEXC_API_KEY, MEXC_API_SECRET)
	
	for true {
		if needBalance {
			var err error
			balance, err = mexcFetchBalance(ctx)
			if err != nil {
				fmt.Println("err8=", err)
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			needBalance = false
		}
		time.Sleep(1000 * time.Millisecond)
		klines, err := mexcFetchKlines(ctx, ASSET, MEXC_INTERVAL_1_MINUTE, 601)
		if err != nil {
			fmt.Println("err1=", err)
			continue
		}
		ema600, err := getEMA(klines, 600)
		if err != nil {
			fmt.Println("err2=", err)
		}
		ema200, err := getEMA(klines, 200)
		if err != nil {
			fmt.Println("err3=", err)
		}
		if waitingToSell == false && lastKlinesOver(klines, 60, ema600) && lastKlinesOver(klines, 2, ema200) {
			currentPrice   := klines[len(klines)-1].Close + 10.0 // +10 to perform market like order, as fill or kill
			quantityWanted := balance.Map[QUOTE_ASSET].Free/currentPrice
			
			currentOrder, err := mexcNewOrder(ctx, ASSET, MEXC_SIDE_BUY, MEXC_ORDER_TYPE_IMMEDIATE_OR_CANCEL, quantityWanted, 0, currentPrice)
			// big problem if the fucking json decoder crashes, but should never happen
			// if mexc are not mfs
			if err != nil {
				fmt.Println("err7=", err)
				continue
			}
			time.Sleep(1000 * time.Millisecond)
			orderStatus, err := mexcQueryOrder(ctx, currentOrder)
			if err != nil {
				fmt.Println("err4=", err)
				needBalance = true
				continue
			}
			if orderStatus.Status == MEXC_ORDER_STATUS_FILLED {
				needBalance   = true
				waitingToSell = true
			}
		}
		if waitingToSell == true && lastKlinesUnder(klines, 1, ema200) {
			currentPrice   := klines[len(klines)-1].Close - 10.0
			quantityWanted := balance.Map[BASE_ASSET].Free
			
			currentOrder, err := mexcNewOrder(ctx, ASSET, MEXC_SIDE_SELL, MEXC_ORDER_TYPE_IMMEDIATE_OR_CANCEL, quantityWanted, 0, currentPrice)
			if err != nil {
				fmt.Println("err5=", err)
				continue
			}
			time.Sleep(1000 * time.Millisecond)
			orderStatus, err := mexcQueryOrder(ctx, currentOrder)
			if err != nil {
				fmt.Println("err6=", err)
				needBalance = true
				continue
			}
			if orderStatus.Status == MEXC_ORDER_STATUS_FILLED {
				needBalance   = true
				waitingToSell = false
			}
		}
	}
	return
}

func macd_strat() {
	
}

func main() {

}
