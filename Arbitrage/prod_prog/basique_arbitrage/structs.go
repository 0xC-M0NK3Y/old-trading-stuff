package main

type fetchData struct {
	bid		float64
	bidQnt	float64
	ask		float64
	askQnt	float64
}

type marketBook struct {
	symbol		string
	market		string
	bestBid		float64
	qntBid		float64
	bestAsk		float64
	qntAsk		float64
}

type globalBook struct {
	okx		marketBook
	mecx	marketBook
}

type orderData struct {
	side		int
	orderType	int
	market		string
	symbol		string
	price		float64
	qnt			float64
}

type orderStatus struct {
	filled	float64
	state	string
}
