#!/bin/bash
# Name:
#    gen_processing_reports.sh -  script for generating the block processing reports
#
# Usage: ./gen_processing_reports.sh [options]
# For help on available options, run: ./gen_processing_reports.sh --help
#
# Description:
#    Produces block processing reports in the HTML format.
#
#    The script requires a linux environment with installed commands hwinfo, free, git, go, sqlite3, and curl.
#    The script must be invoked in the main directory of the Aida repository.
#

# default values
dbimpl="carmen"
dbvariant="gofile"
carmenschema="5"
vmimpl="lfvm"
outputdir="./output"

# Parse command-line arguments using --option-name format
print_usage() {
    echo "Usage: $0 \\"
    echo "  [--output-dir <dir>]       # Output directory for results (optional, default: ./output)"
    echo "  [--db-impl <name>]         # Database implementation (optional, default: carmen)"
    echo "  [--db-variant <name>]      # Database variant (optional, default: go-file)"
    echo "  [--carmen-schema <name>]   # Carmen schema version (optional, default: 5)"
    echo "  [--vm-impl <name>]         # VM implementation (optional, default: lfvm)"
    exit 1
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --output-dir) outputdir="$2"; shift 2 ;;
        --db-impl) dbimpl="$2"; shift 2 ;;
        --db-variant) dbvariant="$2"; shift 2 ;;
        --carmen-schema) carmenschema="$2"; shift 2 ;;
        --vm-impl) vmimpl="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; print_usage ;;
    esac
done

# logging
log() {
    echo "$(date) $1" | tee -a "$outputdir/block_processing.log"
}

# Check if profile.db in output directory exists
if [ ! -f "$outputdir/profile.db" ]; then
    echo "Error: profile.db not found in output directory: $outputdir"
    exit 1
fi

#  HardwareDescription() retrieve the hardware description of the profile.db recording server
HardwareDescription()  {
    sqlite3 "$outputdir/profile.db" "SELECT processor, memory, ' disks: ', disks FROM metadata ORDER BY id DESC LIMIT 1;"
}

# CreateTimestamp() retrieves the timestamp of the profile.db recording creation
CreateTimestamp() {
    sqlite3 "$outputdir/profile.db" "SELECT createTimestamp FROM metadata ORDER BY id DESC LIMIT 1;"
}

# OperatingSystemDescription() queries the operating system of the server.
OperatingSystemDescription() {
    sqlite3 "$outputdir/profile.db" "SELECT os FROM metadata ORDER BY id DESC LIMIT 1;"
}

## GitHash() queries the git hash of the Aida repository
GitHash() {
   git rev-parse HEAD
}

## GoVersion() queries the Go version installed on the server
GoVersion() {
    go version
}

## Machine() queries machine name and IP address of server
Machine() {
    sqlite3 "$outputdir/profile.db" "SELECT machine FROM metadata ORDER BY id DESC LIMIT 1;"
}

# query the configuration
log "query configuration ..."
hw=`HardwareDescription`
tm=`CreateTimestamp`
os=`OperatingSystemDescription`
machine=`Machine`
gh=`GitHash`
go=`GoVersion`
statedb="$dbimpl($dbvariant $carmenschema)"

# render R Markdown file
log "render block processing report ..."
./scripts/knit.R -p "GitHash='$gh', HwInfo='$hw', CreateTimestamp='$tm', OsInfo='$os', Machine='$machine', GoInfo='$go', VM='$vmimpl', StateDB='$statedb'" \
                 -d "$outputdir/profile.db" -f html -o block_processing.html -O $outputdir scripts/reports/block_processing.rmd

# produce mainnet report
log "render mainnet report ..."
./scripts/knit.R -p "GitHash='$gh', HwInfo='$hw', CreateTimestamp='$tm', OsInfo='$os', Machine='$machine', GoInfo='$go', VM='$vmimpl', StateDB='$statedb'" \
                 -d "$outputdir/profile.db" -f html -o mainnet_report.html -O $outputdir scripts/reports/mainnet_report.rmd

# produce wallet transfer report
log "render wallet transfer report ..."
./scripts/knit.R -p "GitHash='$gh', HwInfo='$hw', CreateTimestamp='$tm', OsInfo='$os', Machine='$machine', GoInfo='$go', VM='$vmimpl', StateDB='$statedb'" \
                 -d "$outputdir/profile.db" -f html -o wallet_transfer.html -O $outputdir scripts/reports/wallet_transfer.rmd

# produce contract creation report
log "render contract creation report ..."
./scripts/knit.R -p "GitHash='$gh', HwInfo='$hw', CreateTimestamp='$tm', OsInfo='$os', Machine='$machine', GoInfo='$go', VM='$vmimpl', StateDB='$statedb'" \
                 -d "$outputdir/profile.db" -f html -o contract_creation.html -O $outputdir scripts/reports/contract_creation.rmd

# produce contract execution report
log "render contract execution report ..."
./scripts/knit.R -p "GitHash='$gh', HwInfo='$hw', CreateTimestamp='$tm', OsInfo='$os', Machine='$machine', GoInfo='$go', VM='$vmimpl', StateDB='$statedb'" \
                 -d "$outputdir/profile.db" -f html -o contract_execution.html -O $outputdir scripts/reports/contract_execution.rmd

# produce parallel experiment report
log "render parallel experiment report ..."
./scripts/knit.R -p "GitHash='$gh', HwInfo='$hw', CreateTimestamp='$tm', OsInfo='$os', Machine='$machine', GoInfo='$go', VM='$vmimpl', StateDB='$statedb'" \
                 -d "$outputdir/profile.db" -f html -o parallel_experiment.html -O $outputdir scripts/reports/parallel_experiment.rmd
