package types

// Requirements represents the content
// Ansible Requirements file.
type Requirements struct {
	// Roles is the list of roles defined
	// in the Requirements file.
	Roles []Role

	// Children is the list of requirements
	// files included by the Requirement file.
	Childrens []*Requirements
}

// Role is an Ansible Role definition.
// If the value of the Include is different
// from the nil string, it means that
// the Role represents an Include of another
// requirement file. In this case all
// the other fields must be set to the nil string.
type Role struct {
	Source  string
	Scm     string
	Version string
	Name    string

	Include string
}
