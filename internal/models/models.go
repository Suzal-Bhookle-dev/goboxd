package models

type Limits struct {
	WallTimeS    int `json:"wall_time_s" yaml:"wall_time_s"`
	MemoryKB     int `json:"memory_kb" yaml:"memory_kb"`
	MaxProcesses int `json:"max_processes" yaml:"max_processes"`
}

type StepConfig struct {
	Limits *Limits  `json:"limits,omitempty"`
	Flags  []string `json:"flags,omitempty"`
}

type TestCase struct {
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
}

type RunRequest struct {
	Language         string      `json:"language"`
	Source           string      `json:"source"`
	SourceFilename   string      `json:"source_filename,omitempty"`
	ArtifactFilename string      `json:"artifact_filename,omitempty"`
	Build            *StepConfig `json:"build,omitempty"`
	Run              *StepConfig `json:"run,omitempty"`
	Tests            []TestCase  `json:"tests"`
}

type StepResult struct {
	Status     string `json:"status"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	DurationMs int64  `json:"duration_ms"`
}

type TestResult struct {
	Status       string `json:"status"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	DurationMs   int64  `json:"duration_ms"`
	MemoryPeakKb int    `json:"memory_peak_kb"`
}

type RunResponse struct {
	Status string       `json:"status"`
	Build  *StepResult  `json:"build,omitempty"`
	Tests  []TestResult `json:"tests"`
}
