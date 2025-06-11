#!/usr/bin/env bash

# Without these 3 deps, vhs is silently dying:

ffmpeg -version >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "Error: 'ffmpeg' is not installed or not found in path."
fi

ttyd -version >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "Error: 'ttyd' is not installed or not found in path."
    echo "Install ttyd from https://github.com/tsl0922/ttyd, just take the binary and put it in your PATH."
fi

jq -V >/dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "Error: 'jq' is not installed or not found in path."
    echo "Run apt install jq or brew install jq to install it."
fi

set -e

# cd to this file location
cd "$(dirname "$(readlink -f "$0")")"

export VHS_PUBLISH=false # disable "Host your GIF on vhs.charm.sh: vhs publish <file>.gif" message
TMP_HOME=temp
HOSTS_FILE="${TMP_HOME}"/hosts.yaml
STATE_FILE="${TMP_HOME}"/state.yaml

function cleanup() {
    rm -f out.gif "${TMP_HOME}"/*
}

function cleanup_or_die() {
    local hosts_compare_failed="$1"
    local state_compare_failed="$2"
    local file_name="$3"

    if [ "$hosts_compare_failed" -ne 0 ]; then
        echo "Failed: ${file_name}!"
        exit 1
    fi

    if [ "$state_compare_failed" -ne 0 ]; then
        echo "Failed: ${file_name}!"
        exit 1
    fi

    cleanup
}

mkdir -p "${TMP_HOME}"
cleanup

for file in *.tape; do
    vhs "$file" >/dev/null

    hosts_compare_failed=0
    if [ -f "${HOSTS_FILE}" ]; then
        diff "${HOSTS_FILE}" "expected/${file%.*}_hosts.yaml" # "${file%.*}" = file extension removed
        hosts_compare_failed=$?
    fi

    state_compare_failed=0
    if [ -f "${STATE_FILE}" ]; then
        diff "${STATE_FILE}" "expected/${file%.*}_state.yaml"
        state_compare_failed=$?
    fi

    cleanup_or_die $hosts_compare_failed $state_compare_failed "$file"
done
