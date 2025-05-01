// Package xt provides the interface to extract archive files based on job parameters.
package xt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	job.setupProgress()

	total := archives.Count()
	count := 0
	size := uint64(0)
	fCount := 0
	start := time.Now()

	for folder, files := range archives {
		for _, archive := range files {
			count++
			log.Printf("==> Extracting Archive (%d/%d): %s", count, total, archive)

			output, fSize, files, duration, err := job.processArchive(folder, archive)
			if err != nil {
				log.Printf("[ERROR] Extracting: %v", err)
			} else {
				log.Printf("==> Extracted Archive %s to %s in %v: bytes: %d, files: %d",
					archive, output, duration.Round(time.Millisecond), fSize, len(files))
			}

			if len(files) > 0 && job.Verbose {
				log.Printf("==> Files:\n - %s", strings.Join(files, "\n - "))
			}

			size += fSize
			fCount += len(files)
		}
	}

	log.Printf("==> Done.")
	log.Printf("==> Extracted %d archives; wrote %d bytes into %d files in %v",
		total, size, fCount, time.Since(start).Round(time.Millisecond))
}

func (j *Job) setupProgress() {
	every := float64(5) //nolint:mnd

	isTerm := term.IsTerminal(int(os.Stdout.Fd()))
	if isTerm {
		every = 0.1
	}

	if !j.DebugLog { // Only print progress if debug is off.
		j.progress = make(chan xtractr.Progress)
		go xtractr.ArchiveProgress(every, j.progress, term.IsTerminal(int(os.Stdout.Fd())), false)
	}
}

func (j *Job) processArchive(folder, archive string) (string, uint64, []string, time.Duration, error) {
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
		// Remove input path prefix from fileName,
		// append fileName.Dir to job.Output,
		// extract file into job.Output/file(sub)Folder(s).
		file.OutputDir = filepath.Join(j.Output, filepath.Dir(strings.TrimPrefix(archive, folder)))
	}

	start := time.Now()

	size, files, _, err := xtractr.ExtractFile(file)
	if err != nil {
		err = fmt.Errorf("archive: %s: %w", archive, err)
	}

	return file.OutputDir, size, files, time.Since(start), err
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
