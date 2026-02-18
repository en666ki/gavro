package processor

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/en666ki/gavro/internal/reader"
)

type mockReader struct {
	records []reader.Record
	index   int
}

func (r *mockReader) Read() (reader.Record, error) {
	if r.index >= len(r.records) {
		return nil, io.EOF
	}
	rec := r.records[r.index]
	r.index++
	return rec, nil
}

func (r *mockReader) Close() error { return nil }

type wrappedEOFReader struct {
	records []reader.Record
	index   int
}

func (r *wrappedEOFReader) Read() (reader.Record, error) {
	if r.index >= len(r.records) {
		return nil, fmt.Errorf("stream ended: %w", io.EOF)
	}
	rec := r.records[r.index]
	r.index++
	return rec, nil
}

func (r *wrappedEOFReader) Close() error { return nil }

type errorReader struct {
	records  []reader.Record
	index    int
	errorAt  int
	errorMsg string
}

func (r *errorReader) Read() (reader.Record, error) {
	if r.index == r.errorAt {
		return nil, fmt.Errorf("%s", r.errorMsg)
	}
	if r.index >= len(r.records) {
		return nil, io.EOF
	}
	rec := r.records[r.index]
	r.index++
	return rec, nil
}

func (r *errorReader) Close() error { return nil }

type mockWriter struct {
	records []interface{}
	flushed bool
}

func (w *mockWriter) Write(record interface{}) error {
	w.records = append(w.records, record)
	return nil
}

func (w *mockWriter) Flush() error {
	w.flushed = true
	return nil
}

type errorWriter struct {
	errorAt int
	count   int
}

func (w *errorWriter) Write(record interface{}) error {
	if w.count == w.errorAt {
		return fmt.Errorf("write failed")
	}
	w.count++
	return nil
}

func (w *errorWriter) Flush() error { return nil }

type mockFilter struct {
	matchFn func(record map[string]interface{}) bool
}

func (f *mockFilter) Matches(record map[string]interface{}) (bool, error) {
	return f.matchFn(record), nil
}

type errorFilter struct{}

func (f *errorFilter) Matches(record map[string]interface{}) (bool, error) {
	return false, fmt.Errorf("filter error")
}

// --- Tests ---

func testRecords() []reader.Record {
	return []reader.Record{
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25},
		{"name": "Charlie", "age": 35},
	}
}

func TestProcess(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	proc := NewProcessor(r, w)
	if err := proc.Process(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 3 {
		t.Errorf("expected 3 records, got %d", len(w.records))
	}
	if !w.flushed {
		t.Error("expected Flush to be called")
	}
}

func TestProcessEmptyReader(t *testing.T) {
	r := &mockReader{records: nil}
	w := &mockWriter{}

	proc := NewProcessor(r, w)
	if err := proc.Process(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 0 {
		t.Errorf("expected 0 records, got %d", len(w.records))
	}
	if !w.flushed {
		t.Error("expected Flush to be called even for empty input")
	}
}

func TestProcessWrappedEOF(t *testing.T) {
	r := &wrappedEOFReader{records: testRecords()}
	w := &mockWriter{}

	proc := NewProcessor(r, w)
	if err := proc.Process(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 3 {
		t.Errorf("expected 3 records, got %d", len(w.records))
	}
}

func TestProcessReadError(t *testing.T) {
	r := &errorReader{
		records:  testRecords(),
		errorAt:  2,
		errorMsg: "disk failure",
	}
	w := &mockWriter{}

	proc := NewProcessor(r, w)
	err := proc.Process()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "read error") {
		t.Errorf("expected 'read error' in message, got: %v", err)
	}
	if !strings.Contains(err.Error(), "disk failure") {
		t.Errorf("expected 'disk failure' in message, got: %v", err)
	}
	if len(w.records) != 2 {
		t.Errorf("expected 2 records before error, got %d", len(w.records))
	}
}

func TestProcessWriteError(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &errorWriter{errorAt: 1}

	proc := NewProcessor(r, w)
	err := proc.Process()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "write error") {
		t.Errorf("expected 'write error' in message, got: %v", err)
	}
	if !strings.Contains(err.Error(), "write failed") {
		t.Errorf("expected 'write failed' in message, got: %v", err)
	}
}

func TestProcessWithFilter(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	f := &mockFilter{matchFn: func(record map[string]interface{}) bool {
		age, _ := record["age"].(int)
		return age > 28
	}}

	proc := NewProcessor(r, w, WithFilter(f))
	if err := proc.Process(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 2 {
		t.Errorf("expected 2 filtered records, got %d", len(w.records))
	}
}

func TestProcessWithFilterNoMatches(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	f := &mockFilter{matchFn: func(record map[string]interface{}) bool {
		return false
	}}

	proc := NewProcessor(r, w, WithFilter(f))
	if err := proc.Process(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 0 {
		t.Errorf("expected 0 records, got %d", len(w.records))
	}
	if !w.flushed {
		t.Error("expected Flush even with no matches")
	}
}

func TestProcessWithFilterError(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	proc := NewProcessor(r, w, WithFilter(&errorFilter{}))
	err := proc.Process()
	if err == nil {
		t.Fatal("expected filter error")
	}
	if !strings.Contains(err.Error(), "filter error") {
		t.Errorf("expected 'filter error' in message, got: %v", err)
	}
}

func TestProcessWithTransform(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	transform := func(record reader.Record) (interface{}, error) {
		return record["name"], nil
	}

	proc := NewProcessor(r, w, WithTransform(transform))
	if err := proc.Process(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 3 {
		t.Errorf("expected 3 records, got %d", len(w.records))
	}
	if w.records[0] != "Alice" {
		t.Errorf("expected 'Alice', got %v", w.records[0])
	}
	if w.records[1] != "Bob" {
		t.Errorf("expected 'Bob', got %v", w.records[1])
	}
}

func TestProcessWithTransformError(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	transform := func(record reader.Record) (interface{}, error) {
		return nil, fmt.Errorf("transform failed")
	}

	proc := NewProcessor(r, w, WithTransform(transform))
	err := proc.Process()
	if err == nil {
		t.Fatal("expected transform error")
	}
	if !strings.Contains(err.Error(), "transform error") {
		t.Errorf("expected 'transform error' in message, got: %v", err)
	}
}

func TestProcessWithFilterAndTransform(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	f := &mockFilter{matchFn: func(record map[string]interface{}) bool {
		age, _ := record["age"].(int)
		return age > 28
	}}

	transform := func(record reader.Record) (interface{}, error) {
		return map[string]interface{}{"n": record["name"]}, nil
	}

	proc := NewProcessor(r, w, WithFilter(f), WithTransform(transform))
	if err := proc.Process(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 2 {
		t.Errorf("expected 2 records, got %d", len(w.records))
	}

	first, ok := w.records[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", w.records[0])
	}
	if first["n"] != "Alice" {
		t.Errorf("expected 'Alice', got %v", first["n"])
	}
}

func TestProcessWithLimit(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	proc := NewProcessor(r, w)
	if err := proc.ProcessWithLimit(2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 2 {
		t.Errorf("expected 2 records, got %d", len(w.records))
	}
}

func TestProcessWithLimitZero(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	proc := NewProcessor(r, w)
	if err := proc.ProcessWithLimit(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// limit <= 0 reads all records
	if len(w.records) != 3 {
		t.Errorf("expected 3 records with limit=0, got %d", len(w.records))
	}
}

func TestProcessWithLimitExceedsRecords(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	proc := NewProcessor(r, w)
	if err := proc.ProcessWithLimit(100); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 3 {
		t.Errorf("expected 3 records (all available), got %d", len(w.records))
	}
}

func TestProcessWithLimitAndFilter(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	f := &mockFilter{matchFn: func(record map[string]interface{}) bool {
		return true // all match
	}}

	proc := NewProcessor(r, w, WithFilter(f))
	if err := proc.ProcessWithLimit(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 1 {
		t.Errorf("expected 1 record, got %d", len(w.records))
	}
}

func TestProcessWithLimitFilterAndTransform(t *testing.T) {
	r := &mockReader{records: testRecords()}
	w := &mockWriter{}

	f := &mockFilter{matchFn: func(record map[string]interface{}) bool {
		age, _ := record["age"].(int)
		return age > 20
	}}

	transform := func(record reader.Record) (interface{}, error) {
		return record["name"], nil
	}

	proc := NewProcessor(r, w, WithFilter(f), WithTransform(transform))
	if err := proc.ProcessWithLimit(2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.records) != 2 {
		t.Errorf("expected 2 records, got %d", len(w.records))
	}
	if w.records[0] != "Alice" {
		t.Errorf("expected 'Alice', got %v", w.records[0])
	}
	if w.records[1] != "Bob" {
		t.Errorf("expected 'Bob', got %v", w.records[1])
	}
}

func TestProcessWithLimitWrappedEOF(t *testing.T) {
	r := &wrappedEOFReader{records: testRecords()}
	w := &mockWriter{}

	proc := NewProcessor(r, w)
	if err := proc.ProcessWithLimit(100); err != nil {
		t.Fatalf("unexpected error with wrapped EOF: %v", err)
	}

	if len(w.records) != 3 {
		t.Errorf("expected 3 records, got %d", len(w.records))
	}
}
