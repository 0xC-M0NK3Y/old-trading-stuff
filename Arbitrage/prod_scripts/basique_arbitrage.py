import threading
import os
import sys
import ccxt
import pandas as pd
import time
from datetime import datetime

"""
	Script arbitrage basique, ccxt, okx mecx, ETH/USDC
	LOGFILE = ou ca print les ordres effectue
	STATUSTIME = tous les combien de temps ca print que les threads tournent encore (dans stdout)
"""

EXCHANGES = [
				['okx', "", "", ""],
			 	['mexc', "", ""]
			]

SYMBOL = 'ETH/USDC'

STATUSTIME = 60*60

order_buy = {}
order_sell = {}

def get_symbol(symbol, sens):
	if sens == 'buy':
		return symbol.split('/')[1]
	elif sens == 'sell':
		return symbol.split('/')[0]
	return ''

def order_thread(exchanges, order_mutex, logfp):
	orderers = {}
	global order_buy
	global order_sell
	for i in range(len(exchanges)):
		try:
			if len(exchanges[i]) > 3:
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
	p = 0
	while 1:
		if int(time.time()) % STATUSTIME == 0:
			if p == 0:
				print(f'{int(time.time())} Orderer running')
			p = 1
		else:
			p = 0
		order_mutex.acquire()
		if order_buy != {} and order_sell != {}:
			try:
				bal1 = orderers[order_buy['exchange']].fetch_balance()[get_symbol(order_buy['symbol'], 'buy')]['free'] / order_buy['price']
				bal2 = orderers[order_sell['exchange']].fetch_balance()[get_symbol(order_sell['symbol'], 'sell')]['free']
				amount = min(bal1, bal2, order_buy['quantity'], order_sell['quantity'])
				if amount > 0.0001:
					ord1 = orderers[order_buy['exchange']].create_limit_buy_order(order_buy['symbol'], amount, order_buy['price'])
					ord2 = orderers[order_sell['exchange']].create_limit_sell_order(order_sell['symbol'], amount, order_sell['price'])
					print(f'Created orders ({int(time.time())})\nOrder_buy:\n{ord1}\nOrder_sell:\n{ord2}\n', file=logfp)
					while 1:
						time.sleep(0.1)
						ord1 = orderers[order_buy['exchange']].fetch_order(ord1['id'], order_buy['symbol'])
						ord2 = orderers[order_sell['exchange']].fetch_order(ord2['id'], order_sell['symbol'])
						if ord1['filled'] != 'None' and ord2['filled'] != 'None':
							print(f'Filled orders ({int(time.time())})\nOrder_buy:\n{ord1}\nOrder_sell:\n{ord2}\n', file=logfp)
							logfp.flush()
							break
						if int(time.time()) % 60 == 0:
							print(f'{int(time.time())} : \n{ord1}\n{ord2}')
			except Exception as e:
				print(f'except :  {str(e)}')
				pass
				order_buy = {}
				order_sell = {}
		order_mutex.release()
		time.sleep(0.0001)

def price_fetcher(exchange_id, symbol, df, df_mutex):
	try:
		exchange = getattr(ccxt, exchange_id)()
	except:
		print('ERROR STARTING PRICE FETCH THREAD '+exchange_id)
		exit(1)
	print('started price fetcher '+exchange_id)
	p = 0
	while 1:
		if int(time.time()) % STATUSTIME == 0:
			if p == 0:
				print(f'{int(time.time())} Fetcher {exchange_id} running')
			p = 1
		else:
			p = 0
		orderbook = exchange.fetch_order_book(symbol)
		df_mutex.acquire()
		df['bid_'+exchange_id] = orderbook['bids'][0][:2]
		df['ask_'+exchange_id] = orderbook['asks'][0][:2]
		df_mutex.release()
		time.sleep(0.00001)


def arbitrage(df, df_mutex, order_mutex):
	best_bid   = 0
	best_ask   = 9999999999
	bid_val    = 0
	ask_val    = 0
	market_bid = ''
	market_ask = ''
	global order_buy
	global order_sell

	print('started arbitrage')
	p = 0
	while 1:
		time.sleep(0.00001)
		#print(df)
		if int(time.time()) % STATUSTIME == 0:
			if p == 0:
				print(f'{int(time.time())} Arbitror running')
			p = 1
		else:
			p = 0
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
		if best_bid - best_ask > 0.04:
			order_mutex.acquire()
			if order_buy == {} and order_sell == {}:
				order_buy = {
								"exchange": market_ask.split('_')[1],
								"symbol": SYMBOL,
								"quantity": ask_val,
								"price": best_ask
							 }
				order_sell = {
								"exchange": market_bid.split('_')[1],
								"symbol": SYMBOL,
								"quantity": bid_val,
								"price": best_bid
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

	if len(sys.argv) == 2:
		logfp = open(sys.argv[1], 'w')
		print(f'Logs in {sys.argv[1]}')
	else:
		logfp = sys.stdout
		print('Logs in stdout')

	print(f'Start, status every {STATUSTIME}')

	orderer = threading.Thread(target=order_thread, args=(EXCHANGES, order_mutex, logfp))
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

def tmppp():

	tmp = getattr(ccxt, EXCHANGES[1][0])({
		'apiKey': EXCHANGES[1][1],
		'secret': EXCHANGES[1][2]
#		'password': EXCHANGES[0][3]
	})
	b = tmp.fetch_balance()
	print(b)
	print('usdc:')
	print(b['USDC']['free'])
	print('eth')
	print(b['ETH']['free'])
	exit()

	#print(tmp.fetch_order(626299286552526851, 'BNB/USDC'))

	#tmp.cancel_order(626299286552526851, 'BNB/USDC')


	o = tmp.fetch_order_book('ETH/USDC')
	print(o['asks'][0])
	print(o['asks'][0][0])

	if o['asks'][0][1] > b['USDC']['free']/o['asks'][0][0]:
		orde = tmp.create_limit_buy_order('ETH/USDC', b['USDC']['free']/o['asks'][0][0], o['asks'][0][0])
		print(orde)
#	print(b)


if __name__ == '__main__':
	#tmppp()
	main()

