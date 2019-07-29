#!/bin/bash
set -e

if [[ -n "${MAUTIC_DB_HOST}" ]]; then
    exec "$@" -host="${MAUTIC_DB_HOST}" -port="${MAUTIC_DB_PORT}" -user="${MAUTIC_DB_USER}" -db="${MAUTIC_DB_NAME}" -tableprefix="${MAUTIC_TABLE_PREFIX}" -pass="${MAUTIC_DB_PASSWORD}"
else
    exec "$@"
fi
