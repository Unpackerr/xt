//nolint:forbidigo
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Unpackerr/xt/pkg/xt"
	flag "github.com/spf13/pflag"
	"golift.io/version"
	"golift.io/xtractr"
)

func parseFlags(pwd string) (*xt.Job, *flags) {
	flag.Usage = func() {
		// XXX: Write more "help" info here.
		fmt.Println("If you pass a directory, this app will extract every archive in it.")
		fmt.Printf("Usage: %s [-v] [--output <path>] <path> [paths...]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	job := &xt.Job{}
	flags := &flags{}

	flag.BoolVarP(&flags.PrintVer, "version", "v", false, "Print application version and exit")
	// These cli options create 1 job. Using job files creates N jobs.
	flag.StringVarP(&job.Output, "output", "o", pwd, "Output directory, default is current directory")
	flag.UintVarP(&job.MaxDepth, "max-depth", "d", 0, "Maximum folder depth to recursively search for archives.")
	flag.UintVarP(&job.MinDepth, "min-depth", "m", 0, "Minimum folder depth to recursively search for archives.")
	flag.StringSliceVarP(&job.Include, "extension", "e", nil, "Only extract files with these extensions.")
	flag.StringSliceVarP(&job.Passwords, "password", "P", nil, "Attempt these passwords for rar and 7zip archives.")
	flag.StringSliceVarP(&flags.JobFiles, "job-file", "j", nil, "Read additional extraction jobs from these files.")
	flag.BoolVarP(&job.Preserve, "preserve-paths", "p", false, "Recreate directory hierarchy while extracting.")
	flag.Parse()
	// flag.UintVarP(&job.Recurse, "recurse", "r", 0, "Extract archives inside archives, up to this depth.")

	job.Paths = flag.Args()

	return job, flags
}

// flags contains the non-job flags used on the cli.
type flags struct {
	PrintVer bool
	JobFiles []string
}

func main() {
	// Where we extract to.
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}

	// Get 1 job and other flag info from cli args.
	cliJob, flags := parseFlags(pwd)
	flags.printVer()

	// Read in jobs from 1 or more job files.
	jobs, err := xt.ParseJobs(flags.JobFiles)
	if err != nil {
		log.Fatal("[ERROR]", err)
	}

	// Append cli job to job file jobs.
	if len(cliJob.Paths) > 0 {
		jobs = append(jobs, cliJob)
	}

	// Check for jobs?
	if len(jobs) < 1 || len(jobs[0].Paths) < 1 {
		flag.Usage()
	}

	// Extract the jobs.
	for i, job := range jobs {
		log.Printf("Starting Job %d of %d with %s", i+1, len(jobs), job)
		xt.Extract(job)
	}
}

// printVer prints the version and exits.
// Only if the user passed -v or --version.
func (f *flags) printVer() {
	if !f.PrintVer {
		return
	}

	fmt.Printf("xt v%s-%s (%s)\n", version.Version, version.Revision, version.Branch)
	fmt.Println(" - Supported Extensions:")

	for _, ext := range xtractr.SupportedExtensions() {
		fmt.Println("  ", ext)
	}

	os.Exit(0)
}
