package main

import (
	"flag"
	"log"
	"os"

	"github.com/Unpackerr/xt/pkg/xt"
)

func main() {
	log.SetFlags(0)

	jobs := parseJobs()
	if len(jobs) < 1 || len(jobs[0].Paths) < 1 {
		log.Printf("If you pass a directory, this app will extract every archive in it.")
		log.Fatalf("Usage: %s [-output <path>] <path> [paths...]", os.Args[0])
	}

	for i, job := range jobs {
		log.Printf("Starting Job %d with %d paths, output: %s", i+1, len(job.Paths), job.Output)
		xt.Extract(job)
	}
}

func parseJobs() []*xt.Job {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}

	output := flag.String("output", pwd, "Output directory, default is current directory")

	flag.Parse()

	return []*xt.Job{{Output: *output, Paths: flag.Args()}}
}
