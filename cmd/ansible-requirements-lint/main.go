package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/atosatto/ansible-requirements-lint/pkg/linter"
	"github.com/atosatto/ansible-requirements-lint/pkg/parser"
	"github.com/atosatto/ansible-requirements-lint/pkg/provider"
	"github.com/atosatto/ansible-requirements-lint/pkg/writer"
)

var (
	verbose      = flag.Bool("v", false, "")
	galaxyURL    = flag.String("galaxy", provider.DefaultAnsibleGalaxyURL, "")
	noColor      = flag.Bool("no-color", false, "")
	outFormat    = flag.String("o", "text", "")
	printVersion = flag.Bool("V", false, "")
	printHelp    = flag.Bool("h", false, "")
)

// version will be set at compilation time
var version string

var usage = fmt.Sprintf(`Usage: ansible-requirements-lint [options...] <requirements-file>

Options:
  -v             Enable verbose output.
  -galaxy <url>  Set the Ansible Galaxy URL (default: %s).
  -o <format>    Format of the output, allowed values are text,table (default: text).
  -no-color      Disable color output.
  -V             Print the version number and exit.
  -h             Show this help message and exit.
`, provider.DefaultAnsibleGalaxyURL)

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	// print the version
	if *printVersion {
		errAndExit(fmt.Sprintf("ansible-galaxy-lint v%s", version))
	}

	if flag.NArg() != 1 || *printHelp {
		usageAndExit("")
	}

	var out writer.Writer
	switch *outFormat {
	case "table":
		out = writer.TableWriter{
			Verbose: *verbose,
		}
	default:
		out = writer.TextWriter{
			Verbose: *verbose,
			NoColor: *noColor,
		}
	}

	var requirementsFile = flag.Arg(0)
	if len(requirementsFile) == 0 {
		usageAndExit("")
	}

	// handle Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// call cancel on the context in
	// order to prevent go routines from leaking
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	// parse the requirements file
	requirements, err := parser.UnmarshalFromFile(requirementsFile)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			errAndExit(fmt.Sprintf("unable to open %s: the file does not exists", requirementsFile))
		case os.IsPermission(err):
			errAndExit(fmt.Sprintf("unable to open %s: please check file permissions", requirementsFile))
		default:
			errAndExit(fmt.Sprintf("unable to parse the requirements file: %s", err))
		}
	}

	// create a WaitGroup to make sure
	// to wait for the Linters to finish before
	// exiting the program
	var wg sync.WaitGroup

	// whether to exit with an error code
	// or not
	var exitWithError = false

	// run the Updates Linter
	wg.Add(2)
	updatesLinterResults := make(chan linter.Result)
	updatesLinterOutput := make(chan linter.Result)
	go func() {
		updatesLinter := linter.NewUpdatesLinter()
		updatesLinter.WithAnsibleGalaxyURL(*galaxyURL)
		updatesLinter.Lint(ctx, requirements, updatesLinterResults)
		defer wg.Done()
	}()
	go func() {
		out.WriteUpdates(ctx, os.Stdout, updatesLinterOutput)
		defer wg.Done()
	}()

	// check wether the Updates Linter has reported
	// any Error or Warning and copy back the results
	// to the output channel
	func() {
		for update := range updatesLinterResults {
			select {
			case <-ctx.Done():
				return
			default:
				if update.Level != linter.LevelInfo {
					exitWithError = true
				}
				updatesLinterOutput <- update
			}
		}
	}()
	close(updatesLinterOutput)

	// wait for the Linters to be done
	wg.Wait()

	if exitWithError {
		os.Exit(1)
	}
}

func errAndExit(msg string) {
	fmt.Fprint(os.Stderr, msg)
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprint(os.Stderr, msg)
		fmt.Fprint(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}
