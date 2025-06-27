#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 6 ]; then
    echo "Usage: $0 <UTIL_DB_BIN> <AIDA_DB_BASE_NAME> <FIRST_BLOCK> <LAST_BLOCK> <FIRST_EPOCH> <LAST_EPOCH>"
    exit 1
fi

UTIL_DB_BIN="$1"
AIDA_DB_BASE_NAME="$2"
FIRST_BLOCK="$3"
LAST_BLOCK="$4"
FIRST_EPOCH="$5"
LAST_EPOCH="$6"

echo "Executing metadata insert for first-block:"
echo "${UTIL_DB_BIN} metadata insert --aida-db ${AIDA_DB_BASE_NAME} fb ${FIRST_BLOCK}"
"${UTIL_DB_BIN}" metadata insert \
    --aida-db "${AIDA_DB_BASE_NAME}" \
    fb \
    "${FIRST_BLOCK}"

if [ $? -ne 0 ]; then
    echo "Error: Metadata insert (first-block) command failed. Exiting."
    exit 1
fi
echo "Metadata insert for first-block completed successfully."
echo "---------------------------------"

# --- Command 3: Insert metadata for last-block ---
echo "Executing metadata insert for last-block:"
echo "${UTIL_DB_BIN} metadata insert --aida-db ${AIDA_DB_BASE_NAME} lb ${LAST_BLOCK}"
"${UTIL_DB_BIN}" metadata insert \
    --aida-db "${AIDA_DB_BASE_NAME}" \
    lb \
    "${LAST_BLOCK}"

if [ $? -ne 0 ]; then
    echo "Error: Metadata insert (last-block) command failed. Exiting."
    exit 1
fi
echo "Metadata insert for last-block completed successfully."
echo "---------------------------------"

echo "Executing metadata insert for first-epoch:"
echo "${UTIL_DB_BIN} metadata insert --aida-db ${AIDA_DB_BASE_NAME} fe ${FIRST_EPOCH}"
"${UTIL_DB_BIN}" metadata insert \
    --aida-db "${AIDA_DB_BASE_NAME}" \
    fe \
    "${FIRST_EPOCH}"

if [ $? -ne 0 ]; then
    echo "Error: Metadata insert (first-epoch) command failed. Exiting."
    exit 1
fi
echo "Metadata insert for first-epoch completed successfully."
echo "---------------------------------"

echo "Executing metadata insert for last-epoch:"
echo "${UTIL_DB_BIN} metadata insert --aida-db ${AIDA_DB_BASE_NAME} le ${LAST_EPOCH}"
"${UTIL_DB_BIN}" metadata insert \
    --aida-db "${AIDA_DB_BASE_NAME}" \
    le \
    "${LAST_EPOCH}"

if [ $? -ne 0 ]; then
    echo "Error: Metadata insert (last-epoch) command failed. Exiting."
    exit 1
fi
echo "Metadata insert for last-epoch completed successfully."
echo "---------------------------------"

echo "Executing metadata insert for type (ty):"
echo "${UTIL_DB_BIN} metadata insert --aida-db ${AIDA_DB_BASE_NAME} ty 2"
"${UTIL_DB_BIN}" metadata insert \
    --aida-db "${AIDA_DB_BASE_NAME}" \
    ty \
    2

if [ $? -ne 0 ]; then
    echo "Error: Metadata insert (type) command failed. Exiting."
    exit 1
fi
echo "Metadata insert for type completed successfully."
echo "---------------------------------"

echo "Executing metadata insert for chain ID (ci):"
echo "${UTIL_DB_BIN} metadata insert --aida-db ${AIDA_DB_BASE_NAME} ci 146"
"${UTIL_DB_BIN}" metadata insert \
    --aida-db "${AIDA_DB_BASE_NAME}" \
    ci \
    146

if [ $? -ne 0 ]; then
    echo "Error: Metadata insert (chain ID) command failed. Exiting."
    exit 1
fi
echo "Metadata insert for chain ID completed successfully."
echo "---------------------------------"

echo "Executing metadata insert for time (ti):"
echo "${UTIL_DB_BIN} metadata insert --aida-db ${AIDA_DB_BASE_NAME} ti 0"
"${UTIL_DB_BIN}" metadata insert \
    --aida-db "${AIDA_DB_BASE_NAME}" \
    ti \
    0

if [ $? -ne 0 ]; then
    echo "Error: Metadata insert (time) command failed. Exiting."
    exit 1
fi
echo "Metadata insert for time completed successfully."
echo "---------------------------------"
