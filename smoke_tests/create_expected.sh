#!/bin/bash

for file in *.tape; do
    vhs "$file" --output "$file".txt
    rm out.gif
done