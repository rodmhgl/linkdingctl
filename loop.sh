#!/bin/bash
# Ralph Loop for LinkDing CLI
# Usage: ./loop.sh [plan|build]

set -e

MODE="${1:-build}"
MAX_ITERATIONS=100
ITERATION=0

echo "================================"
echo "Ralph Loop - LinkDing CLI"
echo "Mode: $MODE"
echo "Max iterations: $MAX_ITERATIONS"
echo "================================"

# Ensure we're in the project root
cd "$(dirname "$0")"

# Initialize git if needed
if [ ! -d ".git" ]; then
    git init
    git add -A
    git commit -m "Initial project structure"
fi

case "$MODE" in
    plan)
        echo "Running planning mode (single iteration)..."
        cat PROMPT_plan.md | claude -p --dangerously-skip-permissions
        echo "Planning complete. Review IMPLEMENTATION_PLAN.md"
        exit 0
        ;;
    build)
        echo "Starting build loop..."
        while [ $ITERATION -lt $MAX_ITERATIONS ]; do
            ITERATION=$((ITERATION + 1))
            echo ""
            echo "======== ITERATION $ITERATION / $MAX_ITERATIONS ========"
            echo "Time: $(date)"
            echo ""
            
            OUTPUT=$(cat PROMPT_build.md | claude -p \
                --dangerously-skip-permissions \
                --model sonnet \
                2>&1)
            
            echo "$OUTPUT"
            
            # Check for completion
            if echo "$OUTPUT" | grep -q "<promise>COMPLETE</promise>"; then
                echo ""
                echo "========================================"
                echo "BUILD COMPLETE after $ITERATION iterations!"
                echo "========================================"
                exit 0
            fi
            
            # Brief pause to avoid rate limiting
            sleep 2
        done
        
        echo ""
        echo "========================================"
        echo "MAX ITERATIONS ($MAX_ITERATIONS) reached"
        echo "Review IMPLEMENTATION_PLAN.md for status"
        echo "========================================"
        exit 1
        ;;
    *)
        echo "Unknown mode: $MODE"
        echo "Usage: ./loop.sh [plan|build]"
        exit 2
        ;;
esac
