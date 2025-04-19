// Package xt provides the interface to extract archive files based on job parameters.
package xt

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	count := 0
	size := int64(0)
	fCount := 0
	start := time.Now()

	for folder, files := range archives {
		for _, archiveName := range files {
			count++
			log.Printf("==> Extracting Archive (%d/%d): %s", count, total, archiveName)

			fSize, files, duration, err := job.processArchive(folder, archiveName)
			if err != nil {
				log.Printf("[ERROR] Extracting: %v", err)
			} else {
				log.Printf("==> Extracted Archive %s in %v: bytes: %d, files: %d",
					archiveName, duration.Round(time.Millisecond), fSize, len(files))
			}

			if len(files) > 0 {
				log.Printf("==> Files:\n - %s", strings.Join(files, "\n - "))
			}

			size += fSize
			fCount += len(files)
		}
	}

	log.Printf("==> Done.\n==> Extracted %d archives; wrote %d files totalling %d bytes in %v",
		total, fCount, size, time.Since(start).Round(time.Millisecond))
}

func (j *Job) processArchive(folder, archiveName string) (int64, []string, time.Duration, error) {
	file := &xtractr.XFile{
		FilePath:   archiveName,       // Path to archive being extracted.
		OutputDir:  j.Output,          // Folder to extract archive into.
		FileMode:   j.FileMode.Mode(), // Write files with this mode.
		DirMode:    j.DirMode.Mode(),  // Write folders with this mode.
		Passwords:  j.Passwords,       // (RAR/7zip) Archive password(s).
		SquashRoot: j.SquashRoot,      // Remove single root folder?
	}
	file.SetLogger(j)

	// If preserving the file hierarchy: set the output directory to the same path as the input file.
	if j.Preserve {
		// Remove input path prefix from fileName,
		// append fileName.Dir to job.Output,
		// extract file into job.Output/file(sub)Folder(s).
		file.OutputDir = filepath.Join(j.Output, filepath.Dir(strings.TrimPrefix(folder, archiveName)))
	}

	start := time.Now()

	size, files, _, err := xtractr.ExtractFile(file)
	if err != nil {
		return size, files, time.Since(start), fmt.Errorf("archive: %s: %w", archiveName, err)
	}

	return size, files, time.Since(start), nil
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

		for folder, fileList := range xtractr.FindCompressedFiles(xtractr.Filter{
			Path:          fileName,
			ExcludeSuffix: exclude,
			MaxDepth:      int(j.MaxDepth),
			MinDepth:      int(j.MinDepth),
		}) {
			archives[folder] = fileList
		}
	}

	return archives
}
