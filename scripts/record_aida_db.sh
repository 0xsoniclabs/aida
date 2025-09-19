#!/bin/bash

if [ "$#" -ne 8 ]; then
    echo "Usage: $0 <util-db-binary> <sonic-binary> <sonic-path> <aida-db> <first-block> <last-block> <first-epoch> <last-epoch>"
    echo "Example: $0 /usr/local/src/aida/build/util-db /usr/local/bin/sonic /path/to/sonic.db my_aida_db 100 200 1 2"
    exit 1
fi

UTIL_DB_BIN="$1"
SONIC_BIN="$2"
SONIC_PATH="$3"
AIDA_DB_BASE_NAME="$4"
FIRST_BLOCK="$5"
LAST_BLOCK="$6"
FIRST_EPOCH="$7"
LAST_EPOCH="$8"

echo "Starting util-db and sonic operations..."
echo "---------------------------------"
echo "Arguments received:"
echo "  Util-DB Binary: ${UTIL_DB_BIN}"
echo "  Sonic Binary: ${SONIC_BIN}"
echo "  Sonic Path: ${SONIC_PATH}"
echo "  AIDA DB Base Name: ${AIDA_DB_BASE_NAME}"
echo "  First Block: ${FIRST_BLOCK}"
echo "  Last Block: ${LAST_BLOCK}"
echo "  First Epoch: ${FIRST_EPOCH}"
echo "  Last Epoch: ${LAST_EPOCH}"
echo "---------------------------------"

# TODO
#echo "Prepare from genesis if empty:"

echo "Catching up source db:"
NEXT_EPOCH=$((LAST_EPOCH + 1))
echo "${SONIC_BIN} --datadir ${SONIC_PATH}  --mode validator --exitwhensynced.epoch ${NEXT_EPOCH}"

