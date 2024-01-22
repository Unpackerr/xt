//nolint:forbidigo
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Unpackerr/xt/pkg/xt"
	flag "github.com/spf13/pflag"
	"golift.io/version"
)

func main() {
	log.SetFlags(0)

	jobs := parseJobs()
	if len(jobs) < 1 || len(jobs[0].Paths) < 1 {
		flag.Usage()
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

	output := flag.StringP("output", "o", pwd, "Output directory, default is current directory")
	printVer := flag.BoolP("version", "v", false, "Print application version and exit")

	flag.Usage = func() {
		fmt.Println("If you pass a directory, this app will extract every archive in it.")
		fmt.Printf("Usage: %s [-v] [--output <path>] <path> [paths...]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	flag.Parse()

	if *printVer {
		fmt.Printf("xt v%s\n", version.Version)
		os.Exit(0)
	}

	return []*xt.Job{{Output: *output, Paths: flag.Args()}}
}
