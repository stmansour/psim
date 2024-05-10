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

GSYNC=$(which gsync)
echo "gsync: ${GSYNC}"
if [ "${GSYNC}" = "" ]; then
    GSYNC="./gsync"
fi

echo "gsync: ${GSYNC}"
if [ ! -f "${GSYNC}" ]; then
    echo "gsync not found"
    exit 1
fi

#--------------------------------------------------------------------------
# Function to show usage
#--------------------------------------------------------------------------
usage() {
    echo "Usage: $0 [-d directory] [-f URLList] [-C YYYYMMDD] [-b begin_date -e end_date]"
    echo "  -d directory   Specify the base directory for downloads."
    echo "  -f URLList     Specify a file containing URLs to download."
    echo "  -C YYYYMMDD    Concatenate and process files for specified date."
    echo "  -b begin_date  Specify start date for processing."
    echo "  -e end_date    Specify end date for processing."
    echo "  If no URL file is specified, fetches the latest 15-minute list."
    exit 1
}

#--------------------------------------------------------------------------
# Download the master file list
#--------------------------------------------------------------------------
GetMasterlist() {
    url="http://data.gdeltproject.org/gdeltv2/masterfilelist.txt"
    if ! curl -L "${url}" -o masterlist.txt; then
        echo "Failed to download $url"
        exit 1
    fi
}

#--------------------------------------------------------------------------
# GenerateURLList
#--------------------------------------------------------------------------
GenerateURLList() {
    local start_date=$1
    local end_date=$2
    log "Generating URL list from $start_date to $end_date"
    echo "Generating URL list from $start_date to $end_date"
    "${GSYNC}" -d1 2024-04-15 -d2 2024-04-22 >"${URL_LIST}"
}

#--------------------------------------------------------------------------
# Log function for standardized logging
#--------------------------------------------------------------------------
log() {
    echo "$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" >>$LOGFILE
}

#--------------------------------------------------------------------------
# Function to download a file
#--------------------------------------------------------------------------
DownloadFile() {
    local url=$1
    local date_part
    date_part=$(echo "${url}" | grep -E -o '\d{8}')
    local filename
    filename=$(basename "${url}")
    local dir="$BASE_DIR/$date_part"
    local filepath="$dir/$filename"

    mkdir -p "$dir"
    if [[ -f "$filepath" ]]; then
        log "File already exists: $filepath"
        return
    fi

    log "Downloading $url..."
    if ! curl -L "${url}" -o "$filepath"; then
        log "Failed to download $url"
        echo "Failed to download $url" >>"${FAILED_DOWNLOADS}"
        exit 1
    fi
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

    log "Processing with ${GSYNC} -gf $date_part -verbose"
    if "${GSYNC}" -gf "$date_part" -verbose >"$dir/gsync-$date_part.log"; then
        if [ "$KEEP_ZIPS" -eq 0 ]; then
            rm "$dir"/*.zip
        fi
        log "${GSYNC} completed successfully for $date_part. Zip files deleted."
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
log "Starting gfetch process..."
rm -f "${FAILED_DOWNLOADS}" # Remove any previous failed downloads log

while getopts "d:f:C:b:e:hkm" opt; do
    case "${opt}" in
    b) start_date=${OPTARG} ;;
    C) CONCAT_DATE=${OPTARG} ;;
    d) BASE_DIR=${OPTARG} ;;
    e) end_date=${OPTARG} ;;
    h) usage ;;
    k) KEEP_ZIPS=1
        log "Will keep zip files"
        ;;
    m) GetMasterlist ;;
    *) usage ;;
    esac
done

shift $((OPTIND - 1))

if [[ -n "$CONCAT_DATE" ]]; then
    ConcatFiles "$CONCAT_DATE"
elif [[ -n "$start_date" && -n "$end_date" ]]; then
    START_TIME=$(date +%s)
    log "Start time: ${START_TIME}"
    GenerateURLList "$start_date" "$end_date"
    ProcessRange "$start_date" "$end_date"
    ShowDuration
else
    usage
fi
