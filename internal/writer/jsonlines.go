package writer

import (
	"bufio"
	"encoding/json"
	"io"
)

// JSONLinesWriter writes records in JSON Lines format (one JSON value per line).
type JSONLinesWriter struct {
	encoder *json.Encoder
	writer  *bufio.Writer
	pretty  bool
	first   bool
}

// NewJSONLinesWriter returns a JSONLinesWriter that writes to out.
// When pretty is true, each record is indented and separated by a blank line.
func NewJSONLinesWriter(out io.Writer, pretty bool) *JSONLinesWriter {
	bw := bufio.NewWriter(out)
	enc := json.NewEncoder(bw)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return &JSONLinesWriter{
		encoder: enc,
		writer:  bw,
		pretty:  pretty,
		first:   true,
	}
}

// Write encodes record as JSON and writes it to the output.
func (w *JSONLinesWriter) Write(record interface{}) error {
	if w.pretty && !w.first {
		if err := w.writer.WriteByte('\n'); err != nil {
			return err
		}
	}
	w.first = false

	return w.encoder.Encode(record)
}

// Flush flushes the underlying buffered writer.
func (w *JSONLinesWriter) Flush() error {
	return w.writer.Flush()
}
