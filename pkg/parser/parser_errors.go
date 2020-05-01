package parser

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnexpectedNodeKindError is returned
// when the parser encounters an unexpeted yaml.Node kind
// while parsing roles or requirements.
type UnexpectedNodeKindError struct {
	line int
	kind yaml.Kind
}

// NewUnexpectedNodeKindError creates a new UnexpectedNodeKindError
func NewUnexpectedNodeKindError(line int, kind yaml.Kind) *UnexpectedNodeKindError {
	return &UnexpectedNodeKindError{
		line: line,
		kind: kind,
	}
}

func (e *UnexpectedNodeKindError) Error() string {
	return fmt.Sprintf("unexpected yaml node kind at line %d: %d", e.line, e.kind)
}

// UnexpectedMappingNodeValueError is returned
// when parsing a MappingNode, the parser
// encounters an unexpected dictionary key.
type UnexpectedMappingNodeValueError struct {
	line  int
	value string
}

// NewUnexpectedMappingNodeValueError creates a new UnexpectedMappingNodeValueError
func NewUnexpectedMappingNodeValueError(line int, value string) *UnexpectedMappingNodeValueError {
	return &UnexpectedMappingNodeValueError{
		line:  line,
		value: value,
	}
}

func (e *UnexpectedMappingNodeValueError) Error() string {
	return fmt.Sprintf("unexpected yaml dictionary key or value at line %d: %s", e.line, e.value)
}
