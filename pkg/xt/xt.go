package xt

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golift.io/xtractr"
)

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

			start := time.Now()

			xFile := &xtractr.XFile{
				FilePath:  fileName,            // Path to archive being extracted.
				OutputDir: job.Output,          // Folder to extract archive into.
				FileMode:  job.FileMode.Mode(), // Write files with this mode.
				DirMode:   job.DirMode.Mode(),  // Write folders with this mode.
				Passwords: job.Passwords,       // (RAR/7zip) Archive password(s).
			}

			// If preserving the file hierarchy, set the output directory to the
			// folder of the archive being extracted.
			if job.Preserve {
				xFile.OutputDir = filepath.Dir(fileName)
			}

			size, files, _, err := xtractr.ExtractFile(xFile)
			if err != nil {
				log.Printf("[ERROR] Archive: %s: %v", fileName, err)
				continue
			}

			elapsed := time.Since(start).Round(time.Millisecond)
			log.Printf("==> Extracted Archive %s in %v: bytes: %d, files: %d", fileName, elapsed, size, len(files))
			log.Printf("==> Files:\n - %s", strings.Join(files, "\n - "))
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
