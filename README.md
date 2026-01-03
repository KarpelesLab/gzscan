# gzscan

A multi-threaded gzip file scanner for recovering compressed data from files and disk images.

## Overview

gzscan scans files or block devices for gzip headers, enabling recovery of gzip-compressed data from raw storage. It uses parallel scanning with configurable thread count to maximize throughput on SSDs and NVMe drives.

The scanner searches for gzip magic bytes (`0x1f 0x8b 0x08`) and validates header structures according to [RFC 1952](https://datatracker.ietf.org/doc/html/rfc1952), extracting metadata such as:

- Original filename (if stored)
- Modification timestamp
- Originating operating system
- Extra fields and flags
- First bytes of decompressed content (for validation)

## Installation

```bash
go install github.com/KarpelesLab/gzscan@latest
```

## Usage

```bash
gzscan [flags] <file>
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-threads` | 2x CPU cores | Number of scanning threads |
| `-try_read` | 32 | Bytes to attempt reading from each gzip stream (0 to disable) |

### Examples

Scan a disk image:

```bash
gzscan disk.img
```

Scan a block device directly:

```bash
gzscan /dev/sda
```

Scan with custom thread count:

```bash
gzscan -threads 8 /dev/nvme0n1
```

Disable decompression validation:

```bash
gzscan -try_read 0 disk.img
```

## Output

When a likely gzip file is found, it is displayed with available metadata:

```
2022/04/21 12:21:09 found likely gzip: pos=934375424 stamp=2019-03-23 00:00:01 +0900 JST os=Unix filename=xxxxx.log
```

### Output Fields

| Field | Description |
|-------|-------------|
| `pos` | Byte offset in the scanned file |
| `stamp` | Modification timestamp stored in the gzip header |
| `os` | Operating system where the file was compressed |
| `filename` | Original filename (if stored in header) |
| `extra` | Extra field data in hex (if present) |
| `start` | First bytes of decompressed content (printable chars only) |
| `err` | Decompression error (if try_read failed) |
| `flag` | Header flags (FTEXT, FHCRC, FCOMMENT) |

## Limitations

- **False positives**: The scanner may report byte sequences that look like gzip headers but aren't actual gzip files. Headers with filenames are more likely to be genuine.

- **Incomplete files**: Large gzip files found on disk may be fragmented or have other data written over parts of them. Finding the header helps locate and potentially recover the remaining data.

- **Piped input**: gzip does not store the filename when input is piped (e.g., `getlog | gzip > file.gz`). In these cases, manual recovery may be needed.

## How It Works

1. The file is divided into equal segments based on thread count
2. Each thread scans its segment for the gzip magic bytes `0x1f 0x8b 0x08`
3. Candidate headers are validated by checking:
   - Flag byte constraints (reserved bits must be zero)
   - OS field range (0-13 for known systems)
   - Compression flags (only valid values)
4. Metadata is extracted from valid headers
5. Optionally, the first few bytes are decompressed to verify the stream

Segments overlap by 256 bytes to ensure headers spanning segment boundaries are detected.

## License

See repository for license information.
