#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SQL_FILE="${ROOT_DIR}/db/001_init.sql"

: "${SKYBASE_DB_HOST:=127.0.0.1}"
: "${SKYBASE_DB_PORT:=3306}"
: "${SKYBASE_DB_USER:=root}"
: "${SKYBASE_DB_NAME:=skyvv}"

if [[ -z "${SKYBASE_DB_PASSWORD:-}" ]]; then
  echo "SKYBASE_DB_PASSWORD is required" >&2
  exit 1
fi

MYSQL_PWD="${SKYBASE_DB_PASSWORD}" mysql \
  -h "${SKYBASE_DB_HOST}" \
  -P "${SKYBASE_DB_PORT}" \
  -u "${SKYBASE_DB_USER}" < "${SQL_FILE}"

echo "SkyBase schema initialized successfully on ${SKYBASE_DB_HOST}:${SKYBASE_DB_PORT}/${SKYBASE_DB_NAME}"
