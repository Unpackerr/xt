package xt //nolint:testpackage // tests call unexported extract helpers.

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestRecurseZipInZip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inner := zipBytes(t, "hello.txt", []byte("nested"))
	outer := filepath.Join(dir, "outer.zip")
	writeZip(t, outer, "inner.zip", inner)

	out := filepath.Join(dir, "out")
	Extract(&Job{Paths: []string{outer}, Output: out})

	_, err := os.Stat(filepath.Join(out, "hello.txt"))
	if !os.IsNotExist(err) {
		t.Fatal("recurse is on by default")
	}

	Extract(&Job{Paths: []string{outer}, Output: out, Recurse: true, FileMode: 0o600})

	got, err := os.ReadFile(filepath.Join(out, "hello.txt")) //nolint:gosec
	if err != nil || string(got) != "nested" {
		t.Fatalf("nested file = %q, %v", got, err)
	}

	info, err := os.Stat(filepath.Join(out, "hello.txt"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("nested mode = %v, %v", info, err)
	}
}

func TestRecurseThirdLevelStaysPacked(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	hello := zipBytes(t, "hello.txt", []byte("deep"))
	mid := zipBytes(t, "inner.zip", hello)
	outer := filepath.Join(dir, "outer.zip")
	writeZip(t, outer, "mid.zip", mid)

	out := filepath.Join(dir, "out")
	Extract(&Job{Paths: []string{outer}, Output: out, Recurse: true})

	_, err := os.Stat(filepath.Join(out, "inner.zip"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(filepath.Join(out, "hello.txt"))
	if !os.IsNotExist(err) {
		t.Fatal("third level was extracted")
	}
}

func TestRecurseContinuesAfterNestedFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	good := zipBytes(t, "hello.txt", []byte("ok"))
	outer := filepath.Join(dir, "outer.zip")
	writeZipFiles(t, outer, []namedFile{{"bad.zip", []byte("not a zip")}, {"good.zip", good}})

	out := filepath.Join(dir, "out")
	Extract(&Job{Paths: []string{outer}, Output: out, Recurse: true})

	got, err := os.ReadFile(filepath.Join(out, "hello.txt")) //nolint:gosec
	if err != nil || string(got) != "ok" {
		t.Fatalf("job stopped after nested failure: %q %v", got, err)
	}

	_, err = os.Stat(filepath.Join(out, "bad.zip"))
	if err != nil {
		t.Fatal("nested archive was deleted")
	}
}

func TestRecurseArchiveInTar(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	archive := filepath.Join(dir, "bundle.tar")
	writeTar(t, archive, "nested/inner.zip", zipBytes(t, "hello.txt", []byte("from tar")))

	out := filepath.Join(dir, "out")
	Extract(&Job{Paths: []string{archive}, Output: out, Recurse: true})

	got, err := os.ReadFile(filepath.Join(out, "nested", "hello.txt")) //nolint:gosec
	if err != nil || string(got) != "from tar" {
		t.Fatalf("tar member = %q, %v", got, err)
	}
}

func TestRecurseLeavesUnrelatedArchive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	out := filepath.Join(dir, "out")

	err := os.MkdirAll(out, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	writeZip(t, filepath.Join(out, "other.zip"), "secret.txt", []byte("nope"))
	outer := filepath.Join(dir, "outer.zip")
	writeZip(t, outer, "hello.txt", []byte("yes"))

	Extract(&Job{Paths: []string{outer}, Output: out, Recurse: true})

	_, statErr := os.Stat(filepath.Join(out, "secret.txt"))
	if !os.IsNotExist(statErr) {
		t.Fatal("unrelated archive in the output folder was extracted")
	}
}

func TestRecurseSquashRoot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inner := zipBytes(t, "only/hello.txt", []byte("up"))
	outer := filepath.Join(dir, "outer.zip")
	writeZip(t, outer, "root/inner.zip", inner)

	out := filepath.Join(dir, "out")
	Extract(&Job{Paths: []string{outer}, Output: out, Recurse: true, SquashRoot: true})

	got, err := os.ReadFile(filepath.Join(out, "hello.txt")) //nolint:gosec
	if err != nil || string(got) != "up" {
		t.Fatalf("squashed nested file = %q, %v", got, err)
	}
}

func TestRecurseSquashRootRelativeOutput(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	inner := zipBytes(t, "only/hello.txt", []byte("rel"))
	writeZip(t, filepath.Join(dir, "outer.zip"), "root/inner.zip", inner)

	Extract(&Job{Paths: []string{"outer.zip"}, Output: "out", Recurse: true, SquashRoot: true})

	got, err := os.ReadFile(filepath.Join(dir, "out", "hello.txt")) //nolint:gosec
	if err != nil || string(got) != "rel" {
		t.Fatalf("relative squash = %q, %v", got, err)
	}
}

//nolint:paralleltest // log.SetOutput is process-wide.
func TestRecurseSquashSkipsPreexisting(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out")

	err := os.MkdirAll(out, 0o750)
	if err != nil {
		t.Fatal(err)
	}

	writeZip(t, filepath.Join(out, "keep.zip"), "secret.txt", []byte("nope"))
	archive := filepath.Join(dir, "dot.tar")
	writeTar(t, archive, "./inside.zip", zipBytes(t, "hello.txt", []byte("yes")))

	var buf bytes.Buffer

	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	Extract(&Job{Paths: []string{archive}, Output: out, Recurse: true, SquashRoot: true})

	got, err := os.ReadFile(filepath.Join(out, "hello.txt")) //nolint:gosec
	if err != nil || string(got) != "yes" {
		t.Fatalf("nested from dot tar = %q, %v", got, err)
	}

	_, statErr := os.Stat(filepath.Join(out, "secret.txt"))
	if !os.IsNotExist(statErr) {
		t.Fatal("pre-existing archive was extracted")
	}

	logged := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("Skipping pre-existing")) || !bytes.Contains(buf.Bytes(), []byte("1 nested")) {
		t.Fatalf("log = %s", logged)
	}
}

func TestNestedArchiveNameRules(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	zipName := "a.zip"
	names := []string{"vol.part1.rar", "vol.part2.rar", "movie.r00", "movie.rar", "lone.r00", "note.txt", zipName}

	for _, name := range names {
		writeFile(t, filepath.Join(dir, name), "x")
	}

	job := &Job{}
	got := baseNames(job.nestedArchives(dir, names, nil, nil))
	want := []string{"vol.part1.rar", "movie.rar", "lone.r00", zipName}

	if !sameNames(got, want) {
		t.Fatalf("names = %v, want %v", got, want)
	}

	job.Include = []string{zipSuffix}
	got = baseNames(job.nestedArchives(dir, names, nil, nil))

	if !sameNames(got, []string{zipName}) {
		t.Fatalf("include = %v", got)
	}

	job = &Job{Exclude: []string{zipSuffix}}
	got = baseNames(job.nestedArchives(dir, names, []string{filepath.Join(dir, zipName)}, nil))

	if slices.Contains(got, zipName) || slices.Contains(got, "vol.part2.rar") {
		t.Fatalf("exclude/skip = %v", got)
	}
}

func TestParseJobRecurse(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "job.toml")
	writeFile(t, path, "paths = ['/tmp/a']\nrecurse = true\n")

	jobs, err := ParseJobs([]string{path})
	if err != nil || len(jobs) != 1 || !jobs[0].Recurse {
		t.Fatalf("jobs = %+v, %v", jobs, err)
	}
}

type namedFile struct {
	name string
	body []byte
}

func zipBytes(t *testing.T, name string, body []byte) []byte {
	t.Helper()

	var buf bytes.Buffer

	writeZipTo(t, &buf, []namedFile{{name, body}})

	return buf.Bytes()
}

func writeZip(t *testing.T, path, name string, body []byte) {
	t.Helper()
	writeZipFiles(t, path, []namedFile{{name, body}})
}

func writeZipFiles(t *testing.T, path string, files []namedFile) {
	t.Helper()

	file, err := os.Create(path) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}

	writeZipTo(t, file, files)

	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func writeZipTo(t *testing.T, file io.Writer, files []namedFile) {
	t.Helper()

	writer := zip.NewWriter(file)

	for _, item := range files {
		entry, err := writer.Create(item.name)
		if err != nil {
			t.Fatal(err)
		}

		_, err = entry.Write(item.body)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := writer.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func writeTar(t *testing.T, path, name string, body []byte) {
	t.Helper()

	file, err := os.Create(path) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}

	writer := tar.NewWriter(file)

	err = writer.WriteHeader(&tar.Header{
		Name:     name,
		Mode:     0o644,
		Size:     int64(len(body)),
		Typeflag: tar.TypeReg,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = writer.Write(body)
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

func baseNames(paths []string) []string {
	names := make([]string, len(paths))
	for idx, path := range paths {
		names[idx] = filepath.Base(path)
	}

	return names
}

func sameNames(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	for _, name := range want {
		if !slices.Contains(got, name) {
			return false
		}
	}

	return true
}
