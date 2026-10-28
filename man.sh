#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
action="${1:-status}"
target="${2:-none}"
case "$target" in backend|frontend|none) ;; *) echo "대상은 backend, frontend, none 중 하나입니다." >&2; exit 2;; esac
mkdir -p build .run
pidfile=.run/admin.pid
suffix=""
case "$(uname -s)" in MINGW*|MSYS*|CYGWIN*) suffix=.exe;; esac
binary="$(pwd)/build/maru-admin${suffix}"
running() { [[ -f "$pidfile" ]] && [[ "$(cat "$pidfile")" =~ ^[0-9]+$ ]] && kill -0 "$(cat "$pidfile")" 2>/dev/null && [[ "$(ps -p "$(cat "$pidfile")" -f)" == *maru-admin* ]]; }
start() {
 if running; then echo "Admin이 이미 실행 중입니다. PID $(cat "$pidfile")"; return; fi
 "${GO_BIN:-go}" build -o "$binary" ./admin/backend
 nohup "$binary" >.run/admin.log 2>&1 &
 echo "$!" >"$pidfile"
 sleep 1
 if ! running; then cat .run/admin.log; return 1; fi
 echo "Admin backend와 임베드 frontend를 시작했습니다. PID $(cat "$pidfile")"
}
stop() {
 if running; then kill "$(cat "$pidfile")"; else echo "실행 중인 Admin이 없습니다."; fi
 rm -f "$pidfile"
}
case "$action" in
 start) start;;
 stop) stop;;
 status) if running; then echo "실행 중: PID $(cat "$pidfile")"; else echo "중지됨"; fi;;
 restart) stop; start;;
 *) echo "사용법: ./man.sh [start|stop|status|restart] [backend|frontend|none]" >&2; exit 2;;
esac
