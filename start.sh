#!/usr/bin/env bash
# 一键启动前后端（等价于 ./dev.sh start）
exec "$(dirname "$0")/dev.sh" start "$@"
