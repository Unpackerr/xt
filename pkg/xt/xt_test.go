package xt //nolint:testpackage

import (
	"archive/zip"
	"os"
	"path/filepath"
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
