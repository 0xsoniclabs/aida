#!/bin/bash
# Name:
#    run_block_processing.sh - script for running the parallel experiment for block processing
#
# Usage: ./run_block_processing.sh [options]
# For help on available options, run: ./run_block_processing.sh --help
#
# Description:
#    Profiles block processing including the parallel experiment, produces in the output directory the dataset in the form
#    of a sqlite3 database, reports in the HTML format, and the log file of the script.
#
#    For a jump start to later blocks, it is possible to use an existing state-db database as a source. The state-db has
#    to be compatible with the selected database implementation, variant, Carmen schema and has to be prepared a head of time.
#
#    The script requires a Linux environment with installed commands: hwinfo, free, git, go, sqlite3, and curl.
#    The script must be invoked in the main directory of the Aida repository.
#

# Default values
substateencoding="protobuf"
buffersize="2048"
dbimpl="carmen"
dbvariant="go-file"
carmenschema="5"
vmimpl="lfvm"
startblock="first"
endblock="last"
tmpdir="/tmp"
aidadbpath="./aida-db"
outputdir="./output"
srcdbpath=""

# Parse command-line arguments using --option-name format
print_usage() {
    echo "Usage: $0 \\"
    echo "  [--aida-db <path>]         # Path to the Aida database (optional, default: ./aida-db)"
    echo "  [--output-dir <dir>]       # Output directory for results (optional, default: ./output)"
    echo "  [--db-impl <name>]         # Database implementation (optional, default: carmen)"
    echo "  [--db-variant <name>]      # Database variant (optional, default: go-file)"
    echo "  [--carmen-schema <name>]   # Carmen schema version (optional, default: 5)"
    echo "  [--vm-impl <name>]         # VM implementation (optional, default: lfvm)"
    echo "  [--tmp-dir <dir>]          # Temporary state-db location (optional, default: /tmp)"
    echo "  [--start-block <num>]      # Start block number or 'first' (optional, default: first)"
    echo "  [--end-block <num>]        # End block number or 'last' (optional, default: last)"
    echo "  [--substate-encoding <val>]# Substate encoding (optional, default: protobuf)"
    echo "  [--buffer-size <val>]      # Buffer size (optional, default: 2048)"
    echo "  [--src-db <path>]          # Path to the source state-db database. If empty, create new state-db (optional, default: <empty>)"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --aida-db) aidadbpath="$2"; shift 2 ;;
        --output-dir) outputdir="$2"; shift 2 ;;
        --db-impl) dbimpl="$2"; shift 2 ;;
        --db-variant) dbvariant="$2"; shift 2 ;;
        --carmen-schema) carmenschema="$2"; shift 2 ;;
        --vm-impl) vmimpl="$2"; shift 2 ;;
        --tmp-dir) tmpdir="$2"; shift 2 ;;
        --start-block) startblock="$2"; shift 2 ;;
        --end-block) endblock="$2"; shift 2 ;;
        --substate-encoding) substateencoding="$2"; shift 2 ;;
        --buffer-size) buffersize="$2"; shift 2 ;;
        -h|--help) print_usage ;;
        *) echo "Unknown option: $1"; print_usage ;;
    esac
done

# create output directory if doesn't exist
if [[ ! -e $outputdir ]]; then
    mkdir $outputdir
fi

# logging
log() {
    echo "$(date) $1" | tee -a "$outputdir/block_processing.log"
}

# profile block processing
log "profile block processing from $startblock to $endblock ..."
./build/aida-vm-sdb substate \
    --profile-blocks \
    --profile-db "$outputdir/profile.db" \
    --aida-db "$aidadbpath" \
    --db-impl "$dbimpl" \
    --db-variant "$dbvariant" \
    --carmen-schema "$carmenschema" \
    --vm-impl="$vmimpl" \
    --db-tmp "$tmpdir" \
    --update-buffer-size "$buffersize" \
    --substate-encoding "$substateencoding" \
    ${srcdbpath:+--db-src "$srcdbpath"} \
    "$startblock" "$endblock"

# produce block processing reports
log "produce processing reports ..."
./scripts/gen_processing_reports.sh $dbimpl $dbvariant $carmenschema $vmimpl $outputdir

log "compute tx transitions"
Rscript ./scripts/tx_transitions.R "$outputdir"

log "compute steady state"
Rscript ./scripts/tx_steady_state.R "$outputdir"

log "finished ..."
