#!/bin/bash

# ---------------------------------------------------------------------------
# This script performs a mysqldump of the 'plato' database and stores the
# result in a file with the date and time when the backup was made.
# It also removes backup files older than 'n' days (default is 14 days).
# The script reads the database username and password from a JSON file.
# ---------------------------------------------------------------------------

# Configuration
DATABASE="plato"
BACKUP_DIR="$(dirname "$0")"
DATE=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/plato_backup_${DATE}.sql"
DAYS_TO_KEEP=${1:-14}
CREDENTIALS_FILE="/usr/local/plato/bin/extres.json5"

# ---------------------------------------------------------------------------
# Perform the database backup
# ---------------------------------------------------------------------------
#mysqldump --user="$DB_USER" --password="$DB_PASS" --databases $DATABASE > $BACKUP_FILE
mysqldump --databases $DATABASE > $BACKUP_FILE

# Check if the mysqldump command was successful
if [[ $? -eq 0 ]]; then
    echo "Backup successful: $BACKUP_FILE"

    # ---------------------------------------------------------------------------
    # Find and remove backup files older than the specified number of days
    # ---------------------------------------------------------------------------
    find $BACKUP_DIR -name "plato_backup_*.sql" -type f -mtime +$DAYS_TO_KEEP -exec rm -f {} \;
    echo "Old backups older than $DAYS_TO_KEEP days have been removed."

else
    echo "Backup failed."
    exit 1
fi

