package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/models"
)

func Execute(req models.RunRequest, lang config.Language) (models.RunResponse, error) {
	tmpdir, err := os.MkdirTemp("", "/goboxd-run-")
	if err != nil {
		return models.RunResponse{}, fmt.Errorf("failed to create sandbox dir: %w", err)
	}

	defer os.RemoveAll(tmpdir)

	sourceFilename := req.SourceFilename
	if sourceFilename == "" {
		sourceFilename = lang.SourceFilename
	}

	artifactFilename := req.ArtifactFilename
	if artifactFilename == "" {
		artifactFilename = lang.Artifact
	}

	sourcePath := filepath.Join(tmpdir, sourceFilename)
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0o644); err != nil {
		return models.RunResponse{}, fmt.Errorf("failed to write source file: %w", err)
	}

	response := models.RunResponse{
		Status: "ok",
		Tests:  []models.TestResult{},
	}

	if lang.Build != nil {
		buildStart := time.Now()

		userFlags := ""
		if req.Build != nil {
			userFlags = strings.Join(req.Build.Flags, " ")
		}

		args := replacePlaceholders(lang.Build.Args, sourceFilename, artifactFilename, userFlags)

		buildLimits := mergeLimits(req.Build, lang.Build.Limits)

		res, err := runInNsjail(tmpDir, buildLimits, "", lang.Build.Cmd, args...)

		response.Build = &models.StepResult{
			Status:     "ok",
			Stdout:     res.Stdout,
			Stderr:     res.Stderr,
			DurationMs: time.Since(buildStart).Milliseconds(),
		}

		if err != nil || res.ExitCode != 0 {
			response.Build.Status = "error"
			response.Status = "build_failed"
			return response, nil
		}
	}

	for _, tc := range req.Tests {
		runStart := time.Now()

		runArgs := replacePlaceholders(lang.Run.Args, sourceFilename, artifactFilename, "")

		runLimits := mergeLimits(req.Run, lang.Run.Limits)

		res, _ := runInNsjail(tmpDir, runLimits, tc.Stdin, lang.Run.Cmd, runArgs...)

		testStatus := "ok"
		if strings.TrimSpace(res.Stdout) != strings.TrimSpace(tc.ExpectedStdout) {
			testStatus = "wrong_output"
			response.Status = "wrong_output"
		}
		if res.TimedOut {
			testStatus = "time_limit_exceeded"
			response.Status = "error"
		}

		response.Tests = append(response.Tests, models.TestResult{
			Status:       testStatus,
			Stdout:       res.Stdout,
			Stderr:       res.Stderr,
			DurationMs:   time.Since(runStart).Milliseconds(),
			MemoryPeakKb: res.MemoryKb,
		})
	}

	return response, nil
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
	out := make([]string, len(args))
	replacer := strings.NewReplacer(
		"{{source}}", source,
		"{{artifact}}", artifact,
		"{{flags}}", flags,
	)
	for i, arg := range args {
		out[i] = replacer.Replace(arg)
	}
	return out
}
