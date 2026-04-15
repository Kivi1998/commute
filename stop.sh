#!/usr/bin/env bash
# 一键停止前后端（等价于 ./dev.sh stop）
exec "$(dirname "$0")/dev.sh" stop "$@"
