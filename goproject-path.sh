#!/usr/bin/env bash
echo "$(pwd)" | grep -oP "^$GOPATH\/\K.*"
