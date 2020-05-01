package errors

import (
	"fmt"

	"github.com/atosatto/ansible-requirements-lint/pkg/types"
)

// UnknownScmError is returned when the
// the value of the Scm field is not valid or
// correspondes to an unknown Source Code Management system.
type UnknownScmError struct {
	scm string
}

// NewUnknownScmError creates a new UnknownScmError
func NewUnknownScmError(scm string) *UnknownScmError {
	return &UnknownScmError{scm: scm}
}

// Error converts an UnknownScmError to string
func (e *UnknownScmError) Error() string {
	return fmt.Sprintf("unknown or unsupported scm %s", e.scm)
}

// IsUnknownScmError checks whether nil is an UnknownScmError
func IsUnknownScmError(err error) bool {
	if _, ok := err.(*UnknownScmError); ok {
		return true
	}
	return false
}

// RoleNotFoundError is returned when it is not possible
// to find the role on the upstream sourcsourcee.
type RoleNotFoundError struct {
	role   types.Role
	source string
}

// NewRoleNotFoundError creates a new RoleNotFoundError
func NewRoleNotFoundError(role types.Role, source string) *RoleNotFoundError {
	return &RoleNotFoundError{role: role, source: source}
}

// Error converts a RoleNotFoundError to string
func (e *RoleNotFoundError) Error() string {
	return fmt.Sprintf("unable to find role %s on %s", e.role.Name, e.source)
}

// IsRoleNotFoundError checks whether nil is a RoleNotFoundError
func IsRoleNotFoundError(err error) bool {
	if _, ok := err.(*RoleNotFoundError); ok {
		return true
	}
	return false
}

// RoleVersionNotFoundError is returned when
// the current role version is not found in the list of the
// available ones.
type RoleVersionNotFoundError struct {
	role      types.Role
	available []string
}

// NewRoleVersionNotFoundError creates a new RoleVersionNotFoundError
func NewRoleVersionNotFoundError(role types.Role, available []string) *RoleVersionNotFoundError {
	return &RoleVersionNotFoundError{role: role, available: available}
}

// Error converts a RoleVersionsNotFoundError to string
func (e *RoleVersionNotFoundError) Error() string {
	return fmt.Sprintf("unable to find version %s for role %s in %v", e.role.Version, e.role.Name, e.available)
}

// IsRoleVersionNotFoundError checks whether nil is a RoleVersionNotFoundError
func IsRoleVersionNotFoundError(err error) bool {
	if _, ok := err.(*RoleVersionNotFoundError); ok {
		return true
	}
	return false
}
