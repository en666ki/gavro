package filter

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

// CELFilter evaluates CEL expressions against records.
type CELFilter struct {
	program cel.Program
}

// NewCELFilter compiles a CEL expression and returns a filter that evaluates it.
func NewCELFilter(expression string) (*CELFilter, error) {
	env, err := cel.NewEnv(
		cel.Variable("record", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", issues.Err())
	}

	program, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL program: %w", err)
	}

	return &CELFilter{
		program: program,
	}, nil
}

// Matches reports whether the record satisfies the CEL expression.
func (f *CELFilter) Matches(record map[string]interface{}) (bool, error) {
	result, _, err := f.program.Eval(map[string]interface{}{
		"record": record,
	})

	if err != nil {
		return false, fmt.Errorf("evaluation error: %w", err)
	}

	boolResult, ok := result.(types.Bool)
	if !ok {
		return false, fmt.Errorf("expression must return boolean, got %T", result)
	}

	return bool(boolResult), nil
}

// Filter is implemented by types that can match records.
type Filter interface {
	Matches(record map[string]interface{}) (bool, error)
}

// Compile-time check
var _ Filter = (*CELFilter)(nil)
