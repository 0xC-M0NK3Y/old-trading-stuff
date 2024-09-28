package main

import (
	"math"
	"errors"
)

func getEMA(
	klines []mexcKline,
	ema int) (float64, error) {
	var ret float64
	
	if len(klines) < ema {
		return 0, errors.New("not enought klines to compute ema")
	}
	// lissage: 2
	a  := 2.0 / (float64(ema) + 1.0)
	ret = klines[len(klines)-1].Close

	for i := len(klines)-2; ema > 0; i, ema = i-1, ema-1 {
		ret = klines[i].Close * a + ret * (1 - a)
	}
	return ret, nil
}

func getMACD(
	klines []mexcKline,
	ema1 int,
	ema2 int) (float64, error) {

	if ema1 == ema2 {
		return 0, errors.New("ema1 and ema2 should not be the same")
	}
	emamax = math.Max(ema1, ema2)
	
	if len(klines) < emamax {
		return 0, errors.New("not enought klines to compute macd")
	}
	tmp1, _ := getEMA(klines, ema1)
	tmp2, _ := getEMA(klines, ema2)

	return tmp1 - tmp2, nil	
}
