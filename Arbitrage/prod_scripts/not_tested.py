import threading
import os
import sys
import ccxt
import pandas as pd
import time
from datetime import datetime

EXCHANGES = [
				['okx', "", "", ""],
			 	['mexc', "", ""]
			]

SYMBOL = 'ETH/USDC'

order_buy = {}
order_sell = {}
state = 0

def get_symbol(symbol, sens):
	if sens == 'buy':
		return symbol.split('/')[1]
	elif sens == 'sell':
		return symbol.split('/')[0]
	return ''

def order_thread(exchanges, order_mutex):
	orderers = {}
	global order_buy
	global order_sell
	global state

	for i in range(len(exchanges)):
		try:
			if exchanges[i][0] == 'okx':
				tmp = getattr(ccxt, exchanges[i][0])({
					'apiKey': exchanges[i][1],
					'secret': exchanges[i][2],
					'password': exchanges[i][3]
				})
				orderers[exchanges[i][0]] = tmp
			else:
				tmp = getattr(ccxt, exchanges[i][0])({
					'apiKey': exchanges[i][1],
					'secret': exchanges[i][2]
				})
				orderers[exchanges[i][0]] = tmp
		except:
			print('ERROR CONNECT TO API '+exchanges[i][0])
			exit(1)
	print('started order_thread')
	ord_id = 0
	amount = 0
	while 1:
		order_mutex.acquire()
		if order_buy != {} and order_sell != {}:
			try:
				if order_buy['exchange'] == 'okx':
					okx = order_buy
					mecx = order_sell
				elif order_sell['exchange'] == 'okx':
					okx = order_sell
					mecx = order_buy
				if state == 1:
					order = orderers[okx['exchange']].fetch_order(ord_id, okx['symbol'])
					if order['filled'] != 'None':
						qnt = order['filled']
						tot = order['remaining']
						if qnt > 0:
							print(f'taker order mecx {qnt}')
							if order_buy['exchange'] == mecx['exchange']:
								orderers[mecx['exchange']].create_limit_buy_order(order_sell['symbol'], qnt, mecx['price'])
							elif order_sell['exchange'] == mecx['exchange']:
								orderers[mecx['exchange']].create_limit_sell_order(order_sell['symbol'], qnt, mecx['price'])
						if tot <= 0:
							state = 0
							order_buy = {}
							order_sell = {}
							amount = 0
							ord_id = 0
							print('done mecx taker')
				if state == 3:
					print('try cancel')
					order = orderers[okx['exchange']].fetch_order(ord_id, okx['symbol'])
					if order['status'].lower() != 'open':
						state = 0
						order_buy = {}
						order_sell = {}
						amount = 0
						ord_id = 0
						print('good status canceled')
					else:
						orderers[okx['exchange']].cancel_order(ord_id, okx['symbol'])
				if state == 0:
					bal1 = orderers[order_buy['exchange']].fetch_balance()[get_symbol(order_buy['symbol'], 'buy')]['free'] / order_buy['price']
					bal2 = orderers[order_sell['exchange']].fetch_balance()[get_symbol(order_sell['symbol'], 'sell')]['free']
					amount = min(bal1, bal2, order_buy['quantity'], order_sell['quantity'])
					if amount > 0.0001:
						if order_buy['exchange'] == 'okx':
							order = orderers[order_buy['exchange']].create_limit_buy_order(order_buy['symbol'], amount, order_buy['price'])
						elif order_sell['exchange'] == 'okx':
							order = orderers[order_sell['exchange']].create_limit_sell_order(order_sell['symbol'], amount, order_sell['price'])
						ord_id = order['id']
						state = 1
						print('amount = ')
						print(amount)
						print('--------------')
						print('buy:')
						print(order_buy)
						print('sell:')
						print(order_sell)
						print('--------------')
						print('order okx done')
			except Exception as e:
				print(f'except :  {str(e)}')
				pass
		order_mutex.release()
		time.sleep(0.001)

def price_fetcher(exchange_id, symbol, df, df_mutex):
	try:
		exchange = getattr(ccxt, exchange_id)()
	except:
		print('ERROR STARTING PRICE FETCH THREAD '+exchange_id)
		exit(1)
	print('started price fetcher '+exchange_id)
	while 1:
		orderbook = exchange.fetch_order_book(symbol)
		df_mutex.acquire()
		df['bid_'+exchange_id] = orderbook['bids'][0][:2]
		df['ask_'+exchange_id] = orderbook['asks'][0][:2]
		df_mutex.release()

def arbitrage(df, df_mutex, order_mutex):
	best_bid   = 0
	best_ask   = 9999999999
	bid_val    = 0
	ask_val    = 0
	market_bid = ''
	market_ask = ''
	global order_buy
	global order_sell
	global state

	print('started arbitrage')
	while 1:
		time.sleep(0.00001)
		#print(df)
		df_mutex.acquire()
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
		df_mutex.release()
		order_mutex.acquire()
		if state == 1 and best_bid-best_ask < 0.02:
			state = 3
		elif state == 1:
			if order_buy['exchange'] == 'okx' and best_ask < order_buy['price']:
				state = 3
			elif order_sell['exchange'] == 'okx' and best_bid < order_sell['price']:
				state = 3
		elif  best_bid-best_ask > 0.04:
			if state == 0 and order_buy == {} and order_sell == {}:
				order_buy = {
								"exchange": market_ask.split('_')[1],
								"symbol": SYMBOL,
								"quantity": ask_val,
								"price": best_ask-0.02
							 }
				order_sell = {
								"exchange": market_bid.split('_')[1],
								"symbol": SYMBOL,
								"quantity": bid_val,
								"price": best_bid+0.02
							 }
		order_mutex.release()
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
		columns.append('bid_'+EXCHANGES[i][0])
		columns.append('ask_'+EXCHANGES[i][0])
	df = pd.DataFrame(columns=columns)
	df_mutex = threading.Lock()
	order_mutex = threading.Lock()

	#order_buy = {}
	#order_sell = {}

	print('start')

	orderer = threading.Thread(target=order_thread, args=(EXCHANGES, order_mutex,))
	orderer.start()

	arbitror = threading.Thread(target=arbitrage, args=(df, df_mutex, order_mutex,))
	arbitror.start()

	for i in range(len(EXCHANGES)):
		th.append(threading.Thread(target=price_fetcher, args=(EXCHANGES[i][0], SYMBOL, df, df_mutex,)))
	for i in range(len(EXCHANGES)):
		th[i].start()
	for i in range(len(EXCHANGES)):
		th[i].join()
	arbitror.join()
	orderer.join()
	print('started')

def tmppp():

	tmp = getattr(ccxt, EXCHANGES[0][0])({
		'apiKey': EXCHANGES[0][1],
		'secret': EXCHANGES[0][2],
		'password': EXCHANGES[0][3]
	})
	b = tmp.fetch_balance()
	print(b)
	print('eth')
	print(b['ETH']['free'])
	print('usdc:')
	print(b['USDC']['free'])

	#print(tmp.fetch_order(626299286552526851, 'BNB/USDC'))
	#tmp.cancel_order(626556206697889796, 'ETH/USDC')

	o = tmp.fetch_order_book('ETH/USDC')
	print(o['bids'][0])
	print(o['bids'][0][0])

	if o['bids'][0][1] > b['ETH']['free']:
		orde = tmp.create_limit_sell_order('ETH/USDC', b['ETH']['free'], o['bids'][0][0]+0.01)
		print(orde)
#	print(b)


if __name__ == '__main__':
	tmppp()
	#main()

