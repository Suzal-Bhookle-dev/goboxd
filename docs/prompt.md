# Technical Consultation Log

This document summarizes the technical queries and implementation discussions held during the development of **goboxd**. It serves as a record of the design patterns chosen and the specific technical hurdles overcome during the build process.

## 1. Architectural Design
**Context:** Initiating the project and defining boundaries between the API and the sandbox.
*   **Query:** "I am designing a Remote Code Execution service using Go and Nsjail. How should I structure the project using the Standard Go Project Layout to ensure the sandbox logic is isolated from the HTTP layer? I want to keep the core engine in `internal/sandbox`."

## 2. Framework Selection & Fiber v3
**Context:** Choosing between the standard library and high-performance frameworks.
*   **Query:** "I've decided to use the Fiber v3 beta for this project. What are the best practices for handling the new `Bind().JSON()` API to ensure strict compliance with a predefined JSON contract for a `/run` endpoint?"

## 3. Generic Language Registry (Plug-and-Play)
**Context:** Moving away from hardcoded language logic to a YAML-driven system.
*   **Query:** "I need to implement a 'Plug-and-Play' language registry. I want to define languages in `languages.yaml` including optional build steps and specific filename strategies (like Java's public class requirement). How do I build a generic execution engine in Go that uses string templates for commands and arguments?"

## 4. Nsjail Environment Troubleshooting
**Context:** Resolving low-level sandbox errors during the compilation of C++ and Rust.
*   **Query:** "My Nsjail wrapper is returning 'Read-only file system' when running `g++`, even though I have the source folder bound. How do I properly configure `--cwd` and mount a `tmpfs` to `/tmp` so the compiler can write intermediate objects? Also, what mounts are required for the linker (`ld`) to find system libraries?"

## 5. Security: UID Isolation & Atomic Counters
**Context:** Hardening the sandbox against concurrent process interference.
*   **Query:** "To close the 'UID Collision' security hole, I want every concurrent request to run as a unique Linux user. What is a thread-safe way to rotate through a pool of 50,000 UIDs in Go using the `sync/atomic` package?"

## 6. Output Handling & OOM Prevention
**Context:** Managing runaway programs and memory exhaustion.
*   **Query:** "How should I handle child process output to prevent a user program from flooding the host memory? I need a logic that captures `stdout` but truncates it at 1MB and appends a truncation marker if the limit is exceeded."

## 7. Status Aggregation Logic
**Context:** Implementing the prioritized "Status Vocabulary" required by the contract.
*   **Query:** "The API requires a specific top-level status based on multiple test cases. If a build fails, the tests should be marked as `not_executed`. If tests run, the first failure (like `wrong_output` or `time_exceeded`) must define the top-level status. How do I implement this short-circuit logic cleanly in the execution loop?"

## 8. Linker Optimization (Rust & C++)
**Context:** Debugging 'File size limit exceeded' during the Rust linking phase.
*   **Query:** "My Rust builds are failing with Signal 25 in the sandbox. It seems to be hitting `rlimit_fsize`. What is a safe but sufficient limit to allow modern compiled binaries (which are larger than C binaries) to link successfully while still preventing disk-fill attacks?"

## 9. Docker Multi-Stage Optimization
**Context:** Creating a lean production image that still supports multiple compilers.
*   **Query:** "I am using a multi-stage Dockerfile. How do I structure the `runtime` stage so it only contains the `goboxd` binary and `nsjail`, but also includes the necessary runtimes (Python, Node, GCC) without bloating the image unnecessarily?"

## 10. Integration Testing with Build Tags
**Context:** Verifying the full pipeline across all 7 supported languages.
*   **Query:** "I want to write end-to-end integration tests using `app.Test`. Since these tests require Nsjail and system compilers, how do I use Go build tags (e.g., `//go:build integration`) to ensure they only run within the containerized environment?"