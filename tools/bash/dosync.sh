#!/bin/bash

#------------------------------
# Configuration
#------------------------------
SYNC=/usr/local/plato/bin/psync
TSF=/usr/local/plato/bin/tsf
LOG_DIR=./logs
MAX_LOGS=30
SQLTOCSV=/usr/local/plato/bin/sqltocsv
HTTPDOCPATH=/var/www/html

#------------------------------
# Ensure log directory exists
#------------------------------
mkdir -p "${LOG_DIR}"

#------------------------------------
# Check if the log directory is empty
#------------------------------------
if [ -z "$(ls -A "${LOG_DIR}")" ]; then
    num_logs=0
else
    #-----------------------------------------
    # Get the number of existing log files
    #-----------------------------------------
    num_logs=$(ls -1q "${LOG_DIR}"/*.log | wc -l)

    #-----------------------------------------------------------------
    # Check if maximum logs reached and delete the oldest if necessary
    #-----------------------------------------------------------------
    if [ "${num_logs}" -ge "${MAX_LOGS}" ]; then
        oldest_log=$(ls -1t "${LOG_DIR}"/*.log | tail -1)
        rm -f "${oldest_log}"
    fi
fi

#--------------------------------------
# Generate timestamped log file name
#--------------------------------------
LOG="${LOG_DIR}/psync_$(date +'%Y%m%d_%H%M%S').log"

#------------------------
# Now do the sync...
#------------------------
"${SYNC}" -F -verbose > "${LOG}"

#-----------------------------------------------------------------
# Mark the failure if there were errors in any of the fetches...
#-----------------------------------------------------------------
ERRS=$(grep "Error fetching" */*.log | wc -l)
if [ "${ERRS}x" == "x" ]; then
    ERRS=0
fi
if (( ${ERRS} == 0 )); then 
    echo "Success - $(date)" > ${HTTPDOCPATH}/sync/status.txt
else
    echo "${ERRS} Error(s) - $(date)" > ${HTTPDOCPATH}/sync/status.txt
fi

#------------------------------------------------------------------------------
# Create new database csv files that include the updates to the sql database...
#------------------------------------------------------------------------------
rm -rf data
"${SQLTOCSV}" >> "${LOG}" ; pushd data ; "${TSF}" platodb.csv ; mv platodb-filled.csv platodb.csv ; popd
tar cvf USDJPYdata.tar data ; gzip -f USDJPYdata.tar ; cp USDJPYdata.tar.gz ${HTTPDOCPATH}/csv/
cp data/platodb.csv /usr/local/plato/bin/data/
cp data/platodb.csv /var/www/html/viewer/data/
