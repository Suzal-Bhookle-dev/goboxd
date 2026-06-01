package api

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/thesouldev/goboxd/internal/models"
)

func (h *Handler) ValidateRequest(req *models.RunRequest) *models.ErrorResponse {
	lang, exists := h.Cfg.GetLanguage(req.Language)
	if !exists {
		return errRes("unknown_language", "The requested language is not supported")
	}

	if req.Source == "" || !utf8.ValidString(req.Source) {
		return errRes("invalid_source", "Source must be valid UTF-8")
	}
	if len(req.Source) > 256*1024 {
		return errRes("source_too_large", "Source code exceeds 256 KiB limit")
	}

	if err := validatePath(req.SourceFilename); err != nil {
		return errRes("invalid_filename", "source_filename: "+err.Error())
	}
	if err := validatePath(req.ArtifactFilename); err != nil {
		return errRes("invalid_filename", "artifact_filename: "+err.Error())
	}

	if req.Build != nil {
		if err := validateFlags(req.Build.Flags, lang.Build.FlagAllowlist); err != nil {
			return errRes("disallowed_flag", err.Error())
		}
	}
	if req.Run != nil {
		if err := validateFlags(req.Run.Flags, lang.Run.FlagAllowlist); err != nil {
			return errRes("disallowed_flag", err.Error())
		}
	}

	if len(req.Tests) == 0 {
		return errRes("missing_tests", "At least one test case is required")
	}

	return nil
}

func validatePath(p string) error {
	if p == "" {
		return nil
	}
	if filepath.Base(p) != p || strings.HasPrefix(p, ".") || strings.Contains(p, "/") || strings.Contains(p, "\\") {
		return fmt.Errorf("must be a single path component with no separators or leading dots")
	}
	if len(p) > 64 {
		return fmt.Errorf("length exceeds 64 characters")
	}
	return nil
}

func validateFlags(provided []string, allowed []string) error {
	for _, p := range provided {
		match := false
		for _, a := range allowed {
			if p == a || (strings.HasSuffix(a, "*") && strings.HasPrefix(p, a[:len(a)-1])) {
				match = true
				break
			}
		}
		if !match {
			return fmt.Errorf("flag '%s' is not in the allow-list", p)
		}
	}
	return nil
}

func errRes(code, msg string) *models.ErrorResponse {
	return &models.ErrorResponse{
		Error: models.ErrorDetail{Code: code, Message: msg},
	}
}
