#!/bin/bash

cd $(dirname "$(readlink -f "$0")")

export GG_HOME=.
export GG_LOG_LEVEL=debug

vhs use_custom_config.tape
vhs goto_and_tmux.tape