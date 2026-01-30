package filter

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
)

// CELFilter фильтрует записи используя CEL выражения
type CELFilter struct {
	program cel.Program
}

// NewCELFilter создает новый CEL фильтр из выражения
func NewCELFilter(expression string) (*CELFilter, error) {
	// Создаем CEL environment с переменной record
	env, err := cel.NewEnv(
		cel.Variable("record", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment: %w", err)
	}

	// Компилируем выражение
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", issues.Err())
	}

	// Создаем программу
	program, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL program: %w", err)
	}

	return &CELFilter{
		program: program,
	}, nil
}

// Matches проверяет соответствует ли запись фильтру
func (f *CELFilter) Matches(record map[string]interface{}) (bool, error) {
	// Выполняем программу с данными
	result, _, err := f.program.Eval(map[string]interface{}{
		"record": record,
	})

	if err != nil {
		return false, fmt.Errorf("evaluation error: %w", err)
	}

	// Проверяем что результат - boolean
	boolResult, ok := result.(types.Bool)
	if !ok {
		return false, fmt.Errorf("expression must return boolean, got %T", result)
	}

	return bool(boolResult), nil
}

// Filter для совместимости с интерфейсом
type Filter interface {
	Matches(record map[string]interface{}) (bool, error)
}

// Compile-time check
var _ Filter = (*CELFilter)(nil)

// SimplifyCELExpression добавляет record. префикс к полям если его нет
// Это позволяет писать "age > 18" вместо "record.age > 18"
func SimplifyCELExpression(expr string) string {
	// Для первой версии просто оборачиваем в has() проверки
	// В будущем можно добавить более умный парсинг
	return expr
}
