//nolint:forbidigo
package main

import (
	"fmt"
	"log"
	"os"

	flag "github.com/spf13/pflag"
	"golift.io/version"
	"golift.io/xt/pkg/xt"
	"golift.io/xtractr"
)

func parseFlags(pwd string) (*xt.Job, *flags) {
	flag.Usage = func() {
		fmt.Println("A folder is always searched, including its sub folders.")
		fmt.Println("--max-depth and --min-depth limit that search. --recurse is not used for it.")
		fmt.Println("--max-depth 1 searches the path and one level of sub folders. 0 is unlimited.")
		fmt.Println("An archive inside an archive stays packed.")
		fmt.Println("--recurse adds one second pass over what was just extracted, and it runs once.")
		fmt.Println("Example: pass 1 extracts bundle.zip and writes disc.img.zst;")
		fmt.Println("pass 2 extracts that and writes disc.img; disc.img stays packed.")
		fmt.Printf("Usage: %s [-v] [--output <path>] <path> [paths...]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	job := &xt.Job{}
	flags := &flags{}

	flag.BoolVarP(&flags.PrintVer, "version", "v", false, "Print application version, and supported extensions.")
	// These cli options create 1 job. Using job files creates N jobs.
	flag.StringVarP(&job.Output, "output", "o", pwd, "Output directory, default is current directory.")
	flag.Uint16VarP(&job.MaxDepth, "max-depth", "d", 0,
		"Folder search depth. 0 is unlimited. 1 searches the path and one level of sub folders.")
	flag.Uint16VarP(&job.MinDepth, "min-depth", "m", 0, "Minimum folder depth to recursively search for archives.")
	flag.StringSliceVarP(&job.Include, "extension", "e", nil, "Only extract files with these extensions.")
	flag.StringSliceVarP(&job.Passwords, "password", "P", nil, "Attempt these passwords for rar and 7zip archives.")
	flag.BoolVarP(&job.SquashRoot, "squash-root", "S", false,
		"If archive contains only 1 folder at in the root, its contents are moved into output folder.")
	flag.BoolVarP(&job.Verbose, "verbose", "V", false, "Verbose output prints the file list that was extracted.")
	flag.BoolVarP(&job.DebugLog, "debug", "D", false, "Enable debug output.")
	flag.StringSliceVarP(&flags.JobFiles, "job-file", "j", nil, "Read additional extraction jobs from these files.")
	flag.BoolVarP(&job.Preserve, "preserve-paths", "p", false, "Recreate directory hierarchy while extracting.")
	flag.BoolVarP(&job.Recurse, "recurse", "r", false,
		"One extra pass over archives just extracted. Not used for the folder search.")
	flag.Parse()

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

	log.SetOutput(os.Stdout)

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
