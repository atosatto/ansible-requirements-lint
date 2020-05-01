package provider

import (
	"context"

	"github.com/atosatto/ansible-requirements-lint/pkg/types"
)

// The RolesProvider interface define some methods to fetch
// information on Ansible Roles updates from upstream provider
// such as Git and Ansible Galaxy.
type RolesProvider interface {
	VersionsForRole(ctx context.Context, r types.Role) ([]string, error)
}
