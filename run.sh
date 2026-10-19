#!/usr/bin/env bash
set -e
# relative paths
cd -- "$(dirname -- "$0")"

goose -dir sql/schema sqlite3 ./workouts.db up

go build -C ./src -o ../bin/out
bin/out


