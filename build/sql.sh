#!/usr/bin/env bash

# generate 'internal/setup/sql.json' file

CWD="$(dirname "$0")"

go run "${CWD}"/../cmd/sql-dump/main.go -dir="${CWD}"