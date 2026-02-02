#!/bin/bash

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

  # Check for PR created (feature complete, waiting for merge)
  if [[ "$result" == *"<promise>PR_CREATED</promise>"* ]]; then
    echo "PR created, stopping. Merge the PR, then run release.sh to publish the final build."
    exit 0
  fi

  # Check for full completion
  if [[ "$result" == *"<promise>COMPLETE</promise>"* ]]; then
    echo "All features complete!"
    exit 0
  fi
  
  # Check for release completion
  if [[ "$result" == *"<promise>RELEASE_COMPLETE</promise>"* ]]; then
    echo "Release published!"
    exit 0
  fi
done

echo "Reached iteration limit. Review progress and continue if needed."
