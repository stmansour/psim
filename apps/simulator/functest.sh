#!/bin/bash

./simulator -trace -c singleInvestor.json5 -r 12345 > t1
D=$(diff t1 gold/t1.gold)
if [ $? -ne 0 ]; then
    echo "Error: diff failed"
    exit 1
fi

# Continue with the rest of the script...
result="PASS"
if [ "${D}" != "" ]; then
    result="*** FAIL ***"
    echo "output mismatch:"
    echo "${D}"
    exit 1
fi
echo "Single Investor test: ${result}"

./simulator -trace -c linguistics.json5 -r 12345 > t2
D=$(diff t2 gold/t2.gold)
if [ $? -ne 0 ]; then
    echo "Error: diff failed"
    exit 1
fi
result="PASS"
if [ "${D}" != "" ]; then
    result="*** FAIL ***"
    echo "output mismatch:"
    echo "${D}"
    exit 1
fi
echo "Linguistics-only Investor test: ${result}"
