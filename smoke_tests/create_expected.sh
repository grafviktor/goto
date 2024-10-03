#!/bin/bash

for file in *.tape; do
    vhs "$file"
    rm out.gif
done