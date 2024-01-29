#!/bin/bash

echo -n "Single Investor test... "
./simulator -trace -c singleInvestor.json5 -r 12345 > t1
D=$(diff t1 gold/t1.gold)
# Continue with the rest of the script...
result="PASS"
if [ "${D}" != "" ]; then
    result="*** FAIL ***"
    echo "t1 output mismatch:"
    echo "${D}"
    exit 1
fi
echo "${result}"

echo -n "Linguistic Influencers test... "
./simulator -trace -c linguistics.json5 -r 12345 > t2
D=$(diff t2 gold/t2.gold)
result="PASS"
if [ "${D}" != "" ]; then
    result="*** FAIL ***"
    echo "output mismatch:"
    echo "${D}"
    exit 1
fi
echo "${result}"
