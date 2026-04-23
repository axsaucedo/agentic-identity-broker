#!/bin/sh
# Supervisor wrapper for the sample-agent in docker-compose.
# Fetches the agent UUID for client_id=upstream-oauth2-client from the broker admin API,
# starts the binary with BROKER_AGENT_ID set, and restarts it when the UUID changes
# (e.g. after a broker hot-reload wipes in-memory storage and re-seeds with a fresh UUID).

ADMIN_API="${ADMIN_API:-http://identity-broker:14000/api}"
POLL_INTERVAL=5

fetch_uuid() {
  UUIDS=$(
    curl -sf --connect-timeout 2 --max-time 5 "${ADMIN_API}/agents" \
      -H "X-Remote-User: admin@example.com" 2>/dev/null | \
      grep -o '"id":"[^"]*","client_id":"upstream-oauth2-client"' | \
      grep -o '"id":"[^"]*"' | cut -d'"' -f4
  )

  [ -z "$UUIDS" ] && return 1

  FIRST_UUID=$(printf '%s\n' "$UUIDS" | sed -n '1p')
  SECOND_UUID=$(printf '%s\n' "$UUIDS" | sed -n '2p')

  if [ -n "$SECOND_UUID" ]; then
    echo "[start-sample-agent] Multiple agents found for client_id=upstream-oauth2-client; waiting for an unambiguous match..." >&2
    return 1
  fi

  printf '%s\n' "$FIRST_UUID"
}

wait_for_uuid() {
  echo "[start-sample-agent] Waiting for broker and seed data..." >&2
  while true; do
    UUID=$(fetch_uuid)
    [ -n "$UUID" ] && echo "$UUID" && return
    sleep 2
  done
}

is_alive() {
  kill -0 "$1" 2>/dev/null || return 1
  if [ -r "/proc/$1/status" ]; then
    grep -q '^State:.*Z' "/proc/$1/status" && return 1
  fi
  return 0
}

BINARY_PID=

cleanup() {
  if [ -n "$BINARY_PID" ]; then
    kill "$BINARY_PID" 2>/dev/null
    wait "$BINARY_PID" 2>/dev/null
  fi
  exit 0
}
trap cleanup INT TERM

while true; do
  BROKER_AGENT_ID=$(wait_for_uuid)
  export BROKER_AGENT_ID
  echo "[start-sample-agent] Starting with BROKER_AGENT_ID=${BROKER_AGENT_ID}"

  ./tmp/sample-agent mocks/sample-agent &
  BINARY_PID=$!
  UUID_CHANGED=

  while is_alive "$BINARY_PID"; do
    sleep "$POLL_INTERVAL"
    NEW_UUID=$(fetch_uuid)
    if [ -n "$NEW_UUID" ] && [ "$NEW_UUID" != "$BROKER_AGENT_ID" ]; then
      echo "[start-sample-agent] UUID changed (broker reloaded), restarting..." >&2
      kill "$BINARY_PID" 2>/dev/null
      wait "$BINARY_PID" 2>/dev/null
      BINARY_PID=
      UUID_CHANGED=1
      break
    elif [ -z "$NEW_UUID" ]; then
      echo "[start-sample-agent] fetch_uuid failed, retrying..." >&2
    fi
  done

  # Reap child if not already reaped (handles clean exit and zombie)
  if [ -n "$BINARY_PID" ]; then
    wait "$BINARY_PID" 2>/dev/null
    BINARY_PID=
  fi

  if [ -z "$UUID_CHANGED" ]; then
    echo "[start-sample-agent] Binary exited unexpectedly, restarting after delay..." >&2
    sleep 2
  fi
done
