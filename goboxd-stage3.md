This morning you proved your sandbox can run many languages. This afternoon we find out how many requests it can run at once before it falls over.

We hand every team the same Java program. It is a memory-heavy workload: each run allocates a large block of heap, touches it, does a little compute, and holds it for a moment. One run is harmless. A hundred runs landing at the same time is a different story. Your job is to push your own goboxd service with this program under rising concurrent load, find the exact point where it starts failing, and show us the curve.

## The setup

You run the provided `MemoryHog.java` (full source at the bottom of this page) through your own goboxd service, by sending it to `POST /run` over and over, concurrently, at a controlled request rate.

Run your service inside its container with a fixed resource cap so everyone's numbers mean the same thing and the evaluation can be fair:

- **2 vCPU**
- **2 GB RAM**

Per-request timeout is **10 seconds**. A request that takes longer than 10s counts as a failed request, the same as a non-2xx response.

## What you do

Drive load in steps. Hold each target request rate for **30 seconds**, record the result as one row, then step up. A reasonable ladder:

```
5, 10, 25, 50, 75, 100, 150, 200, 300, 400 requests/sec
```

Keep climbing until you hit the **first failed request**, then run two or three steps past that so the shape of the curve after the break is visible. Every step produces one row in your CSV.

## The breaking point

Your breaking point is the offered request rate at which the **first failed request** appears (a non-2xx response or a request that exceeds the 10s timeout). Mark it clearly on your graph. This single number is the headline result of the challenge.

## What to submit

Everything goes in `docs/loadtest/` on your team branch, plotted with the CSV as proof:

| File | What it is |
|---|---|
| `results.csv` | One row per load step, schema below |
| `breaking-point.png` | Offered RPS on x, error rate percent on y, breaking point marked |
| `latency.png` | Offered RPS on x, latency on y, three lines: p50, p95, p99 |
| `load-test.<sh/js>` | The exact script you ran, so the run is reproducible |
| `README.md` | Container limits used, the tool you used, your breaking-point RPS, what failed first (memory, concurrency limiter, GC, queue rejection), and how to reproduce |

### CSV schema

```
target_rps,throughput_rps,duration_s,requests,success,failed,error_pct,p50_ms,p95_ms,p99_ms,max_ms
```

### The two graphs

- **breaking-point.png** answers "at what load does it break." Offered RPS across the bottom, error rate percent up the side. It should sit at zero, then climb once you pass the breaking point. Draw a line at the breaking-point RPS.
- **latency.png** is the reactivity curve. Offered RPS across the bottom, response latency up the side, one line each for p50, p95, and p99. This shows how the service feels as it loads up: flat while healthy, bending upward as it saturates.

## Tooling

Use whatever load generator you like (k6, vegeta, hey, wrk, or your own), as long as it produces the CSV columns above and you commit the script. A vegeta starter:

```bash
# target.txt: one POST to your /run endpoint.
# Put the goboxd run-request body (language=java, source=MemoryHog) in run-request.json,
# matching the API contract from the spec.
#
#   POST http://localhost:8080/run
#   Content-Type: application/json
#   @run-request.json

for rate in 5 10 25 50 75 100 150 200 300 400; do
  vegeta attack -rate=${rate}/1s -duration=30s -timeout=10s -targets=target.txt \
    | vegeta report -type=json > report-${rate}.json
done
```

Turn each report into a CSV row:

```bash
echo "target_rps,throughput_rps,duration_s,requests,success,failed,error_pct,p50_ms,p95_ms,p99_ms,max_ms" > results.csv
for rate in 5 10 25 50 75 100 150 200 300 400; do
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
done
```

Plot both graphs from that CSV (matplotlib starter):

```python
import csv, matplotlib.pyplot as plt

rows = list(csv.DictReader(open("results.csv")))
rps  = [float(r["target_rps"]) for r in rows]

plt.figure(); plt.plot(rps, [float(r["error_pct"]) for r in rows], marker="o")
plt.xlabel("Offered RPS"); plt.ylabel("Error rate (%)"); plt.title("Breaking point")
plt.savefig("breaking-point.png", dpi=150, bbox_inches="tight")

plt.figure()
for k, lbl in [("p50_ms","p50"), ("p95_ms","p95"), ("p99_ms","p99")]:
    plt.plot(rps, [float(r[k]) for r in rows], marker="o", label=lbl)
plt.xlabel("Offered RPS"); plt.ylabel("Latency (ms)"); plt.title("RPS vs latency"); plt.legend()
plt.savefig("latency.png", dpi=150, bbox_inches="tight")
```

## The demo

Your team presents the load test live to the SEEK team. Have your CSV and both graphs open. Walk us through the curve, point to your breaking-point RPS, and explain what gave out first: did you run out of memory, did your concurrency limiter start queuing and time out, did the JVM processes pile up. Tell us whether the service **degraded gracefully**, returning clean errors or queuing and then recovering once load dropped, or whether it hard-crashed and stayed down. Graceful behaviour under overload is worth as much as a high number.

## Scoring

| Component | Points |
|---|---|
| Reproducible run: container limits documented, load-test script committed | 10 |
| Breaking point correctly identified and matching the CSV | 15 |
| Both graphs correct and consistent with the CSV | 15 |
| Graceful degradation under overload: clean errors or queuing, service recovers | 10 |
| Live demo: clear walkthrough and a correct read of the failure mode | 10 |

A graph that does not match the committed CSV scores nothing for that graph. Numbers without the CSV to back them do not count.

## Timeline

| Time | What happens |
|---|---|
| 14:00 | Kickoff, program and brief handed out |
| 14:15 | Load testing begins |
| 17:00 | Wrap up, commit results to `docs/loadtest/` |
| 17:00 to 18:00 | Live demos to the SEEK team |
| 18:00 | Hard stop |

## The program

Save this as `MemoryHog.java` and send it as the source in your `/run` requests. Do not modify it. Every team runs the same workload.

```java
public class MemoryHog {
    public static void main(String[] args) throws InterruptedException {
        final int megabytes = readSize();
        final int blockSize = 1 << 20; // 1 MB per block
        final byte[][] blocks = new byte[megabytes][];

        long checksum = 0;

        // Allocate and touch every page so the pages are actually committed to RSS.
        for (int i = 0; i < megabytes; i++) {
            byte[] block = new byte[blockSize];
            for (int j = 0; j < blockSize; j += 4096) {
                block[j] = (byte) (i * 31 + j);
                checksum += block[j];
            }
            blocks[i] = block;
        }

        // Light CPU pass so a run is not pure allocation.
        for (byte[] block : blocks) {
            for (int j = 0; j < block.length; j += 512) {
                checksum += block[j];
            }
        }

        // Hold the memory resident for a moment so concurrent runs pile up.
        Thread.sleep(1000);

        // Deterministic output proves the run actually completed.
        System.out.println("MemoryHog OK mb=" + megabytes + " checksum=" + checksum);
    }

    private static int readSize() {
        String env = System.getenv("MEMHOG_MB");
        if (env != null && !env.isEmpty()) {
            try {
                return Integer.parseInt(env.trim());
            } catch (NumberFormatException ignored) {
            }
        }
        return 150;
    }
}
```

Good luck. Push it until it breaks, then show us exactly where.
