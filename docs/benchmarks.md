# Performance Benchmarks

This report summarizes the performance of the core sandbox engine logic. Benchmarks were conducted using Go's native testing toolchain on the target host environment.

## Environment
- **CPU:** AMD Ryzen 5 5625U with Radeon Graphics (12 Threads)
- **OS:** Linux (amd64)
- **Go Version:** 1.23+

## Internal Logic Performance
These benchmarks measure the overhead of the execution engine's logic (excluding the latency of the Nsjail process fork).

| Operation | Latency (ns/op) | Throughput (ops/sec) | Memory (B/op) | Allocations |
| :--- | :--- | :--- | :--- | :--- |
| **Output Comparison** | 9.08 ns | ~110,000,000 | 0 B | 0 |
| **Limit Merging** | 0.23 ns | ~4,000,000,000 | 0 B | 0 |
| **Placeholder Logic** | 1,127 ns | ~887,000 | 1,544 B | 27 |

### Analysis
1.  **Zero-Allocation Comparison:** The `compareOutput` logic is highly optimized, resulting in 0 bytes of memory allocation. This ensures that the engine can grade millions of test cases without impacting host memory pressure.
2.  **Near-Instant Limit Merging:** Merging request-specific limits with language defaults happens in less than a nanosecond, adding zero measurable overhead to the request lifecycle.
3.  **String Orchestration:** The `ReplacePlaceholders` logic, which prepares the shell commands for Nsjail, completes in approximately 1 microsecond. While it involves some allocations due to string manipulation, it remains well below the threshold of being a bottleneck for the service.

## Conclusion
The Go management layer is highly efficient. The primary latency bottleneck in the system is restricted to the Linux kernel's process forking speed and the overhead of the compilers/runtimes themselves, rather than the `goboxd` orchestration logic.