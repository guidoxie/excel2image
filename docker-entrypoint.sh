#!/bin/bash
set -e

if [ "$(id -u)" -eq '0' ]; then
    exec gosu excel2image ./excel2image
fi

exec ./excel2image
