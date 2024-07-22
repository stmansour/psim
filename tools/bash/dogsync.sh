#!/bin/bash

#---------------------------------------------------------
# Run this command from the directory containing gdelt
#---------------------------------------------------------
GSYNC=/usr/local/plato/bin/gfetch
TSF=/usr/local/plato/bin/tsf
SQLTOCSV=/usr/local/plato/bin/sqltocsv
HTTPDOCPATH=/var/www/html
BASE_DIR="./gdelt"
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
rm -r "$BASE_DIR/$oldest_dir"

#------------------------------------------
# Run gfetch.sh with the calculated dates
#------------------------------------------
log "Initiating ${GSYNC} -b ${seven_days_ago} -e ${yesterday} -k -F"
"${GSYNC}" -b "${seven_days_ago}" -e "${yesterday}" -k -F

#------------------------------------------------------------------------------
# Create new database csv files that include the updates to the sql database...
#------------------------------------------------------------------------------
log "Generating new CSV database fileset"
rm -rf data
"${SQLTOCSV}" ; pushd data ; "${TSF}" platodb.csv ; mv platodb-filled.csv platodb.csv ; popd
log "Posting new CSV fileset to ${HTTPDOCPATH}/csv/"
tar cvf USDJPYdata.tar data ; gzip -f USDJPYdata.tar ; cp USDJPYdata.tar.gz ${HTTPDOCPATH}/csv/
cp data/platodb.csv /usr/local/plato/bin/data/
cp data/platodb.csv /var/www/html/viewer/data/

#------------------------------------------------------------------------------
# Now create the status from the results of yesterday's gdelt update...
#------------------------------------------------------------------------------
filename="./gdelt/${yesterday}/gsync-${yesterday}.log"
target="gdeltStatus.txt"
tail -n 11 "${filename}" > "${target}"
cp "${target}" /var/www/html/sync/

log "Done"
date >> "${LOG}"

