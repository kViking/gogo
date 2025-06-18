#!/bin/bash

# Branch on cli, gui, or tui
if [ "$1" == "cli" ]; then
    shift
    cargo run --bin gogo-cli -- "$@"
elif [ "$1" == "gui" ]; then
    shift
    cargo run --bin gui -- "$@"
else
    echo "Usage: $0 {cli|gui} [args...]"
    exit 1
fi