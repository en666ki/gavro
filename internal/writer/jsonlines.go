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
}

// NewJSONLinesWriter создает новый JSON Lines writer
func NewJSONLinesWriter(out io.Writer) *JSONLinesWriter {
	return &JSONLinesWriter{
		writer: bufio.NewWriter(out),
	}
}

// Write записывает одну запись как одну строку JSON
func (w *JSONLinesWriter) Write(record map[string]interface{}) error {
	// Сериализуем в JSON
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

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
