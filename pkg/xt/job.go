package xt

import (
	"fmt"

	"golift.io/cnfgfile"
)

// Job defines the input data for one extraction run.
type Job struct {
	Paths     []string `json:"paths" yaml:"paths" xml:"path" toml:"paths"`
	Output    string   `json:"output" yaml:"output" xml:"output" toml:"output"`
	Passwords []string `json:"passwords" yaml:"passwords" xml:"password" toml:"passwords"`
	Exclude   []string `json:"excludeSuffix" yaml:"excludeSuffix" xml:"exclude_suffix" toml:"exclude_suffix"`
	MaxDepth  uint     `json:"maxDepth" yaml:"maxDepth" xml:"max_depth" toml:"max_depth"`
	MinDepth  uint     `json:"minDepth" yaml:"minDepth" xml:"min_depth" toml:"min_depth"`
	DirMode   FileMode `json:"dirMode" yaml:"dirMode" xml:"dir_mode" toml:"dir_mode"`
	FileMode  FileMode `json:"fileMode" yaml:"fileMode" xml:"file_mode" toml:"file_mode"`
}

// ParseJobs checks for and reads more jobs in from 0 or more job files.
func ParseJobs(jobFiles []string) ([]*Job, error) {
	jobs := make([]*Job, len(jobFiles))

	for idx, jobFile := range jobFiles {
		jobs[idx] = &Job{}
		// This library simply parses xml, json, toml, and yaml into a data struct.
		// It parses based on file name extension, toml is default. Supports compression.
		err := cnfgfile.Unmarshal(jobs[idx], jobFile)
		if err != nil {
			return nil, fmt.Errorf("bad job file: %w", err)
		}
	}

	return jobs, nil
}

func (j *Job) fixModes() {
	const (
		defaultFileMode = 0o644
		defaultDirMode  = 0o755
	)

	if j.DirMode == 0 {
		j.DirMode = defaultDirMode
	}

	if j.FileMode == 0 {
		j.FileMode = defaultFileMode
	}
}

func (j *Job) String() string {
	j.fixModes()

	sSfx := ""
	if len(j.Paths) > 1 {
		sSfx = "s"
	}

	return fmt.Sprintf("%d path%s, f/d-mode:%s/%s, min/max-depth: %d/%d output: %s",
		len(j.Paths), sSfx, j.FileMode, j.DirMode, j.MinDepth, j.MaxDepth, j.Output)
}
