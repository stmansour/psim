#!/bin/bash

#---------------------------------------------------------
# Run this command from the directory containing gdelt
#---------------------------------------------------------
GSYNC=/usr/local/plato/bin/gfetch.sh
SQLTOCSV=/usr/local/plato/bin/sqltocsv
HTTPDOCPATH=/var/www/html
BASE_DIR="./gsync/gdelt"
LOG=/home/steve/gsync/dogsync.log

#--------------------------------------------------------------------------
# Log function for standardized logging
#--------------------------------------------------------------------------
log() {
    echo "$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" >> "${LOG}"
}

#-------------------------------------------------------------
# function to calculate dates in a cross-platform manner
#-------------------------------------------------------------
calculate_date() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # MacOS system
        date -j -v-"$1"d -f "%Y%m%d" "$(date +%Y%m%d)" +%Y%m%d
    else
        # Linux system
        date -d "$1 days ago" +%Y%m%d
    fi
}

echo "Initiating GDELT Sync" > "${LOG}"
date >> "${LOG}"

#------------------------------------------
# Compute the dates needed for the script
#------------------------------------------
seven_days_ago=$(calculate_date 7)
yesterday=$(calculate_date 1)
oldest_dir=$(calculate_date 8)

#----------------------------------------------------------------
# Remove the oldest directory to maintain only 7 days of data
#----------------------------------------------------------------
log "Removing $BASE_DIR/$oldest_dir"
rm -rf "$BASE_DIR/$oldest_dir"

# Run gfetch.sh with the calculated dates
log "Initiating ${GSYNC} -b ${seven_days_ago} -e ${yesterday} -k -F"
"${GSYNC}" -b "${seven_days_ago}" -e "${yesterday}" -k -F

#------------------------------------------------------------------------------
# Create new database csv files that include the updates to the sql database...
#------------------------------------------------------------------------------
log "Generating new CSV database fileset"
rm -rf data
"${SQLTOCSV}" ; pushd data ; tsf platodb.csv ; mv platodb-filled.csv platodb.csv ; popd
log "Posting new CSV fileset to ${HTTPDOCPATH}/csv/"
tar cvf USDJPYdata.tar data ; gzip -f USDJPYdata.tar ; cp USDJPYdata.tar.gz ${HTTPDOCPATH}/csv/

log "Done"
date >> "${LOG}"