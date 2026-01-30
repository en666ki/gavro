package processor

import (
	"fmt"
	"io"

	"gavro/internal/reader"
	"gavro/internal/writer"
)

// Processor координирует чтение и запись данных
type Processor struct {
	reader reader.Reader
	writer writer.Writer
}

// NewProcessor создает новый процессор
func NewProcessor(r reader.Reader, w writer.Writer) *Processor {
	return &Processor{
		reader: r,
		writer: w,
	}
}

// Process читает все записи и записывает их
func (p *Processor) Process() error {
	for {
		record, err := p.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		if err := p.writer.Write(record); err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}

	// Важно: flush в конце
	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}

	return nil
}

// ProcessWithLimit читает ограниченное количество записей (для будущих флагов)
func (p *Processor) ProcessWithLimit(limit int) error {
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

		if err := p.writer.Write(record); err != nil {
			return fmt.Errorf("write error: %w", err)
		}

		count++
	}

	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}

	return nil
}
