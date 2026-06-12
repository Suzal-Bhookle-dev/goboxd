package sandbox

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/thesouldev/goboxd/internal/models"
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
	MemoryKb int
}

var uidCounter uint32 = 10000

func runInNsjail(workDir string, lim models.Limits, stdin string, cmdPath string, args ...string) (Result, error) {
	uniqueUID := atomic.AddUint32(&uidCounter, 1)
	if uniqueUID > 60000 {
		atomic.StoreUint32(&uidCounter, 10000)
	}
	uidStr := strconv.Itoa(int(uniqueUID))

	nsjailArgs := []string{
		"-Q",
		"-Mo",
		"--user", uidStr,
		"--group", uidStr,
		"--chroot", "/",
		"--cwd", "/run",
		"-E", "PATH=/usr/bin:/bin:/usr/local/bin",
		"-E", "GOCACHE=/tmp/go-cache",
		"-R", "/usr",
		"-R", "/lib",
		"-R", "/lib64",
		"-R", "/bin",
		"-R", "/etc",
		"-R", "/etc/alternatives",
		"-R", "/proc",
		"-m", "none:/tmp:tmpfs:size=67108864",
		"-B", workDir + ":/run",
		"--time_limit", strconv.Itoa(lim.WallTimeS),
		"--rlimit_as", strconv.Itoa(lim.MemoryKB / 1024),
		"--rlimit_fsize", "64",
		"--max_cpus", "1",
		"--", cmdPath,
	}
	nsjailArgs = append(nsjailArgs, args...)

	var outB, errB bytes.Buffer
	cmd := exec.Command("nsjail", nsjailArgs...)

	const MaxOutputSize = 1024 * 1024
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &outB
	cmd.Stderr = &errB

	err := cmd.Run()

	exitCode := 0
	timedOut := false
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			if exitCode == 137 || strings.Contains(err.Error(), "signal: killed") {
				timedOut = true
			}
		} else {
			return Result{}, err
		}
	}

	stdoutStr := outB.String()
	if len(stdoutStr) > MaxOutputSize {
		stdoutStr = stdoutStr[:MaxOutputSize] + "\n[Output Truncated]"
	}

	stderrStr := errB.String()
	if len(stderrStr) > MaxOutputSize {
		stderrStr = stderrStr[:MaxOutputSize] + "\n[Output Truncated]"
	}

	return Result{
		Stdout:   stdoutStr,
		Stderr:   stderrStr,
		ExitCode: exitCode,
		TimedOut: timedOut,
	}, nil
}
