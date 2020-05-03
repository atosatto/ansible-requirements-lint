package writer

import (
	"github.com/atosatto/ansible-requirements-lint/pkg/linter"
	"github.com/atosatto/ansible-requirements-lint/pkg/types"
)

// roleName returns the Name for the Role.
// If no name has been specified in the Role dependency,
// the Source will be returned instead.
func roleName(role types.Role) string {
	if len(role.Name) == 0 {
		return role.Source
	}
	return role.Name
}

// metadataToUpdate converts the metadata of the given
// Result to Update. If the metadata are not an Update
// or are nil, an empty Update struct will be returned.
func metadataToUpdate(res linter.Result) linter.Update {
	if res.Metadata == nil {
		return linter.Update{}
	}

	update, ok := res.Metadata.(linter.Update)
	if !ok {
		return linter.Update{}
	}
	return update
}
