#!/bin/bash
set -e

export PATH=$PATH:/home/suzal/.local/share/mise/installs/go/1.26.3/bin

cd "$(dirname "$0")"

# Start container
echo "Starting container..."
cd ../../
docker compose down
docker compose up --build -d goboxd
echo "Waiting for goboxd to start..."
sleep 5
cd docs/loadtest

echo "target_rps,throughput_rps,duration_s,requests,success,failed,error_pct,p50_ms,p95_ms,p99_ms,max_ms" > results.csv

for rate in 5 10 25 50 75 100 150 200 300 400; do
  echo "---"
  echo "Running load test at ${rate} RPS for 30s..."
  vegeta attack -rate=${rate}/1s -duration=30s -timeout=10s -targets=targets.txt \
    | vegeta report -type=json > report-${rate}.json

  jq -r --arg r "$rate" '
    [ $r,
      .throughput,
      (.duration/1e9),
      .requests,
      (.status_codes["200"] // 0),
      (.requests - (.status_codes["200"] // 0)),
      ((1 - .success) * 100),
      (.latencies["50th"]/1e6),
      (.latencies["95th"]/1e6),
      (.latencies["99th"]/1e6),
      (.latencies.max/1e6)
    ] | @csv' report-${rate}.json >> results.csv

  ERROR_PCT=$(jq -r '((1 - .success) * 100)' report-${rate}.json)
  echo "Error rate at ${rate} RPS: ${ERROR_PCT}%"
done

echo "Generating plots..."
source ../../venv/bin/activate
python3 plot.py

echo "Cleaning up..."
cd ../../
docker compose down
echo "Done!"
