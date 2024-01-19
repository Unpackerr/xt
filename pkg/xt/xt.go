package xt

import (
	"log"
	"os"
	"strings"
	"time"

	"golift.io/xtractr"
)

type Job struct {
	Paths  []string
	Output string
}

func Extract(job *Job) {
	archives := getArchives(job.Paths)
	if len(archives) == 0 {
		log.Println("==> No archives found in:", job.Paths)
	}

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

			size, files, _, err := xtractr.ExtractFile(&xtractr.XFile{
				FilePath:  fileName,   // Path to archive being extracted.
				OutputDir: job.Output, // Folder to extract archive into.
				FileMode:  0o644,      //nolint:gomnd // Write files with this mode.
				DirMode:   0o755,      //nolint:gomnd // Write folders with this mode.
				Password:  "",         // (RAR) Archive password. Blank for none.
			})
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

func getArchives(paths []string) map[string][]string {
	archives := map[string][]string{}

	for _, fileName := range paths {
		fileInfo, err := os.Stat(fileName)
		if err != nil {
			log.Println("[ERROR] Reading archive path:", err)
			continue
		}

		if !fileInfo.IsDir() {
			archives[fileName] = []string{fileName}
			continue
		}

		for k, v := range xtractr.FindCompressedFiles(xtractr.Filter{Path: fileName}) {
			archives[k] = v
		}
	}

	return archives
}
