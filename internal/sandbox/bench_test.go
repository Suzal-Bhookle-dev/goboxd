package sandbox

import (
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/models"
)

func BenchmarkCompareOutput(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compareOutput("hello world\n", "hello world")
	}
}

func BenchmarkMergeLimits(b *testing.B) {
	defaults := config.Limits{WallTimeS: 5, MemoryKB: 1024}
	req := &models.StepConfig{Limits: &models.Limits{WallTimeS: 10}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mergeLimits(req, defaults)
	}
}

func BenchmarkReplacePlaceholders(b *testing.B) {
	args := []string{"g++", "{{source}}", "-o", "{{artifact}}", "{{flags}}"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replacePlaceholders(args, "solution.cpp", "solution", "-O3")
	}
}
