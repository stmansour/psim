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
STATUSFILE=${HTTPDOCPATH}/sync/status.txt

# Function to check logs for errors and output a summary with dates and error types
listErrors() {
    local log_dir="logs"

    # Grep for errors in the logs and parse the output
    grep "Error" "$log_dir"/*.log | while read -r line; do
        # Extract the log file name
        log_file=$(echo "$line" | cut -d':' -f1)
        # Extract the error message
        error_message=$(echo "$line" | cut -d':' -f2-)

        # Extract the date from the log file name
        log_date=$(basename "$log_file" | cut -d'_' -f2)
        log_date_formatted="${log_date:0:4}-${log_date:4:2}-${log_date:6:2}"

        # Extract the error type from the error message
        error_type=$(echo "$error_message" | cut -d':' -f1)

        # Format and append the error summary
        echo "${log_date_formatted}: ${error_type}" >> "${STATUSFILE}"
    done
}

createStatusFile() {
    LOGTODAY=$(/usr/bin/ls -lt logs/ | grep '^-' | head -n 1 | awk '{print $9}')

    TOTALERRS=$(grep "Error fetching" */*.log | wc -l)
    if [ "${TOTALERRS}x" == "x" ]; then
	TOTALERRS=0
    fi

    ERRS=$(grep "Error fetching" "logs/${LOGTODAY}")
    if (( ERRS == 0 )); then
	echo "Success - $(date)" > ${STATUSFILE}
    else
	echo "Failure - $(date)" > ${STATUSFILE}
    fi

    echo "Errors in the last 30 days: ${TOTALERRS}" >> ${STATUSFILE}
}


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
createStatusFile

#------------------------------------------------------------------------------
# Now scan the logs again and append any errors to the status file...
#------------------------------------------------------------------------------
listErrors

#------------------------------------------------------------------------------
# Create new database csv files that include the updates to the sql database...
#------------------------------------------------------------------------------
rm -rf data
"${SQLTOCSV}" >> "${LOG}" ; pushd data ; "${TSF}" platodb.csv ; mv platodb-filled.csv platodb.csv ; popd
tar cvf USDJPYdata.tar data ; gzip -f USDJPYdata.tar ; cp USDJPYdata.tar.gz ${HTTPDOCPATH}/csv/
tar xzvf USDJPYdata.tar.gz -C /usr/local/plato/bin
tar xzvf USDJPYdata.tar.gz -C /var/www/html/viewer
