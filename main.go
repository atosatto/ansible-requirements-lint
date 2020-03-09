package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/atosatto/ansible-requirements-lint/linter"
	"github.com/atosatto/ansible-requirements-lint/provider"
	"github.com/atosatto/ansible-requirements-lint/requirements"
	"github.com/fatih/color"
)

var (
	verbose      = flag.Bool("v", false, "")
	galaxyURL    = flag.String("galaxy", provider.DefaultAnsibleGalaxyURL, "")
	noColor      = flag.Bool("no-color", false, "")
	printVersion = flag.Bool("V", false, "")
	printHelp    = flag.Bool("h", false, "")
)

var version string

var usage = fmt.Sprintf(`Usage: ansible-requirements-lint [options...] <requirements-file>

Options:
  -v             Enable verbose output.
  -galaxy <url>  Set the Ansible Galaxy URL (default: %s).
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
		errAndExit(fmt.Sprintf("ansible-galaxy-lint %s", version))
	}

	if flag.NArg() != 1 || *printHelp {
		usageAndExit("")
	}

	// disables colorized output
	if *noColor {
		color.NoColor = true
	}

	var requirementsFile = flag.Arg(0)
	if len(requirementsFile) == 0 {
		usageAndExit("")
	}

	r, err := requirements.UnmarshalFromFile(requirementsFile)
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

	ctx := context.Background()
	rolesChan := make(chan requirements.Role)
	resultsChan := make(chan linter.Result)

	updatesLinterOpts := linter.UpdatesLinterOptions{
		AnsibleGalaxyURL: *galaxyURL,
	}
	updatesLinter := linter.NewUpdatesLinter(updatesLinterOpts)
	go updatesLinter.RunWithContext(ctx, rolesChan, resultsChan)

	var errOrWarn = false
	for _, role := range r.Roles {
		rolesChan <- role
		result := <-resultsChan

		switch {
		case result.Level == linter.LevelInfo && *verbose:
			color.New(color.Bold, color.FgHiCyan).Printf("INFO: ")
			fmt.Printf("%s.\n", result.Msg)
		case result.Level == linter.LevelWarning:
			color.New(color.Bold, color.FgHiYellow).Printf("WARN: ")
			fmt.Printf("%s.\n", result.Msg)
			errOrWarn = true
		case result.Level == linter.LevelError:
			color.New(color.Bold, color.FgHiRed).Printf("ERR: ")
			fmt.Printf("%v.\n", result.Err)
			errOrWarn = true
		}
	}

	if errOrWarn {
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
