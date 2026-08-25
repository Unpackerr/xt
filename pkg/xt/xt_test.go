package xt //nolint:testpackage

import (
	"archive/zip"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestSetupProgressSkipsWhenIdle(t *testing.T) {
	t.Parallel()

	job := &Job{}
	stop := job.setupProgress(0)
	stop()

	if job.progress != nil {
		t.Fatal("progress printer must not start when there are no archives")
	}

	job = &Job{DebugLog: true}
	stop = job.setupProgress(3)
	stop()

	if job.progress != nil {
		t.Fatal("progress printer must not start when debug logging is on")
	}
}

func TestSetupProgressStops(t *testing.T) {
	t.Parallel()

	job := &Job{}
	stop := job.setupProgress(1)

	if job.progress == nil {
		t.Fatal("expected a progress channel")
	}

	done := make(chan struct{})

	go func() {
		stop()
		stop() // OnceFunc: closing twice must not panic.
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("progress printer did not exit after stop")
	}

	if job.progress != nil {
		t.Fatal("progress channel should be cleared after stop")
	}
}

func TestExtractNoArchives(t *testing.T) {
	t.Parallel()

	Extract(&Job{Paths: []string{t.TempDir()}})
}

func TestExtractZip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "sample.zip")
	createTestZip(t, zipPath, "hello.txt", "hello world")

	out := filepath.Join(dir, "out")

	err := os.MkdirAll(out, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	Extract(&Job{Paths: []string{zipPath}, Output: out})

	got, err := os.ReadFile(filepath.Join(out, "hello.txt")) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "hello world" {
		t.Fatalf("extracted content = %q", got)
	}
}

func TestGetArchivesFileVsDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "sample.zip")
	createTestZip(t, zipPath, "hello.txt", "hello")

	job := &Job{Paths: []string{zipPath}}
	got := job.getArchives()

	if got.Count() != 1 || len(got[zipPath]) != 1 || got[zipPath][0] != zipPath {
		t.Fatalf("file archives = %#v", got)
	}

	job = &Job{Paths: []string{dir}}
	got = job.getArchives()

	if got.Count() != 1 || !slices.Contains(got[dir], zipPath) {
		t.Fatalf("dir archives = %#v", got)
	}
}

func TestGetArchivesIncludeExclude(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "sample.zip")
	isoPath := filepath.Join(dir, "sample.iso")
	createTestZip(t, zipPath, "hello.txt", "hello")
	writeFile(t, isoPath, "not-an-iso")

	included := (&Job{Paths: []string{dir}, Include: []string{".zip"}}).getArchives()
	if included.Count() != 1 || !slices.Contains(included[dir], zipPath) {
		t.Fatalf("include .zip = %#v", included)
	}

	if slices.Contains(included[dir], isoPath) {
		t.Fatalf("include .zip should skip iso: %#v", included)
	}

	excluded := (&Job{Paths: []string{dir}, Exclude: []string{".zip"}}).getArchives()
	if slices.Contains(excluded[dir], zipPath) {
		t.Fatalf("exclude .zip still found zip: %#v", excluded)
	}

	if excluded.Count() != 1 || !slices.Contains(excluded[dir], isoPath) {
		t.Fatalf("exclude .zip = %#v", excluded)
	}
}

func TestGetArchivesMissingPath(t *testing.T) {
	t.Parallel()

	got := (&Job{Paths: []string{filepath.Join(t.TempDir(), "missing")}}).getArchives()
	if got.Count() != 0 {
		t.Fatalf("missing path archives = %#v", got)
	}
}

func TestExtractPreservePaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nested := filepath.Join(root, "nested")

	err := os.MkdirAll(nested, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	createTestZip(t, filepath.Join(nested, "sample.zip"), "hello.txt", "hello world")

	out := filepath.Join(root, "out")
	Extract(&Job{Paths: []string{root}, Output: out, Preserve: true})

	got, err := os.ReadFile(filepath.Join(out, "nested", "hello.txt")) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "hello world" {
		t.Fatalf("extracted content = %q", got)
	}
}

func TestExtractSquashRoot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "sample.zip")
	createTestZip(t, zipPath, "root/hello.txt", "squashed")

	out := filepath.Join(dir, "out")

	err := os.MkdirAll(out, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	Extract(&Job{Paths: []string{zipPath}, Output: out, SquashRoot: true})

	got, err := os.ReadFile(filepath.Join(out, "hello.txt")) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "squashed" {
		t.Fatalf("extracted content = %q", got)
	}

	_, err = os.Stat(filepath.Join(out, "root"))
	if !os.IsNotExist(err) {
		t.Fatalf("squash left root folder: %v", err)
	}
}

func createTestZip(t *testing.T, zipPath, name, body string) {
	t.Helper()

	file, err := os.Create(zipPath) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}

	writer := zip.NewWriter(file)

	entry, err := writer.Create(name)
	if err != nil {
		t.Fatal(err)
	}

	_, err = entry.Write([]byte(body))
	if err != nil {
		t.Fatal(err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}
}
