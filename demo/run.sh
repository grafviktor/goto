#!/bin/bash

set -e

cd "$(dirname "$(readlink -f "$0")")"

export VHS_PUBLISH=false # disable "Host your GIF on vhs.charm.sh: vhs publish <file>.gif" message
export GG_HOME=.
export GG_LOG_LEVEL=debug

function cleanup() {
    rm -f hosts.yaml state.yaml app.log 2>/dev/null
}

cleanup
for file in create_groups.tape; do
    basename=${file%.*}
    if [ -f "${basename}.yaml" ]; then
        cp "${basename}.yaml" hosts.yaml
    fi
    vhs "$file" > /dev/null
    cleanup
done
