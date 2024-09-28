package main

import (
	"io"
	"fmt"
	"bytes"
	"net/http"
	"errors"
	"strings"
	"encoding/hex"
	"crypto/hmac"
	"crypto/sha256"

	//"net/http/httputil"
)


func makeRequest(
	ctx mexcContext,
	method string,
	endpoint string,
	body []byte,
	signed bool) (*http.Response, error) {
	var req *http.Request
	var ret *http.Response
	var err error
	
	if body != nil {
		req, err = http.NewRequest(method, MEXC_BASE_URL+endpoint, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, MEXC_BASE_URL+endpoint, nil)
	}
	if err != nil {
		return nil, err
	}
	if signed {
		req.Header["X-MEXC-APIKEY"] = []string{ctx.ApiKey}
	}
	req.Header.Set("Content-Type", "application/json")

	// usefull to debug
    //reqDump, err := httputil.DumpRequestOut(req, true)
    //if err != nil {
    //    fmt.Println(err)
    //}
    //fmt.Printf("REQUEST:\n%s", string(reqDump))
	
	ret, err = ctx.Client.Do(req)
	if err != nil || ret.StatusCode != 200 {
		tmp, _ := io.ReadAll(ret.Body)
		return nil, errors.New(fmt.Sprintf("%d:%s:%s", ret.StatusCode, err, string(tmp)))
	}
	return ret, nil
}

func getSignature(
	ctx mexcContext,
	body string) string {
	return strings.ToLower(hmac256Hex(body, ctx.ApiSecret))
}

func hmac256Hex(
	str string,
	key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(str))
	sign := mac.Sum(nil)
	ret  := hex.EncodeToString(sign)
	return ret
}
