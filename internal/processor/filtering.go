package processor

import (
	"fmt"
	"io"

	"github.com/en666ki/gavro/internal/reader"
	"github.com/en666ki/gavro/internal/writer"
)

// Filter интерфейс для фильтрации записей
type Filter interface {
	Matches(record map[string]interface{}) (bool, error)
}

// FilteringProcessor обрабатывает записи с фильтрацией
type FilteringProcessor struct {
	reader reader.Reader
	writer writer.Writer
	filter Filter
}

// NewFilteringProcessor создает процессор с фильтром
func NewFilteringProcessor(r reader.Reader, w writer.Writer, f Filter) *FilteringProcessor {
	return &FilteringProcessor{
		reader: r,
		writer: w,
		filter: f,
	}
}

// Process читает записи, фильтрует и пишет совпадающие
func (p *FilteringProcessor) Process() error {
	for {
		record, err := p.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		// Применяем фильтр
		matches, err := p.filter.Matches(record)
		if err != nil {
			return fmt.Errorf("filter error: %w", err)
		}

		// Пишем только если совпало
		if matches {
			if err := p.writer.Write(record); err != nil {
				return fmt.Errorf("write error: %w", err)
			}
		}
	}

	// Flush в конце
	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}

	return nil
}

// ProcessWithLimit обрабатывает с ограничением результатов
func (p *FilteringProcessor) ProcessWithLimit(limit int) error {
	count := 0
	for {
		if limit > 0 && count >= limit {
			break
		}

		record, err := p.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		// Применяем фильтр
		matches, err := p.filter.Matches(record)
		if err != nil {
			return fmt.Errorf("filter error: %w", err)
		}

		// Пишем только если совпало
		if matches {
			if err := p.writer.Write(record); err != nil {
				return fmt.Errorf("write error: %w", err)
			}
			count++
		}
	}

	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}

	return nil
}
