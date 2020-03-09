package linter

import (
	"context"

	"github.com/atosatto/ansible-requirements-lint/requirements"
)

// Level indicates the importance of results
// returned by implementations of the RolesLinter interface.
type Level string

const (
	// LevelInfo must be used for results that can be safely
	// ignore by the caller of the Linter.
	LevelInfo = Level("INFO")

	// LevelWarning must be used to indicated result
	// that must be brough to the attention of the
	// caller of the Linter.
	LevelWarning = Level("WARN")

	// LevelError is used to signal errors to the
	// caller of the Linter.
	LevelError = Level("ERR")
)

// Result holds the RolesLinter evaluation
// on a given role dependency declaration.
type Result struct {
	// The role evaluated by the Linter.
	Role requirements.Role

	// The level of importance of the
	// Linter evaluation.
	Level Level

	// A textual message describing
	// the Linter evaluation of the
	// role.
	Msg string

	// The detailed error returned
	// by the Linter for propagation
	// to the caller.
	// When Err != nil the level should
	// be set to LevelError.
	Err error

	// Optional implementation specific
	// Linter Metadata.
	Metadata interface{}
}

// A RolesLinter analyzes Ansible roles declarations
// and provides linting feedback in form of Result.
type RolesLinter interface {
	RunWithContext(ctx context.Context, roles <-chan requirements.Role, results chan<- Result)
}
