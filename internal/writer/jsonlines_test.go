package writer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteMap(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	record := map[string]interface{}{"name": "Alice", "age": 30}
	if err := w.Write(record); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}
	if parsed["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %v", parsed["name"])
	}
}

func TestWriteString(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	if err := w.Write("hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != `"hello"` {
		t.Errorf("expected '\"hello\"', got %s", output)
	}
}

func TestWriteNumber(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	if err := w.Write(42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "42" {
		t.Errorf("expected '42', got %s", output)
	}
}

func TestWriteNil(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	if err := w.Write(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "null" {
		t.Errorf("expected 'null', got %s", output)
	}
}

func TestWriteSlice(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	if err := w.Write([]string{"a", "b", "c"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != `["a","b","c"]` {
		t.Errorf("expected '[\"a\",\"b\",\"c\"]', got %s", output)
	}
}

func TestWriteMultipleRecords(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	records := []interface{}{
		map[string]interface{}{"name": "Alice"},
		map[string]interface{}{"name": "Bob"},
		map[string]interface{}{"name": "Charlie"},
	}
	for _, rec := range records {
		if err := w.Write(rec); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(line), &parsed); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestWriteCompactNoBlankLines(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	for i := 0; i < 3; i++ {
		if err := w.Write(map[string]interface{}{"i": i}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	if strings.Contains(buf.String(), "\n\n") {
		t.Error("compact mode should not have blank lines between records")
	}
}

func TestWritePrettyIndentation(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, true)

	record := map[string]interface{}{"name": "Alice", "age": 30}
	if err := w.Write(record); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "  ") {
		t.Error("pretty mode should contain indentation")
	}
	if !strings.Contains(output, "\n") {
		t.Error("pretty mode should contain newlines inside the object")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestWritePrettyBlankLineSeparator(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, true)

	for i := 0; i < 3; i++ {
		if err := w.Write(map[string]interface{}{"i": i}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "}\n\n{") {
		t.Error("pretty mode should separate records with blank lines")
	}

	records := strings.Split(strings.TrimSpace(output), "\n\n")
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
	for i, rec := range records {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(rec), &parsed); err != nil {
			t.Errorf("record %d is not valid JSON: %v\nRecord: %s", i, err, rec)
		}
	}
}

func TestWriteNoOutputBeforeFlush(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	if err := w.Write(map[string]interface{}{"x": 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected output after Flush")
	}
}

func TestWriteEmptyInput(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output, got: %s", buf.String())
	}
}

func TestWriteNestedStructure(t *testing.T) {
	var buf bytes.Buffer
	w := NewJSONLinesWriter(&buf, false)

	record := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"address": map[string]interface{}{
				"city": "Wonderland",
			},
		},
		"tags": []string{"admin", "user"},
	}

	if err := w.Write(record); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	user, ok := parsed["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected user to be a map")
	}
	if user["name"] != "Alice" {
		t.Errorf("expected user.name=Alice, got %v", user["name"])
	}
}

func TestWriterImplementsInterface(t *testing.T) {
	var _ Writer = (*JSONLinesWriter)(nil)
}
