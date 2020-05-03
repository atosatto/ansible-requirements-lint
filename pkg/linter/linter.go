package linter

import (
	"context"

	"github.com/atosatto/ansible-requirements-lint/pkg/types"
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
	Role types.Role

	// The level of severity of the
	// Linter result.
	Level Level

	// The detailed error returned
	// by the Linter for propagation
	// to the caller.
	// Note that errors should be considered to be
	// blocking only when the Level is set to LevelError.
	Err error

	// Implementation specific
	// Linter Metadata.
	Metadata interface{}
}

// A Linter analyzes Ansible requirements definitions
// and provides linting feedback in form of Result.
type Linter interface {
	Lint(context.Context, *types.Requirements, chan<- Result) error
}
