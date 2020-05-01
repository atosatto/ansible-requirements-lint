package writer

import (
	"context"
	"fmt"
	"io"

	"github.com/atosatto/ansible-requirements-lint/pkg/errors"
	"github.com/atosatto/ansible-requirements-lint/pkg/linter"
	"github.com/fatih/color"
)

// TextWriter writes Linters results in text format
type TextWriter struct {
	Verbose bool
	NoColor bool
}

// WriteUpdates writes Linters results in Text format to the given io.Writer
func (t TextWriter) WriteUpdates(ctx context.Context, w io.Writer, results <-chan linter.Result) error {
	// disable the colored output
	if t.NoColor {
		color.NoColor = true
	}

	// write the Linter results
	for {
		select {
		case <-ctx.Done():
			return nil
		case res, more := <-results:
			// exit if there are no more input results
			if !more {
				return nil
			}

			// get the role name and the Update metadata
			var roleName = roleName(res.Role)
			var meta = metadataToUpdate(res)

			if res.Level != linter.LevelInfo || t.Verbose {

				// print the Linter level
				switch {
				case res.Level == linter.LevelWarning:
					color.New(color.Bold, color.FgHiYellow).Fprintf(w, "WARN: ")
				case res.Level == linter.LevelError:
					color.New(color.Bold, color.FgHiRed).Fprintf(w, "ERR: ")
				default:
					color.New(color.Bold, color.FgHiCyan).Fprintf(w, "INFO: ")
				}

				// print the Linter result
				switch {
				case errors.IsRoleVersionNotFoundError(res.Err) && res.Role.Version != "":
					fmt.Fprintf(w, "%s: unable to find %s between the available versions for the role, tag a new release or use %s.\n", roleName, res.Role.Version, meta.ToVersion)
				case errors.IsRoleVersionNotFoundError(res.Err):
					fmt.Fprintf(w, "%s: no version specified for the role, pin it to version %s to avoid not explicit dependencies.\n", roleName, meta.ToVersion)
				case res.Err != nil:
					fmt.Fprintf(w, "%s: %v.\n", roleName, res.Err)
				case meta.IsUpdate:
					fmt.Fprintf(w, "%s: role not at the latest version, upgrade from %s to %s.\n", roleName, res.Role.Version, meta.ToVersion)
				default:
					fmt.Fprintf(w, "%s: %s is the latest version for the role, no update needed.\n", roleName, res.Role.Version)
				}
			}
		}
	}
}
