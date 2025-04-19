package xt

import (
	"fmt"
	"log"

	"golift.io/cnfgfile"
)

// Job defines the input data for one extraction run.
type Job struct {
	Paths      []string `json:"paths"         toml:"paths"          xml:"path"           yaml:"paths"`
	Output     string   `json:"output"        toml:"output"         xml:"output"         yaml:"output"`
	Passwords  []string `json:"passwords"     toml:"passwords"      xml:"password"       yaml:"passwords"`
	Exclude    []string `json:"excludeSuffix" toml:"exclude_suffix" xml:"exclude_suffix" yaml:"excludeSuffix"`
	Include    []string `json:"includeSuffix" toml:"include_suffix" xml:"include_suffix" yaml:"includeSuffix"`
	MaxDepth   uint16   `json:"maxDepth"      toml:"max_depth"      xml:"max_depth"      yaml:"maxDepth"`
	MinDepth   uint16   `json:"minDepth"      toml:"min_depth"      xml:"min_depth"      yaml:"minDepth"`
	DirMode    FileMode `json:"dirMode"       toml:"dir_mode"       xml:"dir_mode"       yaml:"dirMode"`
	FileMode   FileMode `json:"fileMode"      toml:"file_mode"      xml:"file_mode"      yaml:"fileMode"`
	SquashRoot bool     `json:"squashRoot"    toml:"squash_root"    xml:"squash_root"    yaml:"squashRoot"`
	DebugLog   bool     `json:"debugLog"      toml:"debug_log"      xml:"debug_log"      yaml:"debugLog"`
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

func (j *Job) String() string {
	j.fixModes()

	sSfx := ""
	if len(j.Paths) > 1 {
		sSfx = "s"
	}

	return fmt.Sprintf("%d path%s, f/d-mode:%s/%s, min/max-depth: %d/%d output: %s",
		len(j.Paths), sSfx, j.FileMode, j.DirMode, j.MinDepth, j.MaxDepth, j.Output)
}

// Debugf prints a log message if debug is enabled.
func (j *Job) Debugf(format string, vars ...any) {
	if j.DebugLog {
		log.Printf("[DEBUG] "+format, vars...)
	}
}

// Printf wraps log.Printf.
func (j *Job) Printf(format string, vars ...any) {
	log.Printf(format, vars...)
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
