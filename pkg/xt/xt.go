// Package xt provides the interface to extract archive files based on job parameters.
package xt

import (
	"log"
	"os"
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

	total := 0
	count := 0

	for _, files := range archives {
		total += len(files)
	}

	for _, files := range archives {
		for _, fileName := range files {
			count++
			log.Printf("==> Extracting Archive (%d/%d): %s", count, total, fileName)

			file := &xtractr.XFile{
				FilePath:   fileName,            // Path to archive being extracted.
				OutputDir:  job.Output,          // Folder to extract archive into.
				FileMode:   job.FileMode.Mode(), // Write files with this mode.
				DirMode:    job.DirMode.Mode(),  // Write folders with this mode.
				Passwords:  job.Passwords,       // (RAR/7zip) Archive password(s).
				SquashRoot: job.SquashRoot,      // Remove single root folder?
			}

			file.SetLogger(job)

			start := time.Now()

			size, files, _, err := xtractr.ExtractFile(file)
			if err != nil {
				log.Printf("[ERROR] Archive: %s: %v", fileName, err)
				continue
			}

			log.Printf("==> Extracted Archive %s in %v: bytes: %d, files: %d",
				fileName, time.Since(start).Round(time.Millisecond), size, len(files))

			if len(files) > 0 {
				log.Printf("==> Files:\n - %s", strings.Join(files, "\n - "))
			}
		}
	}
}

func (j *Job) getArchives() map[string][]string {
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
