# gohashmob
 Quickly look up hashes in your terminal using the [HashMob](https://hashmob.net/) API
 
### Features
- 🗣   Read one or more hashes from argument
  - `gohashmob [hash]`, `gohashmob [hash],[hash]`
- 📄   Read hashes from file
  - `gohashmob /path/to/file`
- ↙️   Read hashes from STDIN
  - `cat hashes.txt | gohashmob`
- ✨   Pretty print API response JSON
- 💿   Output founds in hash:plain format
- 🏷   Read API key from environmennt variable
   - `export HASHMOB_KEY=[key]`

### Installation:
```sh-session
go install github.com/n0kovo/gohashmob@latest
```

### Usage:
```sh-session
acidbrn@gibson# gohashmob -h
Reads a list of hashes and looks for their cleartext counterparts in HashMob's database.
If no positional argument is provided and the program detects a pipe, hashes are read from STDIN.
A valid API key must be supplied via the HASHMOB_KEY environment variable.

Usage: ./hashmob [-q] [-n] <hash input> (single hash / comma separated hashes / file path)
  -n	Disable JSON response prettifying
  -no-color
    	Disable colored log output
  -q	Output founds as hash:plain instead of the full API response

Examples:
   ./hashmob -q 098f6bcd4621d373cade4e832627b4f6
   cat hashes.txt | ./hashmob -q
   ./hashmob 098f6bcd4621d373cade4e832627b4f6,5f4dcc3b5aa765d61d8327deb882cf99
   ./hashmob -q path/to/hashes.txt
```
