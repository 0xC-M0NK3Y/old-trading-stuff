package main

import (
	"fmt"
	"time"
)

func orderer(okx_chan chan orderData, mecx_chan chan orderData, status_chan chan int) {

	var status int

	for {
		fmt.Println("[ORDERER] Fetch balances")
		status = STATUS_NOTHING
		// fetch balance to know fonds before ordering
		okx_balance, err := okxFetchBalance(SYMNAM_OKX)
		if err != nil {
			fmt.Println("okxFetchBalances fail: ", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		mecx_balance, err := mecxFetchBalance()
		if err != nil {
			fmt.Println("mecxFetchBalances fail: ", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// receive order
		status     = <-status_chan
		okx_order  := <-okx_chan
		mecx_order := <-mecx_chan

		var amount float64

		// TODO: faire fonction pour "ETH" a partir de SYMBOL_
		if okx_order.side == SIDE_BUY {
			amount = min4(okx_order.qnt, okx_balance["USDC"]/okx_order.price, mecx_order.qnt, mecx_balance["ETH"])
		} else if okx_order.side == SIDE_SELL {
			amount = min4(okx_order.qnt, okx_balance["ETH"], mecx_order.qnt, mecx_balance["USDT"]/mecx_order.price)
		} else {
			fmt.Printf("Err impossible fatal\n")
			status = STATUS_ORDER_FAILED
			status_chan <-status
			time.Sleep(200 * time.Millisecond)
			continue
		}

		fmt.Println("[ORDERER] okx balance:  ", okx_balance)
		fmt.Println("[ORDERER] mecx balance: ", mecx_balance)
		fmt.Println("[ORDERER] okx order:    ", okx_order)
		fmt.Println("[ORDERER] mecx order:   ", mecx_order)
		fmt.Println("[ORDERER] amount: ", amount)

		// if order received
		if status == STATUS_PLACE_ORDER {
			// place maker order okx
			okx_order_id, err := okxPlaceOrder(okx_order)
			if err != nil {
				fmt.Printf("[ORDERER] Err placing okx order %s\n", err)
				status = STATUS_ORDER_FAILED
				status_chan <-status
				time.Sleep(500 * time.Millisecond)
				continue
			}
			// waiting for order to be filled
			status = STATUS_WAITING
			status_chan <- status
			fmt.Printf("Placed okx order %s\n", okx_order_id)

			for {
				// receiving status from arbitror
				status = <-status_chan
				if status == STATUS_CANCEL_ORDER {
					//cancel order
					err = okxCancelOrder(okx_order_id, SYMNAM_OKX)
					if err != nil {
						fmt.Printf("[ORDERER] Err canceling order %s\n", err)
						status_chan <- status
						continue
					}
					break
				}
				okx_order_status, err := okxFetchOrder(okx_order_id, SYMNAM_OKX)
				if err != nil {
					fmt.Printf("Err fetching okx order %s\n", err)
					status_chan <- status
					continue
				}
				if okx_order_status.state == "filled" || okx_order_status.state == "partially_filled" {
					mecx_order.qnt = okx_order_status.filled
					_, err := mecxPlaceOrder(mecx_order) // fill_or_kill
					if err != nil {
						fmt.Printf("Err placing mecx order %s\n", err)
						status_chan <- status
						continue
					}
					amount -= mecx_order.qnt
				}
				if amount <= 0 {
					fmt.Println("Request done success")
					status = STATUS_DONE
					status_chan <- status
					break
				}
				status_chan <- status
			}
		}
	}
}
