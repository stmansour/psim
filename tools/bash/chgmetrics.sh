#!/bin/bash

# Function to replace old metric names with new ones in a file
replace_metrics() {
    local file="$1"
    echo "Modifying: $file"
    sed -i '' \
        -e 's/LSNScore_ECON/GCAM_C3_1_ECON/g' \
        -e 's/LSPScore_ECON/GCAM_C3_2_ECON/g' \
        -e 's/WHAScore_ECON/GCAM_C15_137_ECON/g' \
        -e 's/WHOScore_ECON/GCAM_C15_147_ECON/g' \
        -e 's/WHLScore_ECON/GCAM_C15_148_ECON/g' \
        -e 's/WPAScore_ECON/GCAM_C15_204_ECON/g' \
        -e 's/WDECount_ECON/GCAM_C16_47_ECON/g' \
        -e 's/WDFCount_ECON/GCAM_C16_60_ECON/g' \
        -e 's/WDPCount_ECON/GCAM_C16_121_ECON/g' \
        -e 's/LIMCount_ECON/GCAM_C5_4_ECON/g' \
        -e 's/LSNScore/GCAM_C3_1/g' \
        -e 's/LSPScore/GCAM_C3_2/g' \
        -e 's/WHAScore/GCAM_C15_137/g' \
        -e 's/WHOScore/GCAM_C15_147/g' \
        -e 's/WHLScore/GCAM_C15_148/g' \
        -e 's/WPAScore/GCAM_C15_204/g' \
        -e 's/WDECount/GCAM_C16_47/g' \
        -e 's/WDFCount/GCAM_C16_60/g' \
        -e 's/WDPCount/GCAM_C16_121/g' \
        -e 's/WDMCount/GCAM_C5_4/g' \
        "$file"
}

export -f replace_metrics

# Find all .json5 and .csv files starting from the current directory, and apply the replacement
find . \( -name "*.json5" -o -name "*.csv" -o -name "*.go" \) -exec bash -c 'replace_metrics "$0"' {} \;

echo "Metric names have been updated."
