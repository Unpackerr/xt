package xt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golift.io/xtractr"
)

// RAR multipart names, matching xtractr's folder search: only part 1, and
// .r00 only when that folder has no sibling .rar.
var (
	rarHasParts = regexp.MustCompile(`\.part\d+\.rar$`)
	rarPartOne  = regexp.MustCompile(`\.part0*1\.rar$`)
)

// nestedStats is one second pass over the files an extract just wrote.
type nestedStats struct {
	archives int
	size     uint64
	files    int
}

// extractNested extracts archives that the first pass wrote. It runs once:
// an archive packed inside one of those stays packed. A failure is logged
// and the rest of the pass continues. Nested archives are not deleted.
func (j *Job) extractNested(output string, written, skip []string, preexisting map[string]struct{}) nestedStats {
	archives := j.nestedArchives(output, written, skip, preexisting)
	stats := nestedStats{archives: len(archives)}

	for idx, archive := range archives {
		log.Printf("==> Extracting nested archive (%d/%d): %s", idx+1, len(archives), archive)

		size, count, duration, err := j.extractNestedArchive(archive)
		if err != nil {
			log.Printf("[ERROR] Extracting: %v", err)

			continue
		}

		log.Printf("==> Extracted Archive %s to %s in %v: bytes: %d, files: %d",
			archive, filepath.Dir(archive), duration.Round(time.Millisecond), size, count)

		stats.size += size
		stats.files += count
	}

	return stats
}

// extractNestedArchive extracts one archive into the folder it is already in,
// with the job's passwords, modes, and squash setting.
func (j *Job) extractNestedArchive(archive string) (uint64, int, time.Duration, error) {
	file := &xtractr.XFile{
		FilePath:   archive,
		OutputDir:  filepath.Dir(archive),
		FileMode:   j.FileMode.Mode(),
		DirMode:    j.DirMode.Mode(),
		Passwords:  j.Passwords,
		SquashRoot: j.SquashRoot,
		Updates:    j.progress,
	}
	file.SetLogger(j)

	start := time.Now()

	size, files, _, err := xtractr.ExtractFile(file)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("archive: %s: %w", archive, err)
	}

	return size, len(files), time.Since(start), nil
}

// preexistingNames is the output folder's top-level names before an extract.
// xtractr v0.8 and v0.9 can return every name in that folder when a tar's
// members start with "./" and squash is on. Those names are skipped so a
// second pass does not extract archives the job did not just write.
func preexistingNames(dir string) map[string]struct{} {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return map[string]struct{}{}
	}

	names := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		names[entry.Name()] = struct{}{}
	}

	return names
}

func (j *Job) suffixExcludes() xtractr.Exclude {
	if len(j.Include) > 0 {
		return xtractr.AllExcept(j.Include...)
	}

	return j.Exclude
}

// nestedArchives resolves the extract's file list the way the library writes
// paths, then keeps the names a folder search would keep. A directory in the
// list (squash moved a folder) is searched. The output folder itself is not.
func (j *Job) nestedArchives(output string, written, skip []string, preexisting map[string]struct{}) []string {
	exclude := j.suffixExcludes()
	skipped := resolveSet(output, skip)
	siblings := map[string][]string{}
	seen := map[string]struct{}{}

	found := make([]string, 0, len(written))

	for _, name := range written {
		found = append(found, j.nestedFrom(output, name, exclude, skipped, preexisting, siblings, seen)...)
	}

	return found
}

func (j *Job) nestedFrom(
	output, name string,
	exclude xtractr.Exclude,
	skip, preexisting map[string]struct{},
	siblings map[string][]string,
	seen map[string]struct{},
) []string {
	path := resolveWritten(output, name)
	if path == "" || filepath.Clean(path) == filepath.Clean(output) {
		return nil
	}

	if _, found := skip[path]; found {
		return nil
	}

	if top, ok := topName(output, path); ok && j.alreadyThere(top, preexisting) {
		log.Printf("==> Skipping pre-existing output entry: %s", path)

		return nil
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil
	}

	if info.IsDir() {
		return j.archivesUnder(path, exclude, siblings, seen)
	}

	if !j.wantArchive(path, info, exclude, siblings) {
		return nil
	}

	if _, found := seen[path]; found {
		return nil
	}

	seen[path] = struct{}{}

	return []string{path}
}

func (j *Job) alreadyThere(top string, preexisting map[string]struct{}) bool {
	if !j.SquashRoot || preexisting == nil {
		return false
	}

	_, found := preexisting[top]

	return found
}

func (j *Job) archivesUnder(
	root string,
	exclude xtractr.Exclude,
	siblings map[string][]string,
	seen map[string]struct{},
) []string {
	var found []string

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == root || entry.IsDir() || !entry.Type().IsRegular() {
			return nil
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			return fmt.Errorf("stat %s: %w", path, infoErr)
		}

		if !j.wantArchive(path, info, exclude, siblings) {
			return nil
		}

		if _, ok := seen[path]; ok {
			return nil
		}

		seen[path] = struct{}{}
		found = append(found, path)

		return nil
	})
	if err != nil {
		log.Printf("[ERROR] Reading extracted folder: %v", err)
	}

	return found
}

func (j *Job) wantArchive(path string, info os.FileInfo, exclude xtractr.Exclude, siblings map[string][]string) bool {
	if !info.Mode().IsRegular() {
		return false
	}

	name := strings.ToLower(info.Name())
	if name == "" || name[0] == '.' || exclude.Has(name) {
		return false
	}

	switch {
	case strings.HasSuffix(name, ".rar"):
		return !rarHasParts.MatchString(name) || rarPartOne.MatchString(name)
	case strings.HasSuffix(name, ".r00"):
		return !hasSiblingRar(filepath.Dir(path), name, siblings)
	default:
		return xtractr.IsArchiveFile(name)
	}
}

func hasSiblingRar(dir, r00 string, cache map[string][]string) bool {
	names, ok := cache[dir]
	if !ok {
		entries, err := os.ReadDir(dir)
		if err != nil {
			cache[dir] = nil

			return false
		}

		names = make([]string, len(entries))
		for idx, entry := range entries {
			names[idx] = entry.Name()
		}

		cache[dir] = names
	}

	stem := strings.TrimSuffix(strings.TrimSuffix(r00, ".R00"), ".r00") + ".rar"

	for _, name := range names {
		if strings.EqualFold(name, stem) {
			return true
		}
	}

	return false
}

func resolveSet(output string, paths []string) map[string]struct{} {
	set := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}

		set[resolveWritten(output, path)] = struct{}{}
	}

	return set
}

// resolveWritten matches the library: an absolute path already inside output
// is the file that was written. Anything else is a tar member name, joined
// onto the output folder.
func resolveWritten(output, path string) string {
	if path == "" {
		return ""
	}

	cleaned := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(cleaned) && withinDir(output, cleaned) {
		return cleaned
	}

	return filepath.Clean(filepath.Join(output, cleaned))
}

func topName(output, path string) (string, bool) {
	if !withinDir(output, path) || filepath.Clean(path) == filepath.Clean(output) {
		return "", false
	}

	rel, err := filepath.Rel(output, path)
	if err != nil || rel == "." || rel == "" {
		return "", false
	}

	top, _, _ := strings.Cut(rel, string(filepath.Separator))
	if top == "" || top == "." || top == ".." {
		return "", false
	}

	return top, true
}

func withinDir(base, target string) bool {
	rel, err := filepath.Rel(filepath.Clean(base), filepath.Clean(target))
	if err != nil {
		return false
	}

	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
