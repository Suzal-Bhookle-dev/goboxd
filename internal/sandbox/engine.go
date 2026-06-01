package sandbox

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/models"
)

func Execute(req models.RunRequest, lang config.Language) (models.RunResponse, error) {
	tmpDir, err := os.MkdirTemp("", "gobox-")
	if err != nil {
		return models.RunResponse{}, err
	}
	defer os.RemoveAll(tmpDir)

	srcFile := lang.SourceFilename
	if lang.SourceFilenameStrategy == "from_request" && req.SourceFilename != "" {
		srcFile = req.SourceFilename
	}

	artFile := lang.Artifact
	if lang.ArtifactFilenameStrategy == "from_request" && req.ArtifactFilename != "" {
		artFile = req.ArtifactFilename
	}

	err = os.WriteFile(filepath.Join(tmpDir, srcFile), []byte(req.Source), 0o644)
	if err != nil {
		return models.RunResponse{}, err
	}

	var response models.RunResponse

	if lang.Build != nil {
		start := time.Now()
		uFlags := ""
		if req.Build != nil {
			uFlags = strings.Join(req.Build.Flags, " ")
		}

		buildCmd := replaceString(lang.Build.Cmd, srcFile, artFile, uFlags)
		args := replacePlaceholders(lang.Build.Args, srcFile, artFile, uFlags)

		res, err := runInNsjail(tmpDir, mergeLimits(req.Build, lang.Build.Limits), "", buildCmd, args...)

		bStatus := "ok"
		if err != nil {
			bStatus = "internal_error"
		} else if res.ExitCode != 0 {
			bStatus = "failed"
		}

		response.Build = &models.StepResult{
			Status:     bStatus,
			Stdout:     res.Stdout,
			Stderr:     res.Stderr,
			DurationMs: time.Since(start).Milliseconds(),
		}

		if bStatus != "ok" {
			response.Status = "build_failed"
			for range req.Tests {
				response.Tests = append(response.Tests, models.TestResult{Status: "not_executed"})
			}
			return response, nil
		}
	}

	firstFailureStatus := ""
	for _, tc := range req.Tests {
		start := time.Now()

		runCmd := replaceString(lang.Run.Cmd, srcFile, artFile, "")
		args := replacePlaceholders(lang.Run.Args, srcFile, artFile, "")

		res, err := runInNsjail(tmpDir, mergeLimits(req.Run, lang.Run.Limits), tc.Stdin, runCmd, args...)

		tStatus := "accepted"

		if err != nil {
			tStatus = "internal_error"
		} else if res.TimedOut {
			tStatus = "time_exceeded"
		} else if res.ExitCode != 0 {
			tStatus = "runtime_error"
		} else {
			tStatus = compareOutput(res.Stdout, tc.ExpectedStdout)
		}

		if firstFailureStatus == "" && tStatus != "accepted" {
			firstFailureStatus = tStatus
		}

		response.Tests = append(response.Tests, models.TestResult{
			Status:       tStatus,
			Stdout:       res.Stdout,
			Stderr:       res.Stderr,
			DurationMs:   time.Since(start).Milliseconds(),
			MemoryPeakKb: res.MemoryKb,
		})
	}

	if firstFailureStatus != "" {
		response.Status = firstFailureStatus
	} else {
		response.Status = "accepted"
	}

	return response, nil
}

func compareOutput(actual, expected string) string {
	if actual == expected {
		return "accepted"
	}
	if strings.TrimSpace(actual) == strings.TrimSpace(expected) {
		return "output_whitespace_mismatch"
	}

	simplify := func(s string) string {
		return strings.Join(strings.Fields(s), "")
	}
	if simplify(actual) == simplify(expected) {
		return "output_whitespace_mismatch"
	}

	return "wrong_output"
}

func mergeLimits(reqStep *models.StepConfig, defaults config.Limits) models.Limits {
	final := models.Limits{
		WallTimeS:    defaults.WallTimeS,
		MemoryKB:     defaults.MemoryKB,
		MaxProcesses: defaults.MaxProcesses,
	}

	if reqStep != nil && reqStep.Limits != nil {
		if reqStep.Limits.WallTimeS > 0 {
			final.WallTimeS = reqStep.Limits.WallTimeS
		}
		if reqStep.Limits.MemoryKB > 0 {
			final.MemoryKB = reqStep.Limits.MemoryKB
		}
		if reqStep.Limits.MaxProcesses > 0 {
			final.MaxProcesses = reqStep.Limits.MaxProcesses
		}
	}
	return final
}

func replacePlaceholders(args []string, source, artifact, flags string) []string {
	var out []string
	replacer := strings.NewReplacer(
		"{{source}}", source,
		"{{artifact}}", artifact,
	)

	for _, arg := range args {
		if arg == "{{flags}}" {
			if flags != "" {
				out = append(out, strings.Fields(flags)...)
			}
			continue
		}
		out = append(out, replacer.Replace(arg))
	}
	return out
}

func replaceString(target, source, artifact, flags string) string {
	replacer := strings.NewReplacer(
		"{{source}}", source,
		"{{artifact}}", artifact,
		"{{flags}}", flags,
	)
	return replacer.Replace(target)
}
