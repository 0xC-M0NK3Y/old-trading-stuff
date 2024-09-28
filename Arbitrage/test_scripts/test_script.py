import threading
import os
import sys
import ccxt
import pandas as pd
import time
from datetime import datetime

EXCHANGES = ['okx','mexc']
SYMBOL = 'ETH/USDC'

def price_fetcher(exchange_id, symbol, tab, mutex):
	try:
		exchange = getattr(ccxt, exchange_id)()
	except:
		print('ERROR STARTING THREAD '+exchange_id)
		return

	while 1:
		orderbook = exchange.fetch_order_book(symbol)
		mutex.acquire()
		tab['bid_'+exchange_id] = orderbook['bids'][0][:2]
		tab['ask_'+exchange_id] = orderbook['asks'][0][:2]
		mutex.release()

def arbitrage(df, mutex, logfp):
	best_bid   = 0
	best_ask   = 9999999999
	bid_val    = 0
	ask_val    = 0
	market_bid = ''
	market_ask = ''
	while 1:
		print(df)
		time.sleep(0.1)
		###################################
		# faire des calcules du genre ici #
		###################################
		mutex.acquire()
		for col in df:
			try:
				tmp = float(df[col].values[0])
				if col.startswith('bid') and tmp > best_bid:
					best_bid   = tmp
					bid_val    = float(df[col].values[1])
					market_bid = col
				elif col.startswith('ask') and tmp < best_ask:
					ask_val    = float(df[col].values[1])
					best_ask   = tmp
					market_ask = col
			except:
				pass
		mutex.release()
		if best_ask < best_bid:
			print(f'[{datetime.now()}]', file=logfp)
			print(f'Bid on {market_bid} : {best_bid}, {bid_val}ETH', file=logfp)
			print(f'Ask on {market_ask} : {best_ask}, {ask_val}ETH', file=logfp)
			print(f'{best_bid-best_ask}*{min(bid_val, ask_val)}ETH = {(best_bid-best_ask)*min(bid_val, ask_val)}ETH', file=logfp)
			print(f'{best_bid-best_ask}*{min(bid_val, ask_val)}ETH - 0.070%*{min(bid_val, ask_val)}ETH = {((best_bid-best_ask)*min(bid_val, ask_val))-(min(bid_val, ask_val)*(0.07/100))}ETH', file=logfp)
			print('-----------------------------------------------------', file=logfp)
			logfp.flush()
		best_bid   = 0
		best_ask   = 999999999
		bid_val    = 0
		ask_val    = 0
		market_bid = ''
		market_ask = ''

def main():
	th = []
	columns = []
	for i in range(len(EXCHANGES)):
		columns.append('bid_'+EXCHANGES[i])
		columns.append('ask_'+EXCHANGES[i])
	df = pd.DataFrame(columns=columns)
	mutex = threading.Lock()


	logfp = open('arbitrage.log', 'w')

	printer = threading.Thread(target=arbitrage, args=(df, mutex, logfp,))
	printer.start()

	for i in range(len(EXCHANGES)):
		th.append(threading.Thread(target=price_fetcher, args=(EXCHANGES[i], SYMBOL, df, mutex,)))
	for i in range(len(EXCHANGES)):
		th[i].start()
	for i in range(len(EXCHANGES)):
		th[i].join()
	printer.join()

if __name__ == '__main__':
	main()
