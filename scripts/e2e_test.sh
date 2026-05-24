#!/usr/bin/env bash
set -euo pipefail

# Archon E2E Test Suite
# This script builds the binary and runs integration tests for doctor, prompt, and piping behaviors.

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "Building Archon..."
make build

# 1. Test Version Command
echo -n "Testing version command... "
VERSION=$(./bin/archon --version)
if [[ "$VERSION" == *"archon version 0.1.0"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (got: $VERSION)${NC}"
    exit 1
fi

# 2. Test Doctor Command
echo -n "Testing doctor command... "
if ./bin/archon doctor > /dev/null; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    exit 1
fi

# 3. Test Mock Prompt Command (Agent-to-Agent Piping)
echo -n "Testing mock prompt command (Piping stdout)... "
MOCK_VAL="Socratic prompt: What are your data consistency requirements?"
export ARCHON_MOCK_RESPONSE="$MOCK_VAL"

# Run prompt command and capture streams
STDOUT_FILE=$(mktemp)
STDERR_FILE=$(mktemp)
trap 'rm -f "$STDOUT_FILE" "$STDERR_FILE"' EXIT

./bin/archon prompt "Hello" > "$STDOUT_FILE" 2> "$STDERR_FILE"

# Verify stdout has the exact response
STDOUT_CONTENT=$(cat "$STDOUT_FILE")
if [ "$STDOUT_CONTENT" = "$MOCK_VAL" ]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (stdout did not match mock value)${NC}"
    echo "Expected: $MOCK_VAL"
    echo "Got: $STDOUT_CONTENT"
    exit 1
fi

# Verify stderr has the thinking message
STDERR_CONTENT=$(cat "$STDERR_FILE")
if [[ "$STDERR_CONTENT" == *"[ Archon ] Thinking..."* ]]; then
    echo -e "${GREEN}PASS (stderr captured thinking)${NC}"
else
    echo -e "${RED}FAIL (stderr missing thinking status)${NC}"
    exit 1
fi
unset ARCHON_MOCK_RESPONSE

# 4. Test Live Prompt Command (Verify connection works E2E)
echo "Testing live E2E prompt with the configured cloud model..."
LIVE_OUT=$(./bin/archon prompt "What is the most critical architectural decision when designing a microservices system?" 2>/dev/null)
if [ -n "$LIVE_OUT" ] && [[ "$LIVE_OUT" != *"Error"* ]]; then
    echo -e "${GREEN}PASS (Live prompt succeeded and returned response)${NC}"
else
    echo -e "${RED}FAIL (Live prompt failed or returned empty)${NC}"
    echo "Output: $LIVE_OUT"
    exit 1
fi

echo -e "\n${GREEN}All E2E Integration tests passed successfully! Archon is ready for production use!${NC}"
