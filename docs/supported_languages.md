# Supported Languages

**goboxd** uses a plug-and-play language registry defined in `languages.yaml`. This allows for seamless addition of new runtimes and compilers without modification to the core Go source code.

## Language Overview

| ID | Language | Type | Compiler/Runtime | Default Time Limit | Default Memory |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `c` | C | Compiled | GCC | 3s | 512 MB |
| `cpp` | C++ | Compiled | G++ | 3s | 512 MB |
| `rust` | Rust | Compiled | rustc | 3s | 128 MB |
| `java` | Java | Compiled | javac / OpenJDK | 5s | 4 GB |
| `py3` | Python 3 | Interpreted | python3 | 9s | 100 MB |
| `js` | JavaScript | Interpreted | Node.js | 5s | 4 GB |
| `verilog` | Verilog | HDL | Icarus Verilog | 5s | 256 MB |

---

## Detailed Specifications

### Compiled Languages (C, C++, Rust)
These languages undergo a build phase before execution. 
*   **Artifacts:** The resulting binary is stored as `solution` and executed directly.
*   **Optimization:** Support for optimization flags (e.g., `-O2`, `-O3`) is provided via the `flag_allowlist`.

### Java
Java uses a dynamic filename strategy to comply with JVM requirements.
*   **Class Names:** The `source_filename` must match the `public class` name defined in the source code (e.g., `Main.java` for `public class Main`).
*   **Memory:** Due to JVM memory reservation patterns, Java requests are allocated a larger virtual address space (4GB) to ensure successful initialization of the class space.

### JavaScript (Node.js)
*   **Runtime:** Node.js executes the source directly.
*   **Memory:** Like Java, Node.js requires a significant virtual memory allocation (4GB) for the V8 engine's code range and heap management.

### Verilog
*   **Workflow:** Uses a two-step pipeline. The source is compiled using `iverilog` into a `.vvp` artifact, which is then executed by the `vvp` runtime.
*   **Output:** Captures standard `$display` and `$monitor` output from the simulation.

---

## Resource Limits

Limits can be defined per-language in the configuration or overridden per-request via the API.

1.  **Wall Time (`wall_time_s`):** The maximum real-world time the process is allowed to run before being killed by Nsjail.
2.  **Memory (`memory_kb`):** Strict RLIMIT_AS (Address Space) limit. Note that runtimes with heavy JIT compilers (Java/Node) require higher values here.
3.  **Process Limit (`max_processes`):** Prevents fork-bombs by limiting the maximum number of PIDs within the sandbox.
4.  **File Size (`rlimit_fsize`):** Restricted to 16MB across all languages to prevent disk-fill attacks.

---

## Adding a New Language

New languages can be added to the `languages.yaml` file using the following schema:

```yaml
- id: <unique_id>
  name: <display_name>
  source_filename: <default_name>
  artifact: <compiled_binary_name>
  build: # Optional
    cmd: <compiler_path>
    args: ["{{flags}}", "-o", "{{artifact}}", "{{source}}"]
    limits: { wall_time_s: 10, memory_kb: 1048576, max_processes: 100 }
    flag_allowlist: ["-O2", "-Wall"]
  run:
    cmd: <runtime_path>
    args: ["{{source}}"]
    limits: { wall_time_s: 5, memory_kb: 256000, max_processes: 64 }