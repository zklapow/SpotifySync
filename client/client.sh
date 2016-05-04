#!/usr/bin/env bash

GODEBUG=cgocheck=0 exec ./client "$@"
