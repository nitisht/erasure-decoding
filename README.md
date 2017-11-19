# erasure-decoding

Simple tool to understand read quorum and data/parity shards used in [reedsoloman](https://github.com/klauspost/reedsolomon) based erasure coding.

This tool calculates all possible data and parity shards based on given total number of shards. It also calculates the space used by total shards
after erasure-coding and gives you a disk space utilization ratio (disk space used after erasure-encoding/input file size). Finally it sets the
parity shards to nil and tries to reconstruct the data, until reconstruction is not possible. This gives you the read quorum required for the given
configuration.

Follow are the assumptions made by this tool:

1. Total Shards is always an even number, greater than 3, less than 257.
2. Parity Shards are at least 2.
3. Data Shards are always greater than or equal to Parity Shards.

## Usage
Build the binary using

```sh
go build erasure.go

```

Run using

```sh
./erasure -t=6 -f=testfile.txt
```

The binary accepts two flags,

- t: Total number of data and parity shards. By default this is set to 6.
- f: Input file to encode using erasure-coding. By default this [testfile](./testfile.txt) is used.
- r: Show the read quorum for given configuration. By default set to `false`.

Here is the sample output for given default values

```sh
Input file size: 220 bytes, Total shards: 6

+-------------+---------------+---------------------+
| DATA SHARDS | PARITY SHARDS | STORAGE USAGE RATIO |
+-------------+---------------+---------------------+
|           3 |             3 |                2.02 |
|           4 |             2 |                1.50 |
+-------------+---------------+---------------------+
``` 