package e2e

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func TestSchemaSimpleFile(t *testing.T) {
	stdout, stderr, exitCode := runGavro("schema", "../../tests/testdata/users.avro")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	// Проверяем что это валидный JSON
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &schema); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Проверяем основные поля схемы
	if schema["name"] != "User" {
		t.Errorf("Expected schema name 'User', got %v", schema["name"])
	}

	if schema["type"] != "record" {
		t.Errorf("Expected schema type 'record', got %v", schema["type"])
	}

	// Проверяем поля
	fields, ok := schema["fields"].([]interface{})
	if !ok || len(fields) != 3 {
		t.Errorf("Expected 3 fields, got %v", fields)
	}
}

func TestSchemaComplexFile(t *testing.T) {
	stdout, stderr, exitCode := runGavro("schema", "../../tests/testdata/complex.avro")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &schema); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Проверяем сложную схему
	if schema["name"] != "ComplexRecord" {
		t.Errorf("Expected schema name 'ComplexRecord', got %v", schema["name"])
	}

	fields := schema["fields"].([]interface{})
	if len(fields) != 6 {
		t.Errorf("Expected 6 fields in complex schema, got %d", len(fields))
	}
}

func TestSchemaPrettyFlag(t *testing.T) {
	stdout, stderr, exitCode := runGavro("schema", "../../tests/testdata/users.avro", "--pretty")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	// Pretty формат должен содержать переносы строк и отступы
	if !strings.Contains(stdout, "\n") {
		t.Error("Pretty output should contain newlines")
	}

	if !strings.Contains(stdout, "  ") {
		t.Error("Pretty output should contain indentation")
	}

	// Должен быть валидный JSON
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &schema); err != nil {
		t.Fatalf("Pretty output is not valid JSON: %v", err)
	}
}

func TestSchemaCompactOutput(t *testing.T) {
	stdout, stderr, exitCode := runGavro("schema", "../../tests/testdata/users.avro")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	// Compact формат должен быть одной строкой (плюс \n в конце)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Errorf("Compact output should be single line, got %d lines", len(lines))
	}
}

func TestSchemaFileNotFound(t *testing.T) {
	_, stderr, exitCode := runGavro("schema", "../../tests/testdata/nonexistent.avro")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code for missing file")
	}

	if !strings.Contains(stderr, "no such file or directory") {
		t.Errorf("Expected 'no such file or directory' error, got: %s", stderr)
	}
}

func TestSchemaInvalidAvroFile(t *testing.T) {
	testCases := []struct {
		name string
		file string
	}{
		{"bad magic", "../../tests/testdata/bad_magic.avro"},
		{"totally empty", "../../tests/testdata/totally_empty.avro"},
		{"garbage", "../../tests/testdata/garbage.avro"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, exitCode := runGavro("schema", tc.file)

			if exitCode == 0 {
				t.Errorf("Expected non-zero exit code for invalid file %s", tc.file)
			}

			if stderr == "" {
				t.Error("Expected error message in stderr")
			}
		})
	}
}

func TestSchemaNoArgs(t *testing.T) {
	_, stderr, exitCode := runGavro("schema")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code when no args provided")
	}

	if !strings.Contains(stderr, "accepts 1 arg(s)") {
		t.Errorf("Expected argument error, got: %s", stderr)
	}
}

func TestSchemaHelp(t *testing.T) {
	stdout, _, exitCode := runGavro("schema", "--help")

	if exitCode != 0 {
		t.Fatal("schema help should return exit code 0")
	}

	if !strings.Contains(stdout, "schema") {
		t.Error("Schema help should mention 'schema'")
	}

	if !strings.Contains(stdout, "--pretty") {
		t.Error("Schema help should mention --pretty flag")
	}
}

func TestSchemaWithJq(t *testing.T) {
	// Проверяем что jq есть
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not found, skipping integration test")
	}

	// gavro schema | jq '.fields[].name'
	gavroCmd := exec.Command("/tmp/"+binaryName, "schema", "../../tests/testdata/users.avro")
	jqCmd := exec.Command("jq", "-r", ".fields[].name")

	var output bytes.Buffer
	jqCmd.Stdin, _ = gavroCmd.StdoutPipe()
	jqCmd.Stdout = &output

	jqCmd.Start()
	gavroCmd.Run()
	jqCmd.Wait()

	result := output.String()
	expectedFields := []string{"name", "age", "email"}

	for _, field := range expectedFields {
		if !strings.Contains(result, field) {
			t.Errorf("Expected field '%s' in schema output", field)
		}
	}
}

func TestSchemaDifferentFiles(t *testing.T) {
	testCases := []struct {
		name           string
		file           string
		expectedName   string
		expectedFields int
	}{
		{"simple users", "../../tests/testdata/users.avro", "User", 3},
		{"complex schema", "../../tests/testdata/complex.avro", "ComplexRecord", 6},
		{"empty file", "../../tests/testdata/empty.avro", "Empty", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, exitCode := runGavro("schema", tc.file)

			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}

			var schema map[string]interface{}
			if err := json.Unmarshal([]byte(stdout), &schema); err != nil {
				t.Fatalf("Output is not valid JSON: %v", err)
			}

			if schema["name"] != tc.expectedName {
				t.Errorf("Expected schema name '%s', got %v", tc.expectedName, schema["name"])
			}

			fields := schema["fields"].([]interface{})
			if len(fields) != tc.expectedFields {
				t.Errorf("Expected %d fields, got %d", tc.expectedFields, len(fields))
			}
		})
	}
}
