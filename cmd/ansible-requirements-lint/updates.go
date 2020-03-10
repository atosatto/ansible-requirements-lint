package main

import (
	"context"
	"fmt"

	"github.com/atosatto/ansible-requirements-lint/linter"
	"github.com/atosatto/ansible-requirements-lint/requirements"
	"github.com/fatih/color"
)

// runUpdatesLinter updates the Roles defined in the Requirements r.
// The total number of updates and errors detected by the updatesLinter is returned as result of the function execution.
func runUpdatesLinter(ctx context.Context, r *requirements.Requirements, updatesLinterOpts linter.UpdatesLinterOptions) int {
	rolesChan := make(chan requirements.Role)
	resultsChan := make(chan linter.Result)

	updatesLinter := linter.NewUpdatesLinter(updatesLinterOpts)
	go updatesLinter.RunWithContext(ctx, rolesChan, resultsChan)

	var numUpdatesOrErr = 0
	for _, role := range r.Roles {
		rolesChan <- *role
		result := <-resultsChan

		switch {
		case result.Level == linter.LevelInfo && *verbose:
			color.New(color.Bold, color.FgHiCyan).Printf("INFO: ")
			fmt.Printf("%s.\n", result.Msg)
		case result.Level == linter.LevelWarning:
			color.New(color.Bold, color.FgHiYellow).Printf("WARN: ")
			fmt.Printf("%s.\n", result.Msg)
			numUpdatesOrErr++
		case result.Level == linter.LevelError:
			color.New(color.Bold, color.FgHiRed).Printf("ERR: ")
			fmt.Printf("%v.\n", result.Err)
			numUpdatesOrErr++
		}
	}

	return numUpdatesOrErr
}
