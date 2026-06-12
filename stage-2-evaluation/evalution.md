# Stage 2 Evaluation Results

This document records the requests sent to and the responses received from the `/run` endpoint for languages `go`, `perl`, and `ada` during Stage 2 evaluation.

## Table of Contents
- [GO](#go-evaluation)
  - [GO - accepted](#go-accepted)
  - [GO - build_failed](#go-build_failed)
  - [GO - runtime_error](#go-runtime_error)
  - [GO - time_exceeded](#go-time_exceeded)
  - [GO - wrong_output](#go-wrong_output)
- [PERL](#perl-evaluation)
  - [PERL - accepted](#perl-accepted)
  - [PERL - runtime_error](#perl-runtime_error)
  - [PERL - time_exceeded](#perl-time_exceeded)
  - [PERL - wrong_output](#perl-wrong_output)
- [ADA](#ada-evaluation)
  - [ADA - accepted](#ada-accepted)
  - [ADA - build_failed](#ada-build_failed)
  - [ADA - runtime_error](#ada-runtime_error)
  - [ADA - time_exceeded](#ada-time_exceeded)
  - [ADA - wrong_output](#ada-wrong_output)

## GO Evaluation

### GO - accepted

**Payload File:** `payloads/go/accepted.json`  
**Execution Status:** `accepted`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "go",
  "source": "package main\nimport (\"bufio\"; \"fmt\"; \"os\")\nfunc main() {\n    r := bufio.NewReader(os.Stdin)\n    var n int\n    fmt.Fscan(r, &n)\n    fmt.Println(n * 2)\n}",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "accepted",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "",
    "duration_ms": 379
  },
  "tests": [
    {
      "status": "accepted",
      "stdout": "42\n",
      "stderr": "",
      "duration_ms": 6,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### GO - build_failed

**Payload File:** `payloads/go/build_failed.json`  
**Execution Status:** `build_failed`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "go",
  "source": "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"x\"",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "build_failed",
  "build": {
    "status": "failed",
    "stdout": "",
    "stderr": "# command-line-arguments\n./main.go:3:30: syntax error: unexpected EOF in argument list; possibly missing comma or )\n",
    "duration_ms": 102
  },
  "tests": [
    {
      "status": "not_executed",
      "stdout": "",
      "stderr": "",
      "duration_ms": 0,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### GO - runtime_error

**Payload File:** `payloads/go/runtime_error.json`  
**Execution Status:** `runtime_error`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "go",
  "source": "package main\nfunc main() { panic(\"boom\") }",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "runtime_error",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "",
    "duration_ms": 102
  },
  "tests": [
    {
      "status": "runtime_error",
      "stdout": "",
      "stderr": "panic: boom\n\ngoroutine 1 [running]:\nmain.main()\n\t/run/main.go:2 +0x27\n",
      "duration_ms": 8,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### GO - time_exceeded

**Payload File:** `payloads/go/time_exceeded.json`  
**Execution Status:** `time_exceeded`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "go",
  "source": "package main\nfunc main() { for {} }",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ],
  "run": {
    "limits": {
      "wall_time_s": 1
    }
  }
}
```

</details>

#### Response
```json
{
  "status": "time_exceeded",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "",
    "duration_ms": 100
  },
  "tests": [
    {
      "status": "time_exceeded",
      "stdout": "",
      "stderr": "",
      "duration_ms": 1004,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### GO - wrong_output

**Payload File:** `payloads/go/wrong_output.json`  
**Execution Status:** `wrong_output`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "go",
  "source": "package main\nimport (\"bufio\"; \"fmt\"; \"os\")\nfunc main() {\n    r := bufio.NewReader(os.Stdin)\n    var n int\n    fmt.Fscan(r, &n)\n    fmt.Println(n * 2)\n}",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "43\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "wrong_output",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "",
    "duration_ms": 200
  },
  "tests": [
    {
      "status": "wrong_output",
      "stdout": "42\n",
      "stderr": "",
      "duration_ms": 6,
      "memory_peak_kb": 0
    }
  ]
}
```

---

## PERL Evaluation

### PERL - accepted

**Payload File:** `payloads/perl/accepted.json`  
**Execution Status:** `accepted`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "perl",
  "source": "my $n = <STDIN>;\nchomp $n;\nprint $n * 2, \"\\n\";",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "accepted",
  "tests": [
    {
      "status": "accepted",
      "stdout": "42\n",
      "stderr": "",
      "duration_ms": 16,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### PERL - runtime_error

**Payload File:** `payloads/perl/runtime_error.json`  
**Execution Status:** `runtime_error`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "perl",
  "source": "die \"boom\\n\";",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "runtime_error",
  "tests": [
    {
      "status": "runtime_error",
      "stdout": "",
      "stderr": "boom\n",
      "duration_ms": 7,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### PERL - time_exceeded

**Payload File:** `payloads/perl/time_exceeded.json`  
**Execution Status:** `time_exceeded`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "perl",
  "source": "while (1) {}",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ],
  "run": {
    "limits": {
      "wall_time_s": 1
    }
  }
}
```

</details>

#### Response
```json
{
  "status": "time_exceeded",
  "tests": [
    {
      "status": "time_exceeded",
      "stdout": "",
      "stderr": "",
      "duration_ms": 1003,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### PERL - wrong_output

**Payload File:** `payloads/perl/wrong_output.json`  
**Execution Status:** `wrong_output`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "perl",
  "source": "my $n = <STDIN>;\nchomp $n;\nprint $n * 2, \"\\n\";",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "43\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "wrong_output",
  "tests": [
    {
      "status": "wrong_output",
      "stdout": "42\n",
      "stderr": "",
      "duration_ms": 7,
      "memory_peak_kb": 0
    }
  ]
}
```

---

## ADA Evaluation

### ADA - accepted

**Payload File:** `payloads/ada/accepted.json`  
**Execution Status:** `accepted`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "ada",
  "source": "with Ada.Text_IO; use Ada.Text_IO;\nwith Ada.Integer_Text_IO; use Ada.Integer_Text_IO;\nprocedure Main is\n  N : Integer;\nbegin\n  Get(N);\n  Put(N * 2, Width => 0);\n  New_Line;\nend Main;",
  "source_filename": "main.adb",
  "artifact_filename": "main",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "accepted",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "x86_64-linux-gnu-gcc-12 -c solution.adb\nsolution.adb:3:11: warning: file name does not match unit name, should be \"main.adb\" [enabled by default]\nx86_64-linux-gnu-gnatbind-12 -x solution.ali\nx86_64-linux-gnu-gnatlink-12 solution.ali -o solution\n",
    "duration_ms": 204
  },
  "tests": [
    {
      "status": "accepted",
      "stdout": "42\n",
      "stderr": "",
      "duration_ms": 6,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### ADA - build_failed

**Payload File:** `payloads/ada/build_failed.json`  
**Execution Status:** `build_failed`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "ada",
  "source": "procedure Main is\nbegin\n  null\nend Main;",
  "source_filename": "main.adb",
  "artifact_filename": "main",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "build_failed",
  "build": {
    "status": "failed",
    "stdout": "",
    "stderr": "x86_64-linux-gnu-gcc-12 -c solution.adb\nsolution.adb:1:11: warning: file name does not match unit name, should be \"main.adb\" [enabled by default]\nsolution.adb:3:07: error: missing \";\"\ngnatmake: \"solution.adb\" compilation error\n",
    "duration_ms": 13
  },
  "tests": [
    {
      "status": "not_executed",
      "stdout": "",
      "stderr": "",
      "duration_ms": 0,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### ADA - runtime_error

**Payload File:** `payloads/ada/runtime_error.json`  
**Execution Status:** `runtime_error`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "ada",
  "source": "with Ada.Text_IO; use Ada.Text_IO;\nprocedure Main is\nbegin\n  raise Program_Error;\nend Main;",
  "source_filename": "main.adb",
  "artifact_filename": "main",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "runtime_error",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "x86_64-linux-gnu-gcc-12 -c solution.adb\nsolution.adb:2:11: warning: file name does not match unit name, should be \"main.adb\" [enabled by default]\nx86_64-linux-gnu-gnatbind-12 -x solution.ali\nx86_64-linux-gnu-gnatlink-12 solution.ali -o solution\n",
    "duration_ms": 123
  },
  "tests": [
    {
      "status": "runtime_error",
      "stdout": "",
      "stderr": "\nraised PROGRAM_ERROR : solution.adb:4 explicit raise\n",
      "duration_ms": 6,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### ADA - time_exceeded

**Payload File:** `payloads/ada/time_exceeded.json`  
**Execution Status:** `time_exceeded`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "ada",
  "source": "procedure Main is\nbegin\n  loop\n    null;\n  end loop;\nend Main;",
  "source_filename": "main.adb",
  "artifact_filename": "main",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "42\n"
    }
  ],
  "run": {
    "limits": {
      "wall_time_s": 1
    }
  }
}
```

</details>

#### Response
```json
{
  "status": "time_exceeded",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "x86_64-linux-gnu-gcc-12 -c solution.adb\nsolution.adb:1:11: warning: file name does not match unit name, should be \"main.adb\" [enabled by default]\nx86_64-linux-gnu-gnatbind-12 -x solution.ali\nx86_64-linux-gnu-gnatlink-12 solution.ali -o solution\n",
    "duration_ms": 98
  },
  "tests": [
    {
      "status": "time_exceeded",
      "stdout": "",
      "stderr": "",
      "duration_ms": 1004,
      "memory_peak_kb": 0
    }
  ]
}
```

---

### ADA - wrong_output

**Payload File:** `payloads/ada/wrong_output.json`  
**Execution Status:** `wrong_output`

<details>
<summary>View Request Payload</summary>

```json
{
  "language": "ada",
  "source": "with Ada.Text_IO; use Ada.Text_IO;\nwith Ada.Integer_Text_IO; use Ada.Integer_Text_IO;\nprocedure Main is\n  N : Integer;\nbegin\n  Get(N);\n  Put(N * 2, Width => 0);\n  New_Line;\nend Main;",
  "source_filename": "main.adb",
  "artifact_filename": "main",
  "tests": [
    {
      "stdin": "21\n",
      "expected_stdout": "43\n"
    }
  ]
}
```

</details>

#### Response
```json
{
  "status": "wrong_output",
  "build": {
    "status": "ok",
    "stdout": "",
    "stderr": "x86_64-linux-gnu-gcc-12 -c solution.adb\nsolution.adb:3:11: warning: file name does not match unit name, should be \"main.adb\" [enabled by default]\nx86_64-linux-gnu-gnatbind-12 -x solution.ali\nx86_64-linux-gnu-gnatlink-12 solution.ali -o solution\n",
    "duration_ms": 129
  },
  "tests": [
    {
      "status": "wrong_output",
      "stdout": "42\n",
      "stderr": "",
      "duration_ms": 8,
      "memory_peak_kb": 0
    }
  ]
}
```

---
