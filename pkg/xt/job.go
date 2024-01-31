package xt

import (
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
}

// ParseJobs checks for and reads more jobs in from 0 or more job files.
func ParseJobs(jobFiles []string) ([]Job, error) {
	jobs := make([]Job, len(jobFiles))

	for idx, jobFile := range jobFiles {
		// This library simply parses xml, json, toml, and yaml into a data struct.
		// It parses based on file name extension, toml is default. Supports compression.
		err := cnfgfile.Unmarshal(&jobs[idx], jobFile)
		if err != nil {
			return nil, err
		}
	}

	return jobs, nil
}
