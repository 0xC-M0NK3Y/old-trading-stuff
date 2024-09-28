#!/bin/bash

TS=$(($(date +%s%N)/1000000))

SIGN=$(echo -n "timestamp=$TS" | openssl dgst -sha256 -hmac "" | cut -d ' ' -f 2)

echo "ts = $TS and sign = $SIGN"

curl -H "X-MEXC-APIKEY: " -H "Content-Type: application/json" -X GET "https://api.mexc.com/api/v3/account?timestamp=$TS&signature=$SIGN"
