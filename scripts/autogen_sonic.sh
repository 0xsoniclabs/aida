#!/bin/bash

if [ "$#" -ne 8 ]; then
    echo "Usage: $0 <util-db-binary> <sonicd-binary> <sonictool-binary> <source-sonic-path> <recording-sonic-path> <aida-db-path> <first-epoch> <last-epoch>"
    echo "Example: $0 /usr/local/src/aida/build/util-db /usr/local/bin/sonicd /usr/local/bin/sonictool /path/to/source-sonic /path/to/recording-sonic aida_db 1 2"
    exit 1
fi

UTIL_DB_BIN="$1"
SONICD_BIN="$2"
SONICTOOL_BIN="$3"
SOURCE_SONIC_PATH="$4"
RECORDING_SONIC_PATH="$5"
AIDA_DB_PATH="$6"
FIRST_EPOCH="$7"
LAST_EPOCH="$8"

echo "Starting util-db and sonic operations..."
echo "---------------------------------"
echo "Arguments received:"
echo "  Util-DB Binary: ${UTIL_DB_BIN}"
echo "  Sonicd Binary: ${SONICD_BIN}"
echo "  Sonictool Binary: ${SONICTOOL_BIN}"
echo "  Source Sonic Path: ${SOURCE_SONIC_PATH}"
echo "  Recording Sonic Path: ${RECORDING_SONIC_PATH}"
echo "  AIDA DB Path: ${AIDA_DB_PATH}"
echo "  First Epoch: ${FIRST_EPOCH}"
echo "  Last Epoch: ${LAST_EPOCH}"
echo "---------------------------------"

# TODO
#echo "Prepare from genesis if empty:"

# ################### SONIC ###################

echo "Catching up source db:"
NEXT_EPOCH=$((LAST_EPOCH + 1))
echo "${SONICD_BIN} --datadir ${SOURCE_SONIC_PATH}  --mode validator --exitwhensynced.epoch ${NEXT_EPOCH}"
${SONICD_BIN} --datadir ${SOURCE_SONIC_PATH}  --mode validator --exitwhensynced.epoch ${NEXT_EPOCH}
echo "Source db catch-up completed."
echo "---------------------------------"

echo "Extracting Events from Source Sonic DB:"
PATCH_EVENTS_OUTPUT="${RECORDING_SONIC_PATH}/events-${FIRST_EPOCH}-${LAST_EPOCH}"
echo "${SONICTOOL_BIN} --datadir ${SOURCE_SONIC_PATH} events export ${PATCH_EVENTS_OUTPUT} ${FIRST_EPOCH} ${LAST_EPOCH}"
${SONICTOOL_BIN} --datadir ${SOURCE_SONIC_PATH} events export ${PATCH_EVENTS_OUTPUT} ${FIRST_EPOCH} ${LAST_EPOCH}

AIDA_DB_PATCH="${AIDA_DB_PATH}-${FIRST_EPOCH}-${LAST_EPOCH}"
echo "Generating AIDA DB Patch Name: ${AIDA_DB_PATCH}"

echo ${SONICTOOL_BIN} --datadir ${RECORDING_SONIC_PATH} events import --mode validator --recording --substate-db ${AIDA_DB_PATCH} ${PATCH_EVENTS_OUTPUT}
${SONICTOOL_BIN} --datadir ${RECORDING_SONIC_PATH} events import --mode validator --recording --substate-db ${AIDA_DB_PATCH} ${PATCH_EVENTS_OUTPUT}

if [ $? -ne 0 ]; then
    echo "Error: Events import command failed. Exiting."
    exit 1
fi

rm -f ${PATCH_EVENTS_OUTPUT}

# ################### SONIC ###################
