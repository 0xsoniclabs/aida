# ShadowDb
## Overview
ShadowDb is a wrapper for any StateDb operations. It runs all operations on two StateDbs simultaneously hence slowing down the command itself.

## Using ShadowDb without existing StateDb
To run, for example, `aida-vm-sdb` with ShadowDb, we need to specify usage with the flag `--shadow-db`. Then, we specify the implementation with `--db-shadow-impl` (carmen, geth...) and the variant with `--db-shadow-variant` (go-file, cpp-file...).
Using `--keep-db` will keep both prime and shadow StateDb in the structure `path/to/state/db/tmp/prime` and `path/to/state/db/tmp/shadow`.

## Using ShadowDb with existing StateDb
To run, for example, `aida-rpc` with ShadowDb, we need to respect the expected structure. First, we specify using ShadowDb with `--shadow-db`. Then we specify the path to StateDb and ShadowDb with `--db-src`, in which we must have two StateDb directories: one named **prime** and the other named **shadow**. Implementation and Variant are both read from `statedb_info.json`.