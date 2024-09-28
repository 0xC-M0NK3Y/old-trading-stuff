package main

import (
	"fmt"
	"time"
)

func print_global_book(glob_book globalBook) {
	fmt.Println("----------------------------------")
	fmt.Printf("OKX  %s best_bid : %f quantity %f\n", glob_book.okx.symbol, glob_book.okx.bestBid, glob_book.okx.qntBid)
	fmt.Printf("OKX  %s best_ask : %f quantity %f\n", glob_book.okx.symbol, glob_book.okx.bestAsk, glob_book.okx.qntAsk)
	fmt.Printf("MECX %s best_bid : %f quantity %f\n", glob_book.mecx.symbol, glob_book.mecx.bestBid, glob_book.mecx.qntBid)
	fmt.Printf("MECX %s best_ask : %f quantity %f\n", glob_book.mecx.symbol, glob_book.mecx.bestAsk, glob_book.mecx.qntAsk)
	fmt.Println("----------------------------------")
}

func main() {

	// test main
/*
	var t orderData
	t.side = SIDE_BUY
	t.symbol = SYMNAM_MECX
	t.price = 1654.24
	t.qnt = 0.00124
	r, e := mecxPlaceOrder(t)
	fmt.Printf("resp = %s : %s\n", r, e)
	return
*/

	// real main
	var glob_book globalBook
	var okx_order orderData
	var mecx_order orderData

	okx_fetch_chan  := make(chan fetchData)
	mecx_fetch_chan := make(chan fetchData)
	okx_order_chan  := make(chan orderData)
	mecx_order_chan := make(chan orderData)
	status_chan     := make(chan int)
	status          := STATUS_NOTHING

	fmt.Println("Starting...")

	glob_book.okx.symbol = SYMNAM_OKX
	glob_book.okx.market = "okx"

	glob_book.mecx.symbol = SYMNAM_MECX
	glob_book.mecx.market = "mecx"

	go okxFetcher(SYMNAM_OKX, okx_fetch_chan)
	go mecxFetcher(SYMNAM_MECX, mecx_fetch_chan)

	go orderer(okx_order_chan, mecx_order_chan, status_chan)

	for {

		// receive data from fetchers
		tmp_okx_data := <-okx_fetch_chan
		glob_book.okx.bestBid = tmp_okx_data.bid
		glob_book.okx.qntBid  = tmp_okx_data.bidQnt
		glob_book.okx.bestAsk = tmp_okx_data.ask
		glob_book.okx.qntAsk  = tmp_okx_data.askQnt

		tmp_mecx_data := <-mecx_fetch_chan
		glob_book.mecx.bestBid = tmp_mecx_data.bid
		glob_book.mecx.qntBid  = tmp_mecx_data.bidQnt
		glob_book.mecx.bestAsk = tmp_mecx_data.ask
		glob_book.mecx.qntAsk  = tmp_mecx_data.askQnt

		print_global_book(glob_book)
		fmt.Printf("[TRADE] SELL okx %f\n", glob_book.okx.bestBid)
		fmt.Printf("[TRADE] BUY mecx %f\n", glob_book.mecx.bestAsk)
		fmt.Printf("[TRADE]          = %f\n", (glob_book.okx.bestBid - glob_book.mecx.bestAsk))
		fmt.Printf("[TRADE] SELL mecx %f\n", glob_book.mecx.bestBid)
		fmt.Printf("[TRADE] BUY okx %f\n", glob_book.okx.bestAsk)
		fmt.Printf("[TRADE]          = %f\n", (glob_book.mecx.bestBid - glob_book.okx.bestAsk))

		// if order is waiting to be filled
		if status == STATUS_WAITING {
			// if opportunity lost cancel order
			// end of order is either here, or in orderer when filled and taker on mecx
			if okx_order.side == SIDE_BUY && glob_book.mecx.bestBid - okx_order.price < 0.40 {
				status = STATUS_CANCEL_ORDER
			} else if okx_order.side == SIDE_SELL && okx_order.price - glob_book.mecx.bestAsk < -0.30 {
				status = STATUS_CANCEL_ORDER
			}
			status_chan <- status
		}
		if status == STATUS_DONE {
			fmt.Printf("[ARBITROR] Receveid DONE status from orderer\n")
			status = STATUS_NOTHING
			time.Sleep(10*time.Millisecond)
			continue
		}
		if status == STATUS_CANCEL_ORDER {
			status_chan <- status
		}
		// if no order waiting and arbitrage opportunity
		if status == STATUS_NOTHING && glob_book.okx.bestBid - glob_book.mecx.bestAsk > -0.30 {
			// create order
			mecx_order.side      = SIDE_BUY
			mecx_order.orderType = TYPE_TAKER // unused
			mecx_order.price     = glob_book.mecx.bestAsk
			mecx_order.qnt       = glob_book.mecx.qntAsk

			okx_order.side       = SIDE_SELL
			okx_order.orderType  = TYPE_MAKER // unused
			okx_order.price      = glob_book.okx.bestBid + 0.005
			okx_order.qnt        = glob_book.okx.qntBid

			status = STATUS_PLACE_ORDER

		} else if status == STATUS_NOTHING && glob_book.mecx.bestBid - glob_book.okx.bestAsk > 0.40 {
			// create order
			okx_order.side       = SIDE_BUY
			okx_order.orderType  = TYPE_MAKER // unused
			okx_order.price      = glob_book.okx.bestAsk - 0.005
			okx_order.qnt        = glob_book.okx.qntAsk

			mecx_order.side      = SIDE_SELL
			mecx_order.orderType = TYPE_TAKER // unused
			mecx_order.price     = glob_book.mecx.bestBid
			mecx_order.qnt       = glob_book.mecx.qntBid

			status = STATUS_PLACE_ORDER
		}
		if status == STATUS_PLACE_ORDER {
			// if order was created, send it to the orderer
			okx_order.market  = glob_book.okx.market
			okx_order.symbol  = glob_book.okx.symbol
			mecx_order.market = glob_book.mecx.market
			mecx_order.symbol = glob_book.mecx.symbol

			fmt.Println("[ARBITROR] Send order")

			status_chan      <- status
			okx_order_chan   <- okx_order
			mecx_order_chan  <- mecx_order

			fmt.Println("[ARBITROR] Waiting status")
			status = <- status_chan
			if status == STATUS_ORDER_FAILED {
				// if order failed, nothing to do
				fmt.Println("[ARBITROR] Order failed")
				status = STATUS_NOTHING
				continue
			}
			fmt.Println("[ARBITROR] Status = ", status)
		}
	}
}
