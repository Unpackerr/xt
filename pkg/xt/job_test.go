package xt //nolint:testpackage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseJobs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	tomlPath := filepath.Join(dir, "job.toml")
	writeFile(t, tomlPath, "paths = ['/tmp/a']\noutput = '.'\npreserve_paths = true\n")

	jsonPath := filepath.Join(dir, "job.json")
	writeFile(t, jsonPath, `{"paths":["/tmp/b"],"output":".","squashRoot":true}`)

	yamlPath := filepath.Join(dir, "job.yaml")
	writeFile(t, yamlPath, "paths:\n  - /tmp/c\noutput: .\nverbose: true\n")

	jobs, err := ParseJobs([]string{tomlPath, jsonPath, yamlPath})
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 3 {
		t.Fatalf("len(jobs) = %d", len(jobs))
	}

	if jobs[0].Paths[0] != "/tmp/a" || !jobs[0].Preserve || jobs[0].Output != "." {
		t.Fatalf("toml job = %+v", jobs[0])
	}

	if jobs[1].Paths[0] != "/tmp/b" || !jobs[1].SquashRoot {
		t.Fatalf("json job = %+v", jobs[1])
	}

	if jobs[2].Paths[0] != "/tmp/c" || !jobs[2].Verbose {
		t.Fatalf("yaml job = %+v", jobs[2])
	}
}

func TestParseJobsMissingFile(t *testing.T) {
	t.Parallel()

	_, err := ParseJobs([]string{filepath.Join(t.TempDir(), "missing.toml")})
	if err == nil {
		t.Fatal("expected error for missing job file")
	}

	if !strings.Contains(err.Error(), "bad job file") {
		t.Fatalf("error = %v", err)
	}
}

func TestFixModesAndString(t *testing.T) {
	t.Parallel()

	job := &Job{Paths: []string{"a"}, Output: "/out"}
	job.fixModes()

	if job.FileMode.Mode() != 0o644 {
		t.Fatalf("FileMode = %o", job.FileMode.Mode())
	}

	if job.DirMode.Mode() != 0o755 {
		t.Fatalf("DirMode = %o", job.DirMode.Mode())
	}

	job.FileMode = 0o600
	job.DirMode = 0o700
	job.fixModes()

	if job.FileMode.Mode() != 0o600 || job.DirMode.Mode() != 0o700 {
		t.Fatal("fixModes overwrote non-zero modes")
	}

	got := job.String()
	if !strings.Contains(got, "1 path,") || !strings.Contains(got, "f/d-mode:0600/0700") {
		t.Fatalf("String = %q", got)
	}

	job.Paths = []string{"a", "b"}
	if !strings.Contains(job.String(), "2 paths,") {
		t.Fatalf("plural String = %q", job.String())
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()

	err := os.WriteFile(path, []byte(body), 0o600)
	if err != nil {
		t.Fatal(err)
	}
}
