# Security Architecture

This document outlines the security measures implemented in **goboxd** to ensure the safe execution of untrusted code. We have addressed all seven security holes identified in the challenge reference.

## 1. Path Traversal via Filename
*   **Vulnerability:** Use of client-supplied filenames to escape the sandbox directory using `../` or absolute paths.
*   **Fix:** Strict filename validation in the API layer.
*   **Implementation:** The validator ensures that the filename is a single path component, contains no directory separators (`/` or `\`), and does not start with a leading dot.
*   **Location:** `internal/api/validator.go` (See `validatePath` function)

## 2. Shell-style Directory Commands
*   **Vulnerability:** Creating or deleting directories by formatting strings into shell commands (e.g., `rm -rf {dir}`).
*   **Fix:** Exclusive use of native Go filesystem APIs.
*   **Implementation:** We use `os.MkdirTemp` for workspace creation and `os.RemoveAll` for cleanup. No strings are ever passed to a shell interpreter (`sh`, `bash`).
*   **Location:** `internal/sandbox/engine.go`

## 3. Compiler-flag Injection
*   **Vulnerability:** Passing arbitrary flags (like `-fplugin=`) to compilers to achieve code execution during the build phase.
*   **Fix:** Per-language strict allow-listing.
*   **Implementation:** All user-supplied flags are filtered against the `flag_allowlist` defined in `languages.yaml`. Any flag not explicitly permitted results in a `400 Disallowed Flag` error.
*   **Location:** `internal/api/validator.go` (See `validateFlags` function)

## 4. No Request Size Limits
*   **Vulnerability:** Massive source code, high test counts, or large stdin causing server resource exhaustion.
*   **Fix:** Multi-layered size constraints.
*   **Implementation:** 
    *   **HTTP Layer:** Total request body is capped at 1MiB via Fiber configuration.
    *   **Logic Layer:** Source code is capped at 256KiB; Stdin is capped at 64KiB per test.
    *   **Sandbox Layer:** Nsjail is restricted with `--rlimit_fsize 1` to prevent user code from writing massive files to disk.
*   **Location:** `internal/api/validator.go` and `internal/sandbox/nsjail.go`

## 5. UID Collisions Under Load
*   **Vulnerability:** Using a static UID for all sandboxes, allowing concurrent processes to potentially interfere with one another.
*   **Fix:** Process-unique UID assignment via atomic counters.
*   **Implementation:** We use a thread-safe `sync/atomic` counter to rotate through a pool of 50,000 unique UIDs (10,000–60,000). Every concurrent request runs as a unique Linux user.
*   **Location:** `internal/sandbox/nsjail.go` (See `uidCounter` logic)

## 6. Unbounded Child Output
*   **Vulnerability:** runaway programs (e.g., `while(1) print('A')`) filling host memory by capturing infinite `stdout`.
*   **Fix:** Mandatory output truncation.
*   **Implementation:** The engine captures output into a buffer but strictly truncates the resulting strings to 1MiB before returning them to the API handler. A `[Output Truncated]` marker is appended when limits are hit.
*   **Location:** `internal/sandbox/nsjail.go`

## 7. Stale Jail Directories
*   **Vulnerability:** Orphaned directories remaining on the host after a service crash or panic.
*   **Fix:** Dual-phase cleanup strategy.
*   **Implementation:**
    *   **Runtime:** `defer os.RemoveAll` handles cleanup for every standard execution path.
    *   **Startup:** A `cleanStaleJails` function runs when the server boots, scanning the temporary directory for any `gobox-*` folders left over from previous crashes.
*   **Location:** `cmd/goboxd/main.go` and `internal/sandbox/engine.go`

---

## Sandbox Isolation (Nsjail)
In addition to the logic fixes above, every execution is isolated using **Nsjail** with the following parameters:
- **Mount Namespace:** Read-only access to `/usr`, `/bin`, `/lib`, and `/etc`.
- **User Namespace:** Mapping to a restricted, non-root UID.
- **PID Namespace:** Prevention of "fork-bombs" via process limits.
- **Cgroups:** Hard limits on CPU and Memory usage.