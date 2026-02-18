package processor

import (
	"errors"
	"fmt"
	"io"

	"github.com/en666ki/gavro/internal/filter"
	"github.com/en666ki/gavro/internal/reader"
	"github.com/en666ki/gavro/internal/writer"
)

// Processor coordinates reading, filtering, transforming, and writing records.
type Processor struct {
	reader    reader.Reader
	writer    writer.Writer
	filter    filter.Filter
	transform func(reader.Record) (interface{}, error)
}

// Option configures a Processor.
type Option func(*Processor)

// WithFilter sets a filter that records must satisfy to be written.
func WithFilter(f filter.Filter) Option {
	return func(p *Processor) {
		p.filter = f
	}
}

// WithTransform sets a function that transforms each record before writing.
func WithTransform(t func(reader.Record) (interface{}, error)) Option {
	return func(p *Processor) {
		p.transform = t
	}
}

// NewProcessor creates a Processor that reads from r and writes to w.
func NewProcessor(r reader.Reader, w writer.Writer, opts ...Option) *Processor {
	p := &Processor{
		reader: r,
		writer: w,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Process reads all records and writes them to the writer.
func (p *Processor) Process() error {
	return p.ProcessWithLimit(0)
}

// ProcessWithLimit reads up to limit records. A limit of 0 means no limit.
func (p *Processor) ProcessWithLimit(limit int) error {
	count := 0
	for {
		if limit > 0 && count >= limit {
			break
		}

		record, err := p.reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		if p.filter != nil {
			matches, err := p.filter.Matches(record)
			if err != nil {
				return fmt.Errorf("filter error: %w", err)
			}
			if !matches {
				continue
			}
		}

		var output interface{} = record
		if p.transform != nil {
			output, err = p.transform(record)
			if err != nil {
				return fmt.Errorf("transform error: %w", err)
			}
		}

		if err := p.writer.Write(output); err != nil {
			return fmt.Errorf("write error: %w", err)
		}

		count++
	}

	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("flush error: %w", err)
	}

	return nil
}
