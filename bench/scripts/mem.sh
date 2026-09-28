#!/usr/bin/env bash
# Measure container memory of the scheduler probe at a few target counts.
# This is an exploration tool, not a benchmark suite: it settles for a fixed
# number of seconds and reports `docker stats`, not a steady-state profile.
set -euo pipefail
cd "$(dirname "$0")/.."

targets=${TARGETS:-"0 1000 5000"}
interval=${INTERVAL:-60s}
settle=${SETTLE:-25}

docker build -q -t octopulse-probe-sched -f Dockerfile.sched . >/dev/null
for n in $targets; do
  docker rm -f "octopulse-sched-$n" >/dev/null 2>&1 || true
  docker run -d --name "octopulse-sched-$n" --cpus=1 --memory=512m --network=host \
    octopulse-probe-sched -targets "$n" -interval "$interval" >/dev/null
  sleep "$settle"
  printf 'targets=%-6s %s\n' "$n" "$(docker stats "octopulse-sched-$n" --no-stream --format 'Mem={{.MemUsage}} CPU={{.CPUPerc}}')"
  docker logs "octopulse-sched-$n" 2>&1 | tail -1
  docker rm -f "octopulse-sched-$n" >/dev/null
done
