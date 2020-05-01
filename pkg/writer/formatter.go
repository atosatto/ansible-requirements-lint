package writer

import (
	"context"
	"io"

	"github.com/atosatto/ansible-requirements-lint/pkg/linter"
)

// Writer writes Linters results to
// a given output in a structured format
type Writer interface {
	WriteUpdates(ctx context.Context, output io.Writer, input <-chan linter.Result) error
}
