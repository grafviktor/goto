#!/bin/bash

set -e

TMP_HOME=temp
HOSTS_FILE="${TMP_HOME}"/hosts.yaml

function clean() {
    rm -f out.gif "${TMP_HOME}"/*
}

mkdir -p "${TMP_HOME}"
clean

vhs 1_create.tape > /dev/null
diff "${HOSTS_FILE}" expected/1_create.yaml || (echo Failed: 1_create.tape; exit 1)
clean

vhs 2_copy.tape > /dev/null
diff "${HOSTS_FILE}" expected/2_copy.yaml || (echo Failed: 2_copy.tape; exit 1)
clean

vhs 3_delete.tape > /dev/null
diff "${HOSTS_FILE}" expected/3_delete.yaml || (echo Failed: 3_delete.tape; exit 1)
clean
