package writer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// JSONLinesWriter записывает записи в формате JSON Lines
type JSONLinesWriter struct {
	writer *bufio.Writer
	pretty bool
	first  bool
}

// NewJSONLinesWriter создает новый JSON Lines writer
func NewJSONLinesWriter(out io.Writer, pretty bool) *JSONLinesWriter {
	return &JSONLinesWriter{
		writer: bufio.NewWriter(out),
		pretty: pretty,
		first:  true,
	}
}

// Write записывает одну запись как одну строку JSON
func (w *JSONLinesWriter) Write(record map[string]interface{}) error {
	var data []byte
	var err error

	// Сериализуем в JSON
	if w.pretty {
		data, err = json.MarshalIndent(record, "", "  ")
	} else {
		data, err = json.Marshal(record)
	}

	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	// В pretty режиме добавляем пустую строку между записями
	if w.pretty && !w.first {
		if err := w.writer.WriteByte('\n'); err != nil {
			return err
		}
	}
	w.first = false

	// Записываем строку + newline
	if _, err := w.writer.Write(data); err != nil {
		return err
	}

	if err := w.writer.WriteByte('\n'); err != nil {
		return err
	}

	return nil
}

// Flush сбрасывает буфер
func (w *JSONLinesWriter) Flush() error {
	return w.writer.Flush()
}
