#!/bin/bash

# This script runs a series of 'util-db' commands for data scraping,
# metadata insertion, and database hash generation.
# It requires seven arguments:
# 1. util-db-binary: The path to the 'util-db' executable (e.g., /usr/local/src/aida/build/util-db)
# 2. sonic-path: The path to the Sonic database.
# 3. aida-db: The base name for the AIDA database.
# 4. first-block: The starting block number for scraping and metadata.
# 5. last-block: The ending block number for scraping and metadata.
# 6. first-epoch: The starting epoch number for metadata.
# 7. last-epoch: The ending epoch number for metadata.

# Check if the correct number of arguments is provided
if [ "$#" -ne 7 ]; then
    echo "Usage: $0 <util-db-binary> <sonic-path> <aida-db> <first-block> <last-block> <first-epoch> <last-epoch>"
    echo "Example: $0 /usr/local/src/aida/build/util-db /path/to/sonic.db my_aida_db 100 200 1 2"
    exit 1
fi

# Assign command-line arguments to descriptive variables
UTIL_DB_BIN="$1"
SONIC_PATH="$2"
AIDA_DB_PATH="$3"
FIRST_BLOCK="$4"
LAST_BLOCK="$5"
FIRST_EPOCH="$6"
LAST_EPOCH="$7"

# Construct the target database name for the 'scrape' command
TARGET_DB_FOR_SCRAPE="${AIDA_DB_PATH}-${FIRST_BLOCK}-${LAST_BLOCK}"

echo "Starting util-db operations..."
echo "---------------------------------"
echo "Arguments received:"
echo "  Util-DB Binary: ${UTIL_DB_BIN}"
echo "  Sonic Path: ${SONIC_PATH}"
echo "  AIDA DB Base Name: ${AIDA_DB_PATH}"
echo "  First Block: ${FIRST_BLOCK}"
echo "  Last Block: ${LAST_BLOCK}"
echo "  First Epoch: ${FIRST_EPOCH}"
echo "  Last Epoch: ${LAST_EPOCH}"
echo "  Generated Scrape Target DB Name: ${TARGET_DB_FOR_SCRAPE}"
echo "---------------------------------"

echo "Executing scrape command:"
echo "${UTIL_DB_BIN} scrape --target-db ${TARGET_DB_FOR_SCRAPE} --db ${-} --log debug ${FIRST_BLOCK} ${LAST_BLOCK}"
"${UTIL_DB_BIN}" scrape \
    --target-db "${TARGET_DB_FOR_SCRAPE}" \
    --db "${SONIC_PATH}" \
    --log debug \
    "${FIRST_BLOCK}" \
    "${LAST_BLOCK}"

if [ $? -ne 0 ]; then
    echo "Error: Scrape command failed. Exiting."
    exit 1
fi
echo "Scrape command completed successfully."
echo "---------------------------------"

# Call metadata_generator.sh with the arguments
"${BASH_SOURCE%/*}/metadata_generator.sh" "${UTIL_DB_BIN}" "${AIDA_DB_PATH}" "${FIRST_BLOCK}" "${LAST_BLOCK}" "${FIRST_EPOCH}" "${LAST_EPOCH}"

echo "Executing generate-db-hash command:"
echo "${UTIL_DB_BIN} generate-db-hash --aida-db ${AIDA_DB_PATH}"
"${UTIL_DB_BIN}" generate-db-hash \
    --aida-db "${AIDA_DB_PATH}"

if [ $? -ne 0 ]; then
    echo "Error: Generate DB Hash command failed. Exiting."
    exit 1
fi
echo "Generate DB Hash command completed successfully."
echo "---------------------------------"

echo "All util-db commands executed successfully."

echo "Generating patch file for AIDA database: ${AIDA_DB_PATH}.tar.gz"
# Generate the patch file as a compressed tar.gz archive
tar vfc - "${AIDA_DB_PATH}" | pigz > "${AIDA_DB_PATH}.tar.gz"
