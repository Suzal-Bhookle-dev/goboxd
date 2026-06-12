# Stage 3 Load Test Results

## Configuration
- **Container Limits:** 2 vCPU, 2 GB RAM (configured in `docker-compose.yml` via `deploy.resources.limits`).
- **Tooling:** `vegeta` was used as the load generator to step through 5, 10, 25, 50, 75, 100, 150, 200, 300, 400 RPS.

## Breaking Point

**Breaking Point RPS:** `< 5 RPS`

The unmodified `goboxd` service broke immediately at the first tier of the ladder (5 RPS).

### What Failed First?
**CPU Saturation / Lack of Concurrency Limiting.**
The `goboxd` service has no internal concurrency limit or request queuing. When Vegeta sends 5 requests per second, the service immediately spawns 5 concurrent `javac` and `java` (JVM) sandbox processes.
- A single request takes ~2s to complete (800ms compilation, 1200ms execution).
- At 5 RPS, the 2 vCPUs become severely saturated almost instantly.
- Processes starve for CPU time, causing response latencies to quickly exceed the hard 10-second Vegeta timeout.
- This results in the requests being marked as "failed" (timeout) by the load generator.

The service **did not degrade gracefully**. It hard-crashed under its own concurrency load, leading to a near 100% timeout error rate right from the beginning, without returning clean HTTP 429 or 503 errors.

## Plots
- `breaking-point.png`: Shows the error rate immediately shooting to ~99% at 5 RPS.
- `latency.png`: Shows the latencies hitting the 10,000ms ceiling (timeout).

## Reproducing the Run
Execute the load test script:
```bash
bash load-test.sh
```
This script will:
1. Rebuild and launch the `goboxd` container.
2. Run the `vegeta` load ladder using the `MemoryHog.json` payload.
3. Output the CSV and plot the graphs using python.
