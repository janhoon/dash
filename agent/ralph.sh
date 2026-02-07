#! /bin/bash

set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <number of iterations>"
  exit 1
fi

iterations="$1"

if [ -z "$iterations" ]; then
  echo "Usage: $0 <number of iterations>"
  exit 1
fi

for i in $(seq 1 "$iterations"); do
  echo "Iteration $i"
  echo "----------------------------------------"

  result=$(cat prompt.md | claude --permission-mode acceptEdits --dangerously-skip-permissions -p)

  echo "$result"

  # Check for full completion (all features done)
  if [[ "$result" == *"<promise>COMPLETE</promise>"* ]]; then
    echo "All features complete!"
    exit 0
  fi
done

echo "Reached iteration limit. Review progress and continue if needed."
