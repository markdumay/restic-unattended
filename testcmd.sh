#!/bin/sh

# Display command-line arguments
while [ -n "$1" ]; do
case "$1" in
        * ) echo "ARG $1"
    esac
    shift
done

# Display environment
export
