#!/bin/bash

# Autonomous Trading Bot Launcher
# Starts the Go trading backend and loops Claude sessions until market close

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
MAX_BUDGET_USD="${MAX_BUDGET_USD:-5.00}"
STOP_TIME="${STOP_TIME:-16:01}"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Prophet Autonomous Trading System${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "  Stop time:  ${STOP_TIME}"
echo -e "  Budget cap: \$${MAX_BUDGET_USD} per session"
echo ""

# Check if trading bot is already running
if lsof -Pi :4534 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Trading bot already running on port 4534${NC}"
else
    echo -e "${YELLOW}Starting Go trading bot...${NC}"

    # Load environment variables
    if [ -f .env ]; then
        export $(cat .env | grep -v '^#' | xargs)
    fi

    # Start the trading bot in background (use binary for speed)
    ALPACA_API_KEY=${ALPACA_API_KEY:-$ALPACA_PUBLIC_KEY} \
    ALPACA_SECRET_KEY=${ALPACA_SECRET_KEY} \
    nohup ./prophet_bot > trading_bot.log 2>&1 &

    echo $! > trading_bot.pid

    # Wait for bot to start
    echo -e "${YELLOW}Waiting for trading bot to initialize...${NC}"
    sleep 5

    # Verify it's running
    if lsof -Pi :4534 -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Trading bot started successfully (PID: $(cat trading_bot.pid))${NC}"
    else
        echo -e "${RED}✗ Failed to start trading bot. Check trading_bot.log${NC}"
        exit 1
    fi
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}System Ready for Autonomous Trading${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Portfolio Status:"
echo "  • Cash: $(curl -s http://localhost:4534/api/v1/account | grep -o '"Cash":[0-9.]*' | cut -d: -f2 || echo 'N/A')"
echo "  • Buying Power: $(curl -s http://localhost:4534/api/v1/account | grep -o '"BuyingPower":[0-9.]*' | cut -d: -f2 || echo 'N/A')"
echo ""

# Loop Claude sessions until STOP_TIME
SESSION_ID=""
SESSION_NUM=0
TOTAL_COST=0

while true; do
    # Check if past stop time
    NOW=$(date +%H:%M)
    if [[ "$NOW" > "$STOP_TIME" || "$NOW" == "$STOP_TIME" ]]; then
        echo -e "${GREEN}Market close time reached ($STOP_TIME). Stopping.${NC}"
        break
    fi

    SESSION_NUM=$((SESSION_NUM + 1))
    LOG_FILE="autonomous_session_$(date +%Y%m%d_%H%M%S).log"

    echo -e "${YELLOW}--- Session $SESSION_NUM ($(date +%H:%M:%S)) --- logging to: $LOG_FILE${NC}"

    # Build claude command — resume previous session if we have one
    CLAUDE_ARGS=(
        --print
        --verbose
        --output-format stream-json
        --permission-mode bypassPermissions
        --max-budget-usd "$MAX_BUDGET_USD"
    )

    if [ -n "$SESSION_ID" ]; then
        # Continue from previous session so Claude has context
        CLAUDE_ARGS+=(--resume "$SESSION_ID")
        CLAUDE_ARGS+=("Continue trading. Current time is $(date '+%H:%M %Z'). Keep going until $STOP_TIME.")
    else
        CLAUDE_ARGS+=("$(cat autonomous_trading_prompt.txt)")
    fi

    # Run Claude session (don't exit script on non-zero)
    set +e
    claude "${CLAUDE_ARGS[@]}" | tee "$LOG_FILE"
    EXIT_CODE=$?
    set -e

    # Extract session ID and cost from the result line for resume
    RESULT_LINE=$(grep '"type":"result"' "$LOG_FILE" 2>/dev/null || true)
    if [ -n "$RESULT_LINE" ]; then
        NEW_SESSION_ID=$(echo "$RESULT_LINE" | grep -o '"session_id":"[^"]*"' | head -1 | cut -d'"' -f4)
        SESSION_COST=$(echo "$RESULT_LINE" | grep -o '"total_cost_usd":[0-9.]*' | cut -d: -f2)
        if [ -n "$NEW_SESSION_ID" ]; then
            SESSION_ID="$NEW_SESSION_ID"
        fi
        if [ -n "$SESSION_COST" ]; then
            TOTAL_COST=$(echo "$TOTAL_COST + $SESSION_COST" | bc)
        fi
    fi

    echo -e "${BLUE}Session $SESSION_NUM finished (exit=$EXIT_CODE, cost=\$${SESSION_COST:-unknown}, total=\$${TOTAL_COST})${NC}"

    # Brief pause before restarting to avoid hammering the API
    sleep 5
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Autonomous trading session complete${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "  Sessions run: $SESSION_NUM"
echo "  Total cost:   \$$TOTAL_COST"
echo ""
echo "Trading bot log: tail -f trading_bot.log"
echo "Activity log: cat activity_logs/activity_$(date +%Y-%m-%d).json"
echo ""
echo "To stop trading bot: kill $(cat trading_bot.pid 2>/dev/null || echo 'N/A')"
