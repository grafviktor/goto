#!/bin/bash

cd $(dirname "$(readlink -f "$0")")

export VHS_PUBLISH=false # disable "Host your GIF on vhs.charm.sh: vhs publish <file>.gif" message
export GG_HOME=.
export GG_LOG_LEVEL=debug

function cleanup() {
    rm hosts.yaml state.yaml app.log 2>/dev/null
}

cleanup
for file in *.tape; do
    basename=${file%.*}
    cp "${basename}.yaml" hosts.yaml 2>/dev/null
    vhs "$file" > /dev/null
    cleanup
done
