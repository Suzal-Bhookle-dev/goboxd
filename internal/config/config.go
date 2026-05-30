package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Limits struct {
	WallTimeS    int `yaml:"wall_time_s"`
	MemoryKB     int `yaml:"memory_kb"`
	MaxProcesses int `yaml:"max_processes"`
}

type StepConfig struct {
	Cmd           string   `yaml:"cmd"`
	Args          []string `yaml:"args"`
	Limits        Limits   `yaml:"limits"`
	FlagAllowlist []string `yaml:"flag_allowlist"`
}

type Language struct {
	ID                     string      `yaml:"id"`
	Name                   string      `yaml:"name"`
	SourceFilename         string      `yaml:"source_filename"`
	SourceFilenameStrategy string      `yaml:"source_filename_strategy"`
	Artifact               string      `yaml:"artifact"`
	Build                  *StepConfig `yaml:"build,omitempty"`
	Run                    StepConfig  `yaml:"run"`
}

type Config struct {
	Languages map[string]Language
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Languages []Language `yaml:"languages"`
	}

	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	cfg := &Config{Languages: make(map[string]Language)}
	for _, l := range raw.Languages {
		cfg.Languages[l.ID] = l
	}

	return cfg, nil
}

func (c *Config) GetLanguage(id string) (Language, bool) {
	l, ok := c.Languages[id]
	return l, ok
}
