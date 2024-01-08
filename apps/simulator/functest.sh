#!/bin/bash

./simulator -trace -c singleInvestor.json5 -r 12345 > t1
D=$(diff t1 gold/t1.gold)
result="PASS"
if [ "${D}" != "" ]; then
    result="*** FAIL ***"
fi
echo "Single Investor test: ${result}"
