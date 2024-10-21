# ddwordlist

A command-line tool to transform CSV files exported from Daum Dictionary into a word list suitable for import into Anki or other flashcard applications.

## Installation

You need to have Go installed.

To install `ddwordlist`, run:

```sh
go install github.com/grepinsight/ddwordlist@latest
```

## Usage

```bash
$ ./ddwordlist -h
Usage:
  ddwordlist --input-csv <input-csv> [--output-csv <output-csv>] [--tags <tags>]
  ddwordlist -i <input-csv> [-o <output-csv>] [--tags <tags>]

Options:
  --input-csv, -i    Input CSV file (required)
  --output-csv, -o   Output CSV file (default: stdout)
  --tags             Tags to use in the third field (default: japanese/daumdict)
  --help, -h         Show help message
  --version, -v      Show version
```k
