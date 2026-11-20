// Package xt provides the interface to extract archive files based on job parameters.
package xt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
	"golift.io/xtractr"
)

// Extract runs a job immediately.
func Extract(job *Job) {
	archives := job.getArchives()
	if len(archives) == 0 {
		log.Println("==> No archives found in:", job.Paths)
	}

	job.fixModes()

	total := archives.Count()

	stopProgress := job.setupProgress(total)
	defer stopProgress()

	count := 0
	size := uint64(0)
	fCount := 0
	nested := 0
	start := time.Now()

	for folder, files := range archives {
		for _, archive := range files {
			count++
			log.Printf("==> Extracting Archive (%d/%d): %s", count, total, archive)

			output, fSize, files, duration, extra, err := job.processArchive(folder, archive)
			if err != nil {
				log.Printf("[ERROR] Extracting: %v", err)
			} else {
				log.Printf("==> Extracted Archive %s to %s in %v: bytes: %d, files: %d",
					archive, output, duration.Round(time.Millisecond), fSize, len(files))
			}

			if len(files) > 0 && job.Verbose {
				log.Printf("==> Files:\n - %s", strings.Join(files, "\n - "))
			}

			size += fSize + extra.size
			fCount += len(files) + extra.files
			nested += extra.archives
		}
	}

	log.Printf("==> Done.")
	logExtracted(total, nested, size, fCount, time.Since(start))
}

func logExtracted(total, nested int, size uint64, files int, elapsed time.Duration) {
	elapsed = elapsed.Round(time.Millisecond)

	if nested > 0 {
		log.Printf("==> Extracted %d archives (%d nested); wrote %d bytes into %d files in %v",
			total+nested, nested, size, files, elapsed)

		return
	}

	log.Printf("==> Extracted %d archives; wrote %d bytes into %d files in %v",
		total, size, files, elapsed)
}

// setupProgress starts a printer goroutine when progress output is useful.
// The returned func closes the channel and waits for that goroutine to exit.
func (j *Job) setupProgress(total int) func() {
	const (
		intervalTTY    = 0.1
		intervalNonTTY = 5
	)

	if j.DebugLog || total < 1 {
		return func() {}
	}

	every := float64(intervalNonTTY)
	isTerm := term.IsTerminal(int(os.Stdout.Fd()))

	if isTerm {
		every = intervalTTY
	}

	j.progress = make(chan xtractr.Progress)
	done := make(chan struct{})

	go func() {
		defer close(done)

		xtractr.ArchiveProgress(every, j.progress, isTerm, false)
	}()

	return sync.OnceFunc(func() {
		close(j.progress)
		<-done

		j.progress = nil
	})
}

func (j *Job) processArchive(folder, archive string) (string, uint64, []string, time.Duration, nestedStats, error) {
	file := &xtractr.XFile{
		FilePath:   archive,           // Path to archive being extracted.
		OutputDir:  j.Output,          // Folder to extract archive into.
		FileMode:   j.FileMode.Mode(), // Write files with this mode.
		DirMode:    j.DirMode.Mode(),  // Write folders with this mode.
		Passwords:  j.Passwords,       // (RAR/7zip) Archive password(s).
		SquashRoot: j.SquashRoot,      // Remove single root folder?
		Updates:    j.progress,
	}
	file.SetLogger(j)

	// If preserving the file hierarchy: set the output directory to the same path as the input file.
	if j.Preserve {
		// Rel(search folder, archive) then join that directory onto job.Output.
		// TrimPrefix leaves a leading separator, and Join treats that as absolute.
		rel, err := filepath.Rel(folder, archive)
		if err != nil {
			rel = strings.TrimPrefix(archive, folder)
		}

		file.OutputDir = filepath.Join(j.Output, filepath.Dir(rel))
	}

	// Absolute, so a second pass can tell a written path from a tar member
	// name, and so squash works for a relative output folder.
	absOutput, err := filepath.Abs(file.OutputDir)
	if err != nil {
		return "", 0, nil, 0, nestedStats{}, fmt.Errorf("output directory: %w", err)
	}

	file.OutputDir = absOutput

	// Read once, and only when both flags are set. See preexistingNames.
	var preexisting map[string]struct{}
	if j.Recurse && j.SquashRoot {
		preexisting = preexistingNames(file.OutputDir)
	}

	start := time.Now()

	size, files, _, err := xtractr.ExtractFile(file)
	if err != nil {
		return file.OutputDir, size, files, time.Since(start), nestedStats{}, fmt.Errorf("archive: %s: %w", archive, err)
	}

	var nested nestedStats
	if j.Recurse {
		nested = j.extractNested(file.OutputDir, files, file.SkipOnRecursion, preexisting)
	}

	return file.OutputDir, size, files, time.Since(start), nested, nil
}

func (j *Job) getArchives() xtractr.ArchiveList {
	archives := map[string][]string{}

	for _, fileName := range j.Paths {
		fileInfo, err := os.Stat(fileName)
		if err != nil {
			log.Println("[ERROR] Reading archive path:", err)
			continue
		}

		if !fileInfo.IsDir() {
			archives[fileName] = []string{fileName}
			continue
		}

		exclude := j.Exclude
		if len(j.Include) > 0 {
			exclude = xtractr.AllExcept(j.Include...)
		}

		for _, fileList := range xtractr.FindCompressedFiles(xtractr.Filter{
			Path:          fileName,
			ExcludeSuffix: exclude,
			MaxDepth:      int(j.MaxDepth),
			MinDepth:      int(j.MinDepth),
		}) {
			// Group archive lists by the parent search folder that found them.
			archives[fileName] = append(archives[fileName], fileList...)
		}
	}

	return archives
}
