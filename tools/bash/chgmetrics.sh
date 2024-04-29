#!/bin/bash

# Function to replace old metric names with new ones in a file
replace_metrics() {
    local file="$1"
    echo "Modifying: $file"

    # Detect operating system and set sed in-place option accordingly
    SED_OPT=""
    if [[ "$(uname)" == "Darwin" ]]; then
        SED_OPT="-i ''" # macOS requires an empty string as an argument to -i
    else
        SED_OPT="-i" # Linux and other Unix-like systems do not
    fi

    # Apply sed command with the appropriate option
    sed $SED_OPT \
        -e 's/Metric=BC/Metric=BusinessConfidence/g' \
        -e 's/Metric=BP/Metric=BuildingPermits/g' \
        -e 's/Metric=CC/Metric=ConsumerConfidence/g' \
        -e 's/Metric=CP/Metric=CorporateProfits/g' \
        -e 's/Metric=CU/Metric=CapacityUtilization/g' \
        -e 's/Metric=DR/Metric=InterestRate/g' \
        -e 's/Metric=GD/Metric=GovernmentDebttoGDP/g' \
        -e 's/Metric=HS/Metric=HousingStarts/g' \
        -e 's/Metric=IE/Metric=InflationExpectations/g' \
        -e 's/Metric=IP/Metric=IndustrialProduction/g' \
        -e 's/Metric=IR/Metric=InflationRate/g' \
        -e 's/Metric=M1/Metric=MoneySupplyM1/g' \
        -e 's/Metric=M2/Metric=MoneySupplyM2/g' \
        -e 's/Metric=MP/Metric=ManufacturingProduction/g' \
        -e 's/Metric=RS/Metric=RetailSalesMoM/g' \
        -e 's/Metric=SP/Metric=StockMarket/g' \
        -e 's/Metric=UR/Metric=UnemploymentRate/g' \
        "$file"
}

export -f replace_metrics

# Find all .json5, .csv, and .go files starting from the current directory, and apply the replacement
find . \( -name "*.json5" -o -name "*.go" \) -exec bash -c 'replace_metrics "$0"' {} \;

echo "Metric names have been updated."

