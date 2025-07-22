#!/bin/bash
# Name:
#    gen_processing_reports.sh -  script for generating the block processing reports
#
# Synopsis:
#    gen_processing_report.sh <db-impl> <db-variant> <carmen-schema> <vm-impl> <output-dir>
#
# Description:
#    Produces block processing reports in the HTML format.
#
#    The script requires a linux environment with installed commands hwinfo, free, git, go, sqlite3, and curl.
#    The script must be invoked in the main directory of the Aida repository.
#

# check the number of command line arguments
if [ "$#" -ne 5 ]; then
    echo "Invalid number of command line arguments supplied"
    exit 1
fi

# assign variables for command line arguments
dbimpl=$1
dbvariant=$2
carmenschema=$3
vmimpl=$4
outputdir=$5

# logging
log() {
    echo "$(date) $1" | tee -a "$outputdir/block_processing.log"
}

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
