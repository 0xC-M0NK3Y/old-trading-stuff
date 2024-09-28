package main

import (
	"encoding/base64"
	"encoding/hex"
	"crypto/hmac"
	"crypto/sha256"
)

func min4(n1 float64, n2 float64, n3 float64, n4 float64) float64 {
	if (n1 <= n2 && n1 <= n3 && n1 <= n4) {
		return n1
	} else if (n2 <= n1 && n2 <= n3 && n2 <= n4) {
		return n2
	} else if (n3 <= n1 && n3 <= n2 && n3 <= n4) {
		return n3
	} else if (n4 <= n1 && n4 <= n2 && n4 <= n3) {
		return n4
	}
	return float64(0.0)
}

func hmac256Base64(str string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(str))
	sign := mac.Sum(nil)
    ret  := base64.StdEncoding.EncodeToString(sign)
	return ret
}

func hmac256Hex(str string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(str))
	sign := mac.Sum(nil)
	ret := hex.EncodeToString(sign)
	return ret
}
