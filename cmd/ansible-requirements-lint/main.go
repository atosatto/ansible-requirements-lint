package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

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
		errAndExit(fmt.Sprintf("ansible-galaxy-lint v%s", version))
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

	// handle Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// call cancel on the context
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

	// run the updates linter
	updatesLinterOpts := linter.UpdatesLinterOptions{
		AnsibleGalaxyURL: *galaxyURL,
	}
	numUpdatesOrErrs := runUpdatesLinter(ctx, r, updatesLinterOpts)

	if numUpdatesOrErrs > 0 {
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
