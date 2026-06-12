package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/models"
)

var (
	buildCache   = make(map[string]string)
	buildCacheMu sync.RWMutex
)

func hashSource(source string) string {
	hash := sha256.Sum256([]byte(source))
	return hex.EncodeToString(hash[:])
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

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

		sourceHash := hashSource(req.Source + lang.ID + uFlags)

		buildCacheMu.RLock()
		cachedArtifactPath, cacheHit := buildCache[sourceHash]
		buildCacheMu.RUnlock()

		bStatus := "ok"
		var stdout, stderr string

		if cacheHit {
			cachedFileName := filepath.Base(cachedArtifactPath)
			err = copyFile(cachedArtifactPath, filepath.Join(tmpDir, cachedFileName))
			if err != nil {
				bStatus = "internal_error"
				stderr = err.Error()
			}
		} else {
			buildCmd := replaceString(lang.Build.Cmd, srcFile, artFile, uFlags)
			args := replacePlaceholders(lang.Build.Args, srcFile, artFile, uFlags)

			res, err := runInNsjail(tmpDir, mergeLimits(req.Build, lang.Build.Limits), "", buildCmd, args...)

			if err != nil {
				bStatus = "internal_error"
			} else if res.ExitCode != 0 {
				bStatus = "failed"
			} else {
				// Find the actual compiled artifact
				artifactFile := artFile
				if _, statErr := os.Stat(filepath.Join(tmpDir, artFile)); os.IsNotExist(statErr) {
					// Java produces .class files
					if _, statErr := os.Stat(filepath.Join(tmpDir, artFile+".class")); statErr == nil {
						artifactFile = artFile + ".class"
					}
				}

				cacheDir, _ := os.MkdirTemp("", "gobox-cache-")
				cachedPath := filepath.Join(cacheDir, artifactFile)
				copyFile(filepath.Join(tmpDir, artifactFile), cachedPath)

				buildCacheMu.Lock()
				buildCache[sourceHash] = cachedPath
				buildCacheMu.Unlock()
			}
			stdout = res.Stdout
			stderr = res.Stderr
		}

		response.Build = &models.StepResult{
			Status:     bStatus,
			Stdout:     stdout,
			Stderr:     stderr,
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
