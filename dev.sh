#!/usr/bin/env bash
#
# Commute 本地开发启停脚本
#   ./dev.sh start     启动前后端
#   ./dev.sh stop      停止
#   ./dev.sh restart   重启
#   ./dev.sh status    查看状态
#   ./dev.sh logs      跟踪两边日志（Ctrl+C 退出但不停服务）
#   ./dev.sh build     只编译后端二进制（前端用 dev server 不需要）
#
# 约定：
#   后端端口 8090  前端端口 5173
#   日志  .dev/logs/backend.log  .dev/logs/frontend.log
#   PID   .dev/pid/backend.pid   .dev/pid/frontend.pid

set -e

# 切到脚本所在目录（确保相对路径正确）；用绝对路径避免 subshell cd 时混乱
cd "$(dirname "$0")"
ROOT="$(pwd)"

BACKEND_PORT=8090
FRONTEND_PORT=5173
DEV_DIR="$ROOT/.dev"
PID_DIR="$DEV_DIR/pid"
LOG_DIR="$DEV_DIR/logs"
BACKEND_PID="$PID_DIR/backend.pid"
FRONTEND_PID="$PID_DIR/frontend.pid"
BACKEND_LOG="$LOG_DIR/backend.log"
FRONTEND_LOG="$LOG_DIR/frontend.log"
BACKEND_BIN="$ROOT/backend/bin/server"

mkdir -p "$PID_DIR" "$LOG_DIR"

# ---- 彩色输出 ----
c_ok="\033[32m"
c_warn="\033[33m"
c_err="\033[31m"
c_dim="\033[90m"
c_reset="\033[0m"
info()  { echo -e "${c_dim}›${c_reset} $*"; }
ok()    { echo -e "${c_ok}✓${c_reset} $*"; }
warn()  { echo -e "${c_warn}!${c_reset} $*"; }
err()   { echo -e "${c_err}✗${c_reset} $*" >&2; }

# ---- 检查端口占用 ----
port_pid() {
  lsof -nP -iTCP:"$1" -sTCP:LISTEN -t 2>/dev/null | head -1
}

# ---- 检查进程存活 ----
pid_alive() {
  local pid="$1"
  [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null
}

read_pid() {
  [ -f "$1" ] && cat "$1" 2>/dev/null
}

# ---- 前置检查 ----
preflight() {
  command -v go  >/dev/null || { err "未找到 go 命令"; exit 1; }
  command -v pnpm >/dev/null || { err "未找到 pnpm 命令"; exit 1; }
  [ -f backend/.env ]   || { err "缺少 backend/.env (可从 .env.example 复制)"; exit 1; }
  [ -f frontend/.env ]  || warn "缺少 frontend/.env（地图可能不显示）"

  if ! PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d commute -c 'SELECT 1' >/dev/null 2>&1; then
    err "无法连接 PostgreSQL (127.0.0.1:5432 db=commute, user=postgres)"
    echo "  确认数据库已启动，且已执行 migrations/"
    exit 1
  fi
}

# ---- 构建后端 ----
build_backend() {
  info "编译后端..."
  ( cd "$ROOT/backend" && go build -o bin/server ./cmd/server ) || { err "后端编译失败"; exit 1; }
  ok "后端二进制: $BACKEND_BIN"
}

# ---- 启动后端 ----
start_backend() {
  if pid=$(port_pid "$BACKEND_PORT") && [ -n "$pid" ]; then
    warn "端口 $BACKEND_PORT 已被 PID $pid 占用，跳过后端启动"
    echo "$pid" > "$BACKEND_PID"
    return
  fi
  build_backend
  info "启动后端 (端口 $BACKEND_PORT)..."
  (
    cd "$ROOT/backend"
    nohup "$BACKEND_BIN" >"$BACKEND_LOG" 2>&1 &
    echo $! > "$BACKEND_PID"
  )
  for i in $(seq 1 20); do
    if curl -sf -o /dev/null "http://127.0.0.1:$BACKEND_PORT/api/v1/health"; then
      ok "后端就绪 → http://127.0.0.1:$BACKEND_PORT/api/v1/health"
      return
    fi
    sleep 0.5
  done
  err "后端 10s 内未就绪，看日志:  tail -f $BACKEND_LOG"
  exit 1
}

# ---- 启动前端 ----
start_frontend() {
  if pid=$(port_pid "$FRONTEND_PORT") && [ -n "$pid" ]; then
    warn "端口 $FRONTEND_PORT 已被 PID $pid 占用，跳过前端启动"
    echo "$pid" > "$FRONTEND_PID"
    return
  fi
  if [ ! -d "$ROOT/frontend/node_modules" ]; then
    info "首次运行，安装前端依赖..."
    ( cd "$ROOT/frontend" && pnpm install ) || { err "pnpm install 失败"; exit 1; }
  fi
  info "启动前端 dev server (端口 $FRONTEND_PORT)..."
  (
    cd "$ROOT/frontend"
    nohup pnpm dev >"$FRONTEND_LOG" 2>&1 &
    echo $! > "$FRONTEND_PID"
  )
  for i in $(seq 1 20); do
    if curl -sf -o /dev/null "http://127.0.0.1:$FRONTEND_PORT/"; then
      ok "前端就绪 → http://localhost:$FRONTEND_PORT/"
      return
    fi
    sleep 0.5
  done
  err "前端 10s 内未就绪，看日志:  tail -f $FRONTEND_LOG"
  exit 1
}

# ---- 停止 ----
stop_one() {
  local name="$1" pidfile="$2" port="$3"
  local pid
  pid=$(read_pid "$pidfile")
  if [ -z "$pid" ]; then
    # 没有 PID 文件，检查端口
    pid=$(port_pid "$port")
  fi
  if [ -n "$pid" ] && pid_alive "$pid"; then
    info "停止 $name (PID $pid)..."
    kill "$pid" 2>/dev/null || true
    for i in $(seq 1 10); do
      pid_alive "$pid" || break
      sleep 0.3
    done
    if pid_alive "$pid"; then
      warn "$name 强制 kill -9"
      kill -9 "$pid" 2>/dev/null || true
    fi
    ok "$name 已停止"
  else
    info "$name 未运行"
  fi
  rm -f "$pidfile"
}

stop_all() {
  stop_one "前端" "$FRONTEND_PID" "$FRONTEND_PORT"
  stop_one "后端" "$BACKEND_PID" "$BACKEND_PORT"
}

# ---- 状态 ----
show_status() {
  for svc in "后端:$BACKEND_PORT:$BACKEND_PID" "前端:$FRONTEND_PORT:$FRONTEND_PID"; do
    IFS=: read -r name port pidfile <<<"$svc"
    local pid
    pid=$(read_pid "$pidfile")
    if [ -n "$pid" ] && pid_alive "$pid" && port_pid "$port" >/dev/null; then
      ok "$name 运行中 PID=$pid  → http://localhost:$port"
    else
      local occupied
      occupied=$(port_pid "$port")
      if [ -n "$occupied" ]; then
        warn "$name 端口 $port 被其他进程占用 (PID $occupied)"
      else
        echo -e "${c_dim}○ $name 未运行${c_reset}"
      fi
    fi
  done
  echo
  echo -e "${c_dim}日志：$LOG_DIR/${c_reset}"
}

# ---- 命令分发 ----
case "${1:-start}" in
  start)
    preflight
    start_backend
    start_frontend
    echo
    echo -e "${c_ok}══════════════════════════════════════════${c_reset}"
    echo -e "  🚀 本地开发环境已启动"
    echo -e "     前端:  http://localhost:$FRONTEND_PORT"
    echo -e "     后端:  http://localhost:$BACKEND_PORT/api/v1/health"
    echo -e "     日志:  ./dev.sh logs"
    echo -e "     停止:  ./dev.sh stop"
    echo -e "${c_ok}══════════════════════════════════════════${c_reset}"
    ;;
  stop)
    stop_all
    ;;
  restart)
    stop_all
    sleep 0.5
    "$0" start
    ;;
  status)
    show_status
    ;;
  logs)
    info "跟踪两边日志（Ctrl+C 退出，不会停止服务）..."
    echo
    tail -F "$BACKEND_LOG" "$FRONTEND_LOG" 2>/dev/null
    ;;
  build)
    build_backend
    ;;
  *)
    echo "用法: $0 {start|stop|restart|status|logs|build}"
    exit 1
    ;;
esac
