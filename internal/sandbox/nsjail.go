package sandbox

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"github.com/thesouldev/goboxd/internal/models"
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	TimedOut bool
	MemoryKb int
}

func runInNsjail(workDir string, lim models.Limits, stdin string, cmdPath string, args ...string) (Result, error) {
	nsjailArgs := []string{
		"-Mo",
		"--chroot", "/",
		"-R", "/usr",
		"-R", "/lib",
		"-R", "/lib64",
		"-R", "/bin",
		"-B", workDir + ":/run",
		"--user", "99999",
		"--group", "99999",
		"--time_limit", strconv.Itoa(lim.WallTimeS),
		"--rlimit_as", strconv.Itoa(lim.MemoryKB / 1024),
		"--max_cpus", "1",
		"--", cmdPath,
	}
	nsjailArgs = append(nsjailArgs, args...)

	var outB, errB bytes.Buffer
	cmd := exec.Command("nsjail", nsjailArgs...)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &outB
	cmd.Stderr = &errB

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return Result{
		Stdout:   outB.String(),
		Stderr:   errB.String(),
		ExitCode: exitCode,
	}, nil
}
