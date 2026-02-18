package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSelectSingleField(t *testing.T) {
	stdout, stderr, exitCode := runGavro("select", "../../tests/testdata/users.avro", "record.name")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	expected := []string{`"Alice"`, `"Bob"`, `"Charlie"`}
	for i, line := range lines {
		if strings.TrimSpace(line) != expected[i] {
			t.Errorf("Line %d: expected %s, got %s", i, expected[i], line)
		}
	}
}

func TestSelectSingleFieldNumeric(t *testing.T) {
	stdout, stderr, exitCode := runGavro("select", "../../tests/testdata/users.avro", "record.age")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	expected := []string{"30", "25", "35"}
	for i, line := range lines {
		if strings.TrimSpace(line) != expected[i] {
			t.Errorf("Line %d: expected %s, got %s", i, expected[i], line)
		}
	}
}

func TestSelectMultipleFields(t *testing.T) {
	stdout, stderr, exitCode := runGavro("select", "../../tests/testdata/users.avro", "record.name", "record.age")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("Line %d is not valid JSON: %v", i, err)
		}

		if _, ok := record["name"]; !ok {
			t.Errorf("Line %d missing 'name' key", i)
		}
		if _, ok := record["age"]; !ok {
			t.Errorf("Line %d missing 'age' key", i)
		}

		if _, ok := record["record.name"]; ok {
			t.Errorf("Line %d has 'record.name' key — prefix should be stripped", i)
		}
	}

	var first map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("Failed to parse first line as JSON: %v", err)
	}
	if first["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", first["name"])
	}
	if first["age"] != float64(30) {
		t.Errorf("Expected age=30, got %v", first["age"])
	}
}

func TestSelectNestedField(t *testing.T) {
	stdout, stderr, exitCode := runGavro("select", "../../tests/testdata/complex.avro", "record.nested.field1")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	if strings.TrimSpace(lines[0]) != `"nested_value"` {
		t.Errorf("Expected '\"nested_value\"', got %s", lines[0])
	}
	if strings.TrimSpace(lines[1]) != `"another"` {
		t.Errorf("Expected '\"another\"', got %s", lines[1])
	}
}

func TestSelectMultipleNestedFields(t *testing.T) {
	stdout, stderr, exitCode := runGavro("select", "../../tests/testdata/complex.avro", "record.nested.field1", "record.nested.field2")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	var first map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if _, ok := first["nested.field1"]; !ok {
		t.Error("Expected 'nested.field1' key")
	}
	if _, ok := first["nested.field2"]; !ok {
		t.Error("Expected 'nested.field2' key")
	}
}

func TestSelectPrettyFlag(t *testing.T) {
	stdout, _, exitCode := runGavro("select", "../../tests/testdata/users.avro", "record.name", "record.age", "--pretty")

	if exitCode != 0 {
		t.Fatal("select --pretty failed")
	}

	if !strings.Contains(stdout, "  ") {
		t.Error("Output should contain indentation in pretty mode")
	}

	if !strings.Contains(stdout, "}\n\n{") {
		t.Error("Records should be separated by blank lines in pretty mode")
	}
}

func TestSelectPrettyShortFlag(t *testing.T) {
	stdout, _, exitCode := runGavro("select", "../../tests/testdata/users.avro", "record.name", "record.age", "-p")

	if exitCode != 0 {
		t.Fatal("select -p failed")
	}

	if !strings.Contains(stdout, "  ") {
		t.Error("Output should contain indentation with -p flag")
	}
}

func TestSelectEmptyFile(t *testing.T) {
	stdout, _, exitCode := runGavro("select", "../../tests/testdata/empty.avro", "record.name")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0 for empty file, got %d", exitCode)
	}

	if strings.TrimSpace(stdout) != "" {
		t.Errorf("Expected empty output for empty file, got: %s", stdout)
	}
}

func TestSelectNonExistentField(t *testing.T) {
	_, stderr, exitCode := runGavro("select", "../../tests/testdata/users.avro", "record.nonexistent")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code for non-existent field")
	}

	if !strings.Contains(stderr, "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %s", stderr)
	}
}

func TestSelectWithoutRecordPrefix(t *testing.T) {
	_, stderr, exitCode := runGavro("select", "../../tests/testdata/users.avro", "name")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code for path without record. prefix")
	}

	if !strings.Contains(stderr, "record.") {
		t.Errorf("Expected error about 'record.' prefix, got: %s", stderr)
	}
}

func TestSelectJustRecordPrefix(t *testing.T) {
	_, stderr, exitCode := runGavro("select", "../../tests/testdata/users.avro", "record.")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code for bare 'record.'")
	}

	if !strings.Contains(stderr, "cannot be just") && !strings.Contains(stderr, "does not exist") {
		t.Errorf("Expected meaningful error, got: %s", stderr)
	}
}

func TestSelectFileNotFound(t *testing.T) {
	_, stderr, exitCode := runGavro("select", "nonexistent.avro", "record.name")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code for missing file")
	}

	if !strings.Contains(stderr, "no such file or directory") {
		t.Errorf("Expected file not found error, got: %s", stderr)
	}
}

func TestSelectNoArgs(t *testing.T) {
	_, stderr, exitCode := runGavro("select")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code when no args")
	}

	if !strings.Contains(stderr, "requires at least 2 arg(s)") {
		t.Errorf("Expected argument error, got: %s", stderr)
	}
}

func TestSelectOnlyFile(t *testing.T) {
	_, stderr, exitCode := runGavro("select", "../../tests/testdata/users.avro")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code with only file arg")
	}

	if !strings.Contains(stderr, "requires at least 2 arg(s)") {
		t.Errorf("Expected argument error, got: %s", stderr)
	}
}

func TestSelectHelp(t *testing.T) {
	stdout, _, exitCode := runGavro("select", "--help")

	if exitCode != 0 {
		t.Fatal("select help should return exit code 0")
	}

	requiredStrings := []string{
		"record.",
		"field",
		"pretty",
	}
	for _, str := range requiredStrings {
		if !strings.Contains(stdout, str) {
			t.Errorf("Help should contain '%s'", str)
		}
	}
}

func TestSelectLargeFile(t *testing.T) {
	stdout, stderr, exitCode := runGavro("select", "../../tests/testdata/large.avro", "record.level")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 10000 {
		t.Errorf("Expected 10000 lines, got %d", len(lines))
	}

	for i, line := range lines[:10] {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, `"`) || !strings.HasSuffix(trimmed, `"`) {
			t.Errorf("Line %d: expected quoted string, got %s", i, line)
		}
	}
}

func TestSelectOutputIsJSONLines(t *testing.T) {
	stdout, _, exitCode := runGavro("select", "../../tests/testdata/users.avro", "record.name", "record.email")

	if exitCode != 0 {
		t.Fatal("select failed")
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for i, line := range lines {
		var obj interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}

		if strings.HasSuffix(line, ",") {
			t.Errorf("Line %d should not end with comma", i)
		}
	}

	if strings.HasPrefix(stdout, "[") {
		t.Error("Output should not be a JSON array")
	}

	if strings.Contains(stdout, "\n\n") {
		t.Error("Should not have blank lines in compact mode")
	}
}
