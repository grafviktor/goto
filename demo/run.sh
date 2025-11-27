#!/bin/bash

set -e

cd "$(dirname "$(readlink -f "$0")")"

export VHS_PUBLISH=false # disable "Host your GIF on vhs.charm.sh: vhs publish <file>.gif" message
export GG_HOME=.
export GG_LOG_LEVEL=debug

function cleanup() {
    rm -rf hosts.yaml state.yaml app.log .ssh themes 2>/dev/null
}

cleanup
for file in *.tape; do
    basename=${file%.*}
    if [ -f "${basename}.yaml" ]; then
        cp "${basename}.yaml" hosts.yaml
    fi
    if [ -f "${basename}.config" ]; then
      mkdir .ssh
      cp "${basename}.config" .ssh/config
    fi

    vhs "$file" > /dev/null
    cleanup
done
