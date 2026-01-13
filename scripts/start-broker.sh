#!/bin/bash
# Broker startup wrapper - Auto-seeds data in background after broker starts
# This runs every time Air rebuilds, allowing auto-seed on code changes

# Color codes
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# PID file for tracking seed job (works in both Docker and native dev)
PID_FILE="./tmp/.seed_job.pid"

# Cleanup function for seed job PID file
cleanup_seed_pid() {
    rm -f "$PID_FILE"
}
trap cleanup_seed_pid EXIT

# Function to apply seed data (runs in background)
apply_seed_data() {
    # Wait for broker to be fully ready (health check + API responding)
    echo -e "${YELLOW}[Broker Wrapper]${NC} Waiting for broker API to be ready..."
    for i in {1..40}; do
        if curl -s http://localhost:14000/health > /dev/null 2>&1 && \
           curl -s http://localhost:14000/api/services > /dev/null 2>&1; then
            break
        fi
        sleep 0.5
    done

    # Set environment variables for seed scripts
    export ADMIN_API="http://localhost:14000/api"
    export BROKER_HEALTH_URL="http://localhost:14000/health"
    export MOCK_SERVER_URL="http://localhost:9000"

    echo -e "${YELLOW}[Broker Wrapper]${NC} Seeding data..."

    # Run seed scripts with improved error visibility
    if [ -f "./scripts/seed-sample-data.sh" ]; then
        # Capture output in temp file for error reporting
        SEED_LOG=$(mktemp)
        trap "rm -f $SEED_LOG" RETURN

        if bash ./scripts/seed-sample-data.sh > "$SEED_LOG" 2>&1; then
            echo -e "${GREEN}[Broker Wrapper]${NC} seed-sample-data.sh completed"
            # Show summary lines only on success
            tail -3 "$SEED_LOG" | grep -E "(Agents created|Services created|Summary)" || true
        else
            echo -e "${RED}[Broker Wrapper]${NC} seed-sample-data.sh FAILED - showing full output:"
            echo "----------------------------------------"
            cat "$SEED_LOG"
            echo "----------------------------------------"
        fi
    fi

    if [ -f "./scripts/register-mock-thirdparty-service.sh" ]; then
        REGISTER_LOG=$(mktemp)
        trap "rm -f $REGISTER_LOG" RETURN

        if bash ./scripts/register-mock-thirdparty-service.sh > "$REGISTER_LOG" 2>&1; then
            echo -e "${GREEN}[Broker Wrapper]${NC} register-mock-thirdparty-service.sh completed"
            # Show success message from script
            tail -2 "$REGISTER_LOG" | grep -E "(Mock OAuth2 service|registered)" || true
        else
            echo -e "${RED}[Broker Wrapper]${NC} register-mock-thirdparty-service.sh FAILED - showing full output:"
            echo "----------------------------------------"
            cat "$REGISTER_LOG"
            echo "----------------------------------------"
        fi
    fi

    echo -e "${GREEN}[Broker Wrapper]${NC} Auto-seed complete"
}

# Only auto-seed if seed scripts are present (indicates docker-compose environment)
if [ -f "./scripts/seed-sample-data.sh" ] || [ -f "./scripts/register-mock-thirdparty-service.sh" ]; then
    echo -e "${YELLOW}[Broker Wrapper]${NC} Auto-seed enabled - seeding will run in background"

    # Kill any existing seed job before starting new one
    if [ -f "$PID_FILE" ]; then
        OLD_PID=$(cat "$PID_FILE")
        if kill -0 "$OLD_PID" 2>/dev/null; then
            echo -e "${YELLOW}[Broker Wrapper]${NC} Terminating previous seed job (PID: $OLD_PID)"
            kill "$OLD_PID" 2>/dev/null
            sleep 1
            # Force kill if still running
            kill -9 "$OLD_PID" 2>/dev/null || true
        fi
        rm -f "$PID_FILE"
    fi

    # Start background seed job (runs after broker starts)
    apply_seed_data &
    SEED_JOB_PID=$!
    echo "$SEED_JOB_PID" > "$PID_FILE"
fi

# Exec the actual broker binary (replaces this script's process)
# The seed job continues running in background after the broker starts
exec ./tmp/agentic-identity-broker
