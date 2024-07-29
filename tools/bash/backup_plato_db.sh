#!/bin/bash

# ---------------------------------------------------------------------------
# This script performs a mysqldump of the 'plato' database and stores the
# result in a file with the date and time when the backup was made.
# It also removes backup files older than 'n' days (default is 14 days).
# The script reads the database username and password from a JSON file.
# ---------------------------------------------------------------------------
LOG="backup.log"
echo "PLATO Database Backup Log" >${LOG}

# Configuration
DATABASE="plato"
DATABASE2="simq"
BACKUP_DIR="$(dirname "$0")"
DATE=$(date +"%Y%m%d_%H%M%S")
DAYS_TO_KEEP=${1:-14}
MYSQL_CONFIG_FILE="/home/steve/.my.cnf"

#---------------------------------------------------------------------------
# dobackup
# $1 = database name
#---------------------------------------------------------------------------
dobackup() {
    local BACKUP_FILE="${BACKUP_DIR}/${1}_backup_${DATE}.sql"
    mysqldump --defaults-extra-file="$MYSQL_CONFIG_FILE" --databases "${1}" >"$BACKUP_FILE"
    if [[ $? -eq 0 ]]; then
        echo "Backup successful: $1" >>${LOG}
        # ---------------------------------------------------------------------------
        # Find and remove backup files older than the specified number of days
        # ---------------------------------------------------------------------------
        find "${BACKUP_DIR}" -name "${1}_backup_*.sql" -type f -mtime +"$DAYS_TO_KEEP" -exec rm -f {} \;
        echo "Old backups older than $DAYS_TO_KEEP days have been removed." >>${LOG}

    else
        echo "Backup failed." >>${LOG}
    fi
}

echo "mysqldump plato" >>${LOG}
dobackup "${DATABASE}"
dobackup "${DATABASE2}"
