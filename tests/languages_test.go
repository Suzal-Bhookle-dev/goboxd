//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/thesouldev/goboxd/internal/api"
	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/models"
)

func setupApp(t *testing.T) *fiber.App {
	cfg, err := config.Load("../languages.yaml")
	if err != nil {
		t.Fatalf("Failed to load config for integration tests: %v", err)
	}

	app := fiber.New()
	api.RegisterRoutes(app, cfg)
	return app
}

func TestHealthz(t *testing.T) {
	app := setupApp(t)
	req, _ := http.NewRequest("GET", "/healthz", nil)
	resp, err := app.Test(req)

	if err != nil || resp.StatusCode != 200 {
		t.Errorf("Healthz failed: expected 200, got %d", resp.StatusCode)
	}
}

func TestAllLanguages(t *testing.T) {
	app := setupApp(t)

	testCases := []struct {
		name    string
		request models.RunRequest
	}{
		{
			name: "C Execution",
			request: models.RunRequest{
				Language: "c",
				Source:   "#include <stdio.h>\nint main() { printf(\"hi\"); return 0; }",
				Tests:    []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "C++ Execution",
			request: models.RunRequest{
				Language: "cpp",
				Source:   "#include <iostream>\nint main() { std::cout << \"hi\"; return 0; }",
				Tests:    []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "Python3 Execution",
			request: models.RunRequest{
				Language: "py3",
				Source:   "print('hi', end='')",
				Tests:    []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "JavaScript Execution",
			request: models.RunRequest{
				Language: "js",
				Source:   "process.stdout.write('hi')",
				Tests:    []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "Rust Execution",
			request: models.RunRequest{
				Language: "rust",
				Source:   "fn main() { print!(\"hi\"); }",
				Tests:    []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "Java Execution",
			request: models.RunRequest{
				Language:         "java",
				SourceFilename:   "Main.java",
				ArtifactFilename: "Main",
				Source:           "public class Main { public static void main(String[] args) { System.out.print(\"hi\"); } }",
				Tests:            []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "Verilog Execution",
			request: models.RunRequest{
				Language: "verilog",
				Source:   "module test; initial begin $write(\"hi\"); $finish; end endmodule",
				Tests:    []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "Perl Execution",
			request: models.RunRequest{
				Language: "perl",
				Source:   "print 'hi';",
				Tests:    []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "Ada Execution",
			request: models.RunRequest{
				Language:         "ada",
				Source:           "with Ada.Text_IO; use Ada.Text_IO; procedure solution is begin Put(\"hi\"); end solution;",
				SourceFilename:   "solution.adb",
				ArtifactFilename: "solution",
				Tests:            []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
		{
			name: "Go Execution",
			request: models.RunRequest{
				Language: "go",
				Source:   "package main\nimport \"fmt\"\nfunc main() {\n    fmt.Print(\"hi\")\n}",
				Tests:    []models.TestCase{{ExpectedStdout: "hi"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.request)
			req, _ := http.NewRequest("POST", "/run", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			var res models.RunResponse
			json.NewDecoder(resp.Body).Decode(&res)

			if res.Status != "accepted" {
				buildErr := "No build step"
				if res.Build != nil {
					buildErr = res.Build.Stderr
				}

				testErr := "No test output"
				if len(res.Tests) > 0 {
					testErr = res.Tests[0].Stderr
				}

				t.Errorf("%s failed: expected accepted, got %s.\nBuild Err: %s\nTest Err: %s",
					tc.name, res.Status, buildErr, testErr)
			}
		})
	}
}
