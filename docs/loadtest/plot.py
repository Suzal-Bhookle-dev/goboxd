import csv
import matplotlib.pyplot as plt
import os

def main():
    if not os.path.exists("results.csv"):
        print("results.csv not found")
        return

    rows = list(csv.DictReader(open("results.csv")))
    if not rows:
        print("results.csv is empty")
        return

    rps = [float(r["target_rps"]) for r in rows]
    error_pct = [float(r["error_pct"]) for r in rows]

    # Find breaking point
    breaking_point = None
    for r, e in zip(rps, error_pct):
        if e > 0:
            breaking_point = r
            break

    plt.figure()
    plt.plot(rps, error_pct, marker="o")
    if breaking_point is not None:
        plt.axvline(x=breaking_point, color='r', linestyle='--', label=f'Breaking Point: {breaking_point} RPS')
        plt.legend()
    
    plt.xlabel("Offered RPS")
    plt.ylabel("Error rate (%)")
    plt.title("Breaking point")
    plt.savefig("breaking-point.png", dpi=150, bbox_inches="tight")

    plt.figure()
    for k, lbl in [("p50_ms","p50"), ("p95_ms","p95"), ("p99_ms","p99")]:
        plt.plot(rps, [float(r[k]) for r in rows], marker="o", label=lbl)
    plt.xlabel("Offered RPS")
    plt.ylabel("Latency (ms)")
    plt.title("RPS vs latency")
    plt.legend()
    plt.savefig("latency.png", dpi=150, bbox_inches="tight")

if __name__ == "__main__":
    main()
