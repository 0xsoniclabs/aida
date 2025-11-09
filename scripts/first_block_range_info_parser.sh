#!/bin/bash
# Usage: ./first_block_range_info_parser.sh <log_file>
# Parses the first number of Block range from the log file

log_file="$1"
if [[ -z "$log_file" ]]; then
  echo "Usage: $0 <log_file>"
  exit 1
fi

# Extract the first number after 'Block range:' and trim whitespace
grep -oP 'Block range:\s*\K\d+' "$log_file" | head -n1 | xargs

