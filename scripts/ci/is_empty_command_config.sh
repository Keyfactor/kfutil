#!/usr/bin/env bash
set -e -o pipefail

file_path=${1:-$HOME/.keyfactor/command_config.json}

if [ -e "$file_path" ]; then
    echo "Error: File '$file_path' found, but it should not exist"
    exit 1
fi
