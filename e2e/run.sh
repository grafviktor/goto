#!/usr/bin/env bash

set -e

# cd to this file location
cd "$(dirname "$(readlink -f "$0")")"

export VHS_PUBLISH=false # disable "Host your GIF on vhs.charm.sh: vhs publish <file>.gif" message
TMP_HOME=temp
HOSTS_FILE="${TMP_HOME}"/hosts.yaml

function cleanup() {
    rm -f out.gif "${TMP_HOME}"/*
}

function cleanup_or_die() {
    local exit_code="$1"
    local file_name="$2"

    if [ "$exit_code" -ne 0 ]; then
        echo "Failed: ${file_name}!"
        exit 1;
    fi

    cleanup
}

mkdir -p "${TMP_HOME}"
cleanup

for file in *.tape; do
    vhs "$file" > /dev/null
    diff "${HOSTS_FILE}" "expected/${file%.*}.yaml" # file extension removed
    cleanup_or_die $? "$file"
done
