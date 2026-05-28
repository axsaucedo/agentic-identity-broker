#!/bin/sh
# Supervisor wrapper for the sample-agent in docker-compose.
# Fetches agent UUIDs for all three client types from the broker admin API,
# starts the binary with BROKER_*_AGENT_ID env vars set, and restarts it when
# any UUID changes (e.g. after a broker hot-reload wipes in-memory storage).

ADMIN_API="${ADMIN_API:-http://identity-broker:14000/api}"
POLL_INTERVAL=5

# fetch_agent_id_by_client_id <client_id>
fetch_agent_id_by_client_id() {
	RESPONSE=$(curl -sf --connect-timeout 2 --max-time 5 "${ADMIN_API}/agents" \
		-H "X-Remote-User: admin@example.com" 2>/dev/null) || return 1
	# Response is a flat JSON array; each object has "id" as first field.
	# Match objects containing the given client_id value and extract the id.
	echo "$RESPONSE" | grep -o '"id":"[^"]*","client_id":"'"$1"'"' |
		grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4
}

# fetch_agent_id_by_display_name <display_name>
fetch_agent_id_by_display_name() {
	RESPONSE=$(curl -sf --connect-timeout 2 --max-time 5 "${ADMIN_API}/agents" \
		-H "X-Remote-User: admin@example.com" 2>/dev/null) || return 1
	# Agents without client_id have "id" immediately before "display_name" in the JSON.
	echo "$RESPONSE" | grep -o '"id":"[^"]*","display_name":"'"$1"'"' |
		grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4
}

# fetch_cimd_client_uri — returns the first client_uri registered for the CIMD Demo Agent
fetch_cimd_client_uri() {
	[ -z "$1" ] && return 1
	RESPONSE=$(curl -sf --connect-timeout 2 --max-time 5 "${ADMIN_API}/agents/$1" \
		-H "X-Remote-User: admin@example.com" 2>/dev/null) || return 1
	echo "$RESPONSE" | grep -o '"client_uris":\["[^"]*"' |
		grep -o '"[^"]*"$' | tr -d '"'
}

# fetch_all_uuids writes PROXY_ID, LOCAL_ID, CIMD_ID to stdout as "proxy:uuid local:uuid cimd:uuid"
# Returns 1 if any UUID is missing.
fetch_all_uuids() {
	PROXY=$(fetch_agent_id_by_client_id "upstream-oauth2-client")
	LOCAL=$(fetch_agent_id_by_display_name "Local Research Agent")
	CIMD=$(fetch_agent_id_by_display_name "CIMD Demo Agent")

	[ -z "$PROXY" ] && return 1
	[ -z "$LOCAL" ] && return 1
	[ -z "$CIMD" ] && return 1

	printf '%s %s %s\n' "$PROXY" "$LOCAL" "$CIMD"
}

wait_for_uuids() {
	echo "[start-sample-agent] Waiting for broker and seed data..." >&2
	while true; do
		RESULT=$(fetch_all_uuids) && echo "$RESULT" && return
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
	UUIDS=$(wait_for_uuids)
	BROKER_PROXY_AGENT_ID=$(echo "$UUIDS" | cut -d' ' -f1)
	BROKER_LOCAL_AGENT_ID=$(echo "$UUIDS" | cut -d' ' -f2)
	BROKER_CIMD_AGENT_ID=$(echo "$UUIDS" | cut -d' ' -f3)
	BROKER_CIMD_CLIENT_URI=$(fetch_cimd_client_uri "$BROKER_CIMD_AGENT_ID")

	# Legacy single-agent env var — keep pointing at the proxy agent for backward compat.
	BROKER_AGENT_ID="$BROKER_PROXY_AGENT_ID"

	export BROKER_AGENT_ID BROKER_PROXY_AGENT_ID BROKER_LOCAL_AGENT_ID BROKER_CIMD_AGENT_ID BROKER_CIMD_CLIENT_URI
	echo "[start-sample-agent] proxy=${BROKER_PROXY_AGENT_ID} local=${BROKER_LOCAL_AGENT_ID} cimd=${BROKER_CIMD_AGENT_ID} cimd_uri=${BROKER_CIMD_CLIENT_URI}"

	./tmp/sample-agent mocks/sample-agent &
	BINARY_PID=$!
	UUID_CHANGED=

	while is_alive "$BINARY_PID"; do
		sleep "$POLL_INTERVAL"
		NEW=$(fetch_all_uuids) || NEW=""
		if [ -n "$NEW" ] && [ "$NEW" != "$UUIDS" ]; then
			echo "[start-sample-agent] Agent UUIDs changed (broker reloaded), restarting..." >&2
			kill "$BINARY_PID" 2>/dev/null
			wait "$BINARY_PID" 2>/dev/null
			BINARY_PID=
			UUID_CHANGED=1
			break
		elif [ -z "$NEW" ]; then
			echo "[start-sample-agent] fetch failed, retrying..." >&2
		fi
	done

	if [ -n "$BINARY_PID" ]; then
		wait "$BINARY_PID" 2>/dev/null
		BINARY_PID=
	fi

	if [ -z "$UUID_CHANGED" ]; then
		echo "[start-sample-agent] Binary exited unexpectedly, restarting after delay..." >&2
		sleep 2
	fi
done
