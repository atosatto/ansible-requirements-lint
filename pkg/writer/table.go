package writer

import (
	"context"
	"fmt"
	"io"

	"github.com/atosatto/ansible-requirements-lint/pkg/errors"
	"github.com/atosatto/ansible-requirements-lint/pkg/linter"
	"github.com/olekukonko/tablewriter"
)

// TableWriter formats the Linters output in an ASCII table format
type TableWriter struct {
	Verbose bool
}

// WriteUpdates write Linters results in ASCII table format to the given io.Writer
func (t TableWriter) WriteUpdates(ctx context.Context, w io.Writer, input <-chan linter.Result) error {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Name", "Current Version", "Latest Version", "Status"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	defer table.Render()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, more := <-input:
			// return if there are no more input results
			if !more {
				return nil
			}

			// get the role name and the Update metadata
			var roleName = roleName(res.Role)
			var meta = metadataToUpdate(res)

			// print the text formatted message
			if res.Level != linter.LevelInfo || t.Verbose {
				switch {
				case errors.IsRoleVersionNotFoundError(res.Err) && res.Role.Version != "":
					// we cannot determine whether there is any update cause we haven't found
					// the version among the one available for the role
					table.Append([]string{roleName, res.Role.Version, meta.ToVersion, "Update"})
				case errors.IsRoleVersionNotFoundError(res.Err):
					// the user has not defined any version for the role
					table.Append([]string{roleName, "-", meta.ToVersion, "Update"})
				case errors.IsRoleNotFoundError(res.Err):
					table.Append([]string{roleName, "-", "-", "Role Not Found"})
				case res.Err != nil:
					// there have been an error fetching for the version
					table.Append([]string{roleName, "-", "-", fmt.Sprintf("Error: %v", res.Err)})
				case meta.IsUpdate:
					// there is an update for the role
					table.Append([]string{roleName, res.Role.Version, meta.ToVersion, "Update"})
				default:
					// the role is at the latest version
					table.Append([]string{roleName, res.Role.Version, meta.ToVersion, "Ok"})
				}
			}
		}
	}
}
