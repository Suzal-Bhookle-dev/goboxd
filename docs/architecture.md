# Architecture Design

This document describes the architectural design of **goboxd**, a high-performance Remote Code Execution (RCE) engine.

## High-Level Overview

goboxd is designed as a stateless HTTP service that acts as an orchestrator for **Nsjail**, a light-weight process isolation tool. The system is built using Go and the **Fiber v3** web framework, following the Standard Go Project Layout.

### The Component Stack
1.  **API Layer (Fiber v3):** Handles HTTP routing, JSON binding, and strict input validation.
2.  **Configuration (YAML):** A plug-and-play registry that defines how different languages should be built and executed.
3.  **Execution Engine:** Coordinates the lifecycle of a code submission (Workspace -> Build -> Test).
4.  **Sandbox (Nsjail):** Enforces kernel-level isolation using Linux Namespaces and Control Groups (cgroups).

---

## Request Lifecycle

When a `POST /run` request is received, it follows this pipeline:

### 1. Validation & Binding
*   The request body is bound to a typed struct using Fiber v3's `c.Bind().JSON()`.
*   The **Validator** performs a "fail-fast" check:
    *   Ensures the language ID exists in the registry.
    *   Sanitizes filenames to prevent path traversal.
    *   Checks user-supplied compiler flags against a per-language allow-list.

### 2. Workspace Initialization
*   The engine creates a unique, isolated temporary directory on the host (typically in `/tmp/gobox-*`).
*   The user's source code is written to this directory using the filename strategy defined in the config (e.g., `solution.cpp` or a user-supplied name for Java).

### 3. Build Phase (Conditional)
*   If the language config includes a `build` block (e.g., C++, Java), the engine triggers a compilation step.
*   The compiler is itself run inside an Nsjail sandbox to prevent host exploitation during build-time.
*   **Short-circuit logic:** If compilation fails (exit code != 0), the engine immediately returns a `build_failed` status and skips all tests.

### 4. Execution & Test Loop
*   The engine iterates through the provided `tests` array.
*   For each test case:
    *   A fresh Nsjail process is spawned.
    *   The `stdin` is piped into the sandbox.
    *   The process is monitored for time and memory limits.
    *   The output is captured, truncated if it exceeds 1MiB, and compared against the `expected_stdout`.

### 5. Status Aggregation
*   The engine implements a **priority-based status model**.
*   The top-level status is determined by the first failure encountered in the test sequence (e.g., `wrong_output` takes priority over subsequent `time_exceeded` results).

### 6. Cleanup
*   A `defer` block ensures the temporary workspace is deleted from the host filesystem immediately after the response is sent.
*   On server boot, a garbage collection sweep removes any orphaned directories from previous system crashes.

---

## Technical Design Decisions

### Plug-and-Play Language Logic
The Go source code contains zero language-specific logic (no `switch language == "python"` blocks). Instead, it uses a template-driven approach where the execution command and arguments are derived from `languages.yaml`. This allows new languages to be added in seconds without recompiling the Go binary.

### Concurrency & Isolation
*   **Unique UIDs:** To prevent concurrent processes from interacting, the engine uses an atomic counter to assign every request a unique Linux UID/GID (ranging from 10,000 to 60,000).
*   **Private Namespaces:** Each Nsjail sandbox is granted its own PID, UTS, Network, and IPC namespaces.

### Memory & Performance
*   **Zero-Copy Truncation:** We use byte-buffers for output capture and truncate strings before they are returned to the API layer to prevent memory exhaustion (OOM) on the host.
*   **Fiber v3:** Chosen for its extremely low allocation overhead and high-performance routing compared to the standard `net/http` library.

---

## Directory Structure Roles

*   `cmd/goboxd/`: The application entry point. Handles config loading and graceful shutdowns.
*   `internal/api/`: The web layer. Includes handlers, the router, and the contract validator.
*   `internal/config/`: The data layer. Manages the parsing and retrieval of language definitions.
*   `internal/models/`: The contract layer. Defines the Request/Response DTOs (Data Transfer Objects).
*   `internal/sandbox/`: The core logic. `engine.go` orchestrates the flow, while `nsjail.go` handles the low-level CLI execution.