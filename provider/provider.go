package provider

import (
	"context"

	"github.com/atosatto/ansible-requirements-lint/requirements"
)

// The RolesProvider interface define some methods to fetch
// information on Ansible Roles updates from upstream provider
// such as Git and Ansible Galaxy.
type RolesProvider interface {
	VersionsForRole(ctx context.Context, r requirements.Role) ([]string, error)
}
