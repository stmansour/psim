#!/bin/bash

#--------------------------------------------------------------------------
# Initialize default values
#--------------------------------------------------------------------------
BASE_DIR="./gdelt"
FAILED_DOWNLOADS="failed_downloads.log"
LOGFILE="gfetch.log"
HEADER=$'DATE\tSourceCollectionIdentifier\tSourceCommonName\tDocumentIdentifier\tCounts\tV2Counts\tThemes\tV2Themes\tLocations\tV2Locations\tPersons\tV2Persons\tOrganizations\tV2Organizations\tV2Tone\tDates\tGCAM\tSharingImage\tRelatedImages\tSocialImageEmbeds\tSocialVideoEmbeds\tQuotations\tAllNames\tAmounts\tTranslationInfo\tExtras'
URL_LIST="$BASE_DIR/urls.txt"
OS="$(uname -s)"
KEEP_ZIPS=0
GSYNCOPTS=""
GSYNC=/usr/local/plato/bin/gsync
MAX_RETRIES=3
RETRY_DELAY=3

#--------------------------------------------------------------------------
# Log function for standardized logging
#--------------------------------------------------------------------------
log() {
    echo "$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" >>$LOGFILE
}

#--------------------------------------------------------------------------
# Function to show usage
#--------------------------------------------------------------------------
usage() {
    cat <<ZZEOF
Usage: $0 [-d directory] [-f URLList] [-CYYYYMMDD] [-b begin_date -e end_date]
    -b begin_date  Specify start date for processing.
    -d directory   Specify the base directory for downloads. The default is ./gdelt
    -b begin_date  Specify start date for processing.
    -e end_date    Specify end date for processing.
    -F             Overwrite miscompares with values computed from processing GDELT files.
    -k             Keep zip files
    -m             Download masterlist.txt.  The default is to download only when needed.

Examples:
    $0 -b 20190701 -e 20190702
    $0 -b 20190701 -e 20190702 -F


ZZEOF
}

#--------------------------------------------------------------------------
# download_file - downloads with curl, retries based on MAX_RETRIES and
#                 RETRY_DELAY
#--------------------------------------------------------------------------
download_file() {
    local url="$1"
    local filepath="$2"
    local retries=0

    while [ "$retries" -lt "$MAX_RETRIES" ]; do
        if curl -L "${url}" -o "${filepath}"; then
            return 0
        else
            retries=$((retries + 1))
            log "Failed to download $url (attempt $retries/$MAX_RETRIES)"
            echo "Failed to download $url (attempt $retries/$MAX_RETRIES)" >>"${FAILED_DOWNLOADS}"
            if [ "$retries" -lt "$MAX_RETRIES" ]; then
                sleep "$RETRY_DELAY"
            fi
        fi
    done

    # If we reach here, all attempts failed
    log "Failed to download $url after $MAX_RETRIES attempts"
    echo "Failed to download $url after $MAX_RETRIES attempts" >>"${FAILED_DOWNLOADS}"
    exit 1
}

#--------------------------------------------------------------------------
# Download the master file list if needed
#--------------------------------------------------------------------------
GetMasterlist() {
    local latest_url_date
    local masterlist_file="masterlist.txt"
    local url="http://data.gdeltproject.org/gdeltv2/masterfilelist.txt"

    # Check if the masterlist file exists and read the last URL date
    if [ -f "$masterlist_file" ]; then
        # Extract the date directly from the last URL in the file
        latest_url_date=$(tail -1 "$masterlist_file" | awk '{print $3}' | grep -oE '[0-9]{8}')

        # Compare latest URL date with the end_date directly
        if [[ "$latest_url_date" < "$end_date" ]]; then
            log "Existing masterlist.txt is outdated. Downloading the latest version..."
            download_file "${url}" "$masterlist_file"
        else
            log "Existing masterlist.txt is up-to-date."
        fi
    else
        log "masterlist.txt does not exist. Downloading now..."
        download_file "${url}" "$masterlist_file"
    fi
}

#--------------------------------------------------------------------------
# GenerateURLList
#--------------------------------------------------------------------------
GenerateURLList() {
    GetMasterlist # only downloads masterlist.txt if needed
    local start_date=$1
    local end_date=$2
    log "Generating URL list from $start_date to $end_date"
    "${GSYNC}" -d1 "${start_date}" -d2 "${end_date}" >"${URL_LIST}"
}

#--------------------------------------------------------------------------
# Function to download a file
#--------------------------------------------------------------------------
DownloadFile() {
    local url=$1
    local date_part
    date_part=$(echo "${url}" | grep -E -o '[0-9]{8}')
    local filename
    filename=$(basename "${url}")
    local dir="$BASE_DIR/$date_part"
    local filepath="$dir/$filename"

    mkdir -p "$dir"
    if [[ -f "$filepath" ]]; then
        log "File already exists: $filepath"
        return
    fi

    # log "Downloading $url..."
    download_file "${url}" "${filepath}"
}

#--------------------------------------------------------------------------
# Function to concatenate and unzip files
#--------------------------------------------------------------------------
ConcatFiles() {
    local date_part=$1
    local dir="$BASE_DIR/$date_part"
    local output_file="$dir/$date_part.csv"

    log "Concatenating files into $output_file..."
    find "$dir" -name '*.zip' -exec unzip -o -d "$dir" '{}' \;
    echo "$HEADER" >"$output_file"
    cat "$dir"/*.gkg.csv >>"$output_file"
    rm "$dir"/*.gkg.csv

    log "Processing with ${GSYNC} -gf $date_part -verbose ${GSYNCOPTS}"
    if "${GSYNC}" -gf "$date_part" -verbose "${GSYNCOPTS}" >"$dir/gsync-$date_part.log"; then
        log "${GSYNC} completed successfully for $date_part."
        if [ "$KEEP_ZIPS" -eq 0 ]; then
            rm "$dir"/*.zip
            rm "${output_file}"
            log "Removed all zip files and generated CSV file for $date_part."
        else
            log "Retained all zip files for $date_part."
        fi
    else
        log "Error during ${GSYNC} for $date_part. Check gsync-$date_part.log for details."
    fi
}

#--------------------------------------------------------------------------
# Main execution logic for daily processing of a range of dates
#--------------------------------------------------------------------------
ProcessRange() {
    local start_date=$1
    local end_date=$2
    local current_date="$start_date"
    # local current_year_month_day=""

    log "Processing range from $start_date to $end_date"
    while [[ ! "$current_date" > "$end_date" ]]; do
        # Process URLs for the current day
        log "Processing URLs for $current_date"
        local day_urls
        day_urls=$(awk -v date="$current_date" '$0 ~ date {print $0}' "$URL_LIST")
        for url in $day_urls; do
            log "Downloading $url..."
            DownloadFile "$url"
        done
        log "ConcatFiles $current_date..."
        ConcatFiles "$current_date"
        # Increment date
        if [ "$OS" == "Darwin" ]; then
            current_date=$(date -j -v+1d -f "%Y%m%d" "$current_date" +%Y%m%d)
        elif [ "$OS" == "Linux" ]; then
            current_date=$(date -d "$current_date + 1 day" +%Y%m%d)
        else
            log "Unsupported OS: $OS"
            exit 1
        fi
    done
}

ShowDuration() {
    END_TIME=$(date +%s)
    log "End time: ${END_TIME}"
    ELAPSED_TIME=$((END_TIME - START_TIME))
    HOURS=$((ELAPSED_TIME / 3600))
    MINUTES=$(((ELAPSED_TIME % 3600) / 60))
    SECONDS=$((ELAPSED_TIME % 60))
    log "Elapsed time: ${HOURS}h ${MINUTES}m ${SECONDS}s"
}

#--------------------------------------------------------------------------
# Main execution starts here
#--------------------------------------------------------------------------
echo "gfetch.sh - GDELT Data Synchronization for Plato" >$LOGFILE
log "Starting gfetch process..."
rm -f "${FAILED_DOWNLOADS}" # Remove any previous failed downloads log

while getopts "d:f:C:b:e:Fhkm" opt; do
    case "${opt}" in
    b) start_date=${OPTARG} ;;
    C) CONCAT_DATE=${OPTARG} ;;
    d) BASE_DIR=${OPTARG} ;;
    e) end_date=${OPTARG} ;;
    F) GSYNCOPTS="-F"
        echo "gsync: -F option to overwrite miscompares"
        ;;
    h) usage
        exit
        ;;
    k) KEEP_ZIPS=1
        log "Will keep zip files"
        ;;
    m) GetMasterlist ;;
    *)
        usage
        exit
        ;;
    esac
done

shift $((OPTIND - 1))

if [[ -n "$CONCAT_DATE" ]]; then
    ConcatFiles "$CONCAT_DATE"
elif [[ -n "$start_date" && -n "$end_date" ]]; then
    START_TIME=$(date +%s)
    log "Start time: ${START_TIME}"
    mkdir -p "$BASE_DIR"
    GenerateURLList "$start_date" "$end_date"
    ProcessRange "$start_date" "$end_date"
    ShowDuration
else
    usage
fi