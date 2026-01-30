package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const binaryName = "gavro"

// TestMain билдит бинарник перед запуском тестов
func TestMain(m *testing.M) {
	// Билдим gavro в /tmp
	build := exec.Command("go", "build", "-o", "/tmp/"+binaryName, "../../main.go")
	if err := build.Run(); err != nil {
		panic("Failed to build binary: " + err.Error())
	}

	// Проверяем что тестовые данные существуют
	// Если нет - нужно запустить: go run tests/testdata/generate.go
	testFiles := []string{
		"../../tests/testdata/users.avro",
		"../../tests/testdata/complex.avro",
		"../../tests/testdata/empty.avro",
	}

	allExist := true
	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			allExist = false
			break
		}
	}

	if !allExist {
		panic("Test data files not found. Run: go run tests/testdata/generate.go")
	}

	code := m.Run()
	os.Exit(code)
}

func runGavro(args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command("/tmp/"+binaryName, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return outBuf.String(), errBuf.String(), exitCode
}

func TestCatSimpleUsers(t *testing.T) {
	stdout, stderr, exitCode := runGavro("cat", "../../tests/testdata/users.avro")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	// Проверяем что каждая строка - валидный JSON
	for i, line := range lines {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}

		// Проверяем наличие полей
		if _, ok := record["name"]; !ok {
			t.Errorf("Line %d missing 'name' field", i)
		}
		if _, ok := record["age"]; !ok {
			t.Errorf("Line %d missing 'age' field", i)
		}
		if _, ok := record["email"]; !ok {
			t.Errorf("Line %d missing 'email' field", i)
		}
	}

	// Проверяем конкретные значения
	var firstUser map[string]interface{}
	json.Unmarshal([]byte(lines[0]), &firstUser)
	if firstUser["name"] != "Alice" {
		t.Errorf("Expected first user to be Alice, got %v", firstUser["name"])
	}
}

func TestCatComplexSchema(t *testing.T) {
	stdout, stderr, exitCode := runGavro("cat", "../../tests/testdata/complex.avro")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(lines))
	}

	// Проверяем первую запись с вложенными структурами
	var record map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Проверяем массив
	if tags, ok := record["tags"].([]interface{}); !ok || len(tags) != 3 {
		t.Errorf("Expected tags array with 3 elements, got %v", record["tags"])
	}

	// Проверяем map
	if metadata, ok := record["metadata"].(map[string]interface{}); !ok || len(metadata) != 2 {
		t.Errorf("Expected metadata map with 2 elements, got %v", record["metadata"])
	}

	// Проверяем вложенный record
	if nested, ok := record["nested"].(map[string]interface{}); !ok {
		t.Errorf("Expected nested record, got %v", record["nested"])
	} else {
		if nested["field1"] != "nested_value" {
			t.Errorf("Expected nested.field1 to be 'nested_value', got %v", nested["field1"])
		}
	}
}

func TestCatEmptyFile(t *testing.T) {
	stdout, _, exitCode := runGavro("cat", "../../tests/testdata/empty.avro")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", exitCode)
	}

	if strings.TrimSpace(stdout) != "" {
		t.Errorf("Expected no output for empty file, got: %s", stdout)
	}
}

func TestCatLargeFile(t *testing.T) {
	stdout, stderr, exitCode := runGavro("cat", "../../tests/testdata/large.avro")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 10000 {
		t.Errorf("Expected 10000 lines, got %d", len(lines))
	}

	// Проверяем случайные строки
	for _, idx := range []int{0, 500, 5000, 9999} {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(lines[idx]), &record); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", idx, err)
		}
	}
}

func TestCatFileNotFound(t *testing.T) {
	_, stderr, exitCode := runGavro("cat", "../../tests/testdata/nonexistent.avro")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code for missing file")
	}

	if !strings.Contains(stderr, "no such file or directory") {
		t.Errorf("Expected 'no such file or directory' error, got: %s", stderr)
	}
}

func TestCatInvalidAvroFile(t *testing.T) {
	testCases := []struct {
		name     string
		file     string
		errorMsg string
	}{
		{"bad magic", "../../tests/testdata/bad_magic.avro", "avro"},
		{"totally empty", "../../tests/testdata/totally_empty.avro", ""},
		{"truncated", "../../tests/testdata/truncated.avro", ""},
		{"garbage", "../../tests/testdata/garbage.avro", "avro"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, exitCode := runGavro("cat", tc.file)

			if exitCode == 0 {
				t.Errorf("Expected non-zero exit code for invalid file %s", tc.file)
			}

			// Программа должна выдать ошибку, а не упасть
			if stderr == "" {
				t.Error("Expected error message in stderr")
			}
		})
	}
}

func TestCatNoArgs(t *testing.T) {
	_, stderr, exitCode := runGavro("cat")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code when no args provided")
	}

	if !strings.Contains(stderr, "accepts 1 arg(s)") {
		t.Errorf("Expected argument error, got: %s", stderr)
	}
}

func TestCatWithJq(t *testing.T) {
	// Проверяем что jq есть
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq not found, skipping integration test")
	}

	// gavro cat | jq 'select(.age > 28)'
	gavroCmd := exec.Command("/tmp/"+binaryName, "cat", "../../tests/testdata/users.avro")
	jqCmd := exec.Command("jq", "select(.age > 28)")

	var output bytes.Buffer
	jqCmd.Stdin, _ = gavroCmd.StdoutPipe()
	jqCmd.Stdout = &output

	jqCmd.Start()
	gavroCmd.Run()
	jqCmd.Wait()

	result := output.String()

	// jq выводит в pretty format с отступами, поэтому считаем только объекты (строки с "name")
	nameLines := strings.Count(result, `"name"`)

	// Должны остаться только Alice (30) и Charlie (35)
	if nameLines != 2 {
		t.Errorf("Expected 2 filtered results, got %d. Output:\n%s", nameLines, result)
	}

	if !strings.Contains(result, "Alice") {
		t.Error("Expected Alice in filtered results")
	}
	if !strings.Contains(result, "Charlie") {
		t.Error("Expected Charlie in filtered results")
	}
	if strings.Contains(result, "Bob") {
		t.Error("Did not expect Bob in filtered results")
	}
}

func TestCatOutputFormat(t *testing.T) {
	stdout, _, exitCode := runGavro("cat", "../../tests/testdata/users.avro")

	if exitCode != 0 {
		t.Fatal("gavro cat failed")
	}

	// Проверяем формат JSON Lines: каждая строка - отдельный JSON объект
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for i, line := range lines {
		// Не должно быть запятых в конце
		if strings.HasSuffix(line, ",") {
			t.Errorf("Line %d should not end with comma", i)
		}

		// Не должно быть массива
		if strings.HasPrefix(stdout, "[") {
			t.Error("Output should not be a JSON array")
		}

		// Каждая строка должна быть валидным JSON объектом
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("Line %d is not valid JSON object: %v", i, err)
		}
	}
}

func TestHelpCommand(t *testing.T) {
	stdout, _, exitCode := runGavro("--help")

	if exitCode != 0 {
		t.Fatal("help should return exit code 0")
	}

	if !strings.Contains(stdout, "gavro") {
		t.Error("Help should mention 'gavro'")
	}
	if !strings.Contains(stdout, "cat") {
		t.Error("Help should mention 'cat' command")
	}
}

func TestCatHelp(t *testing.T) {
	stdout, _, exitCode := runGavro("cat", "--help")

	if exitCode != 0 {
		t.Fatal("cat help should return exit code 0")
	}

	if !strings.Contains(stdout, "JSON Lines") {
		t.Error("Cat help should mention JSON Lines format")
	}
	if !strings.Contains(stdout, "jq") {
		t.Error("Cat help should mention jq")
	}
}

// Benchmark для проверки производительности
func BenchmarkCatLargeFile(b *testing.B) {
	// Генерируем файл если его нет
	if _, err := os.Stat("../../tests/testdata/large.avro"); os.IsNotExist(err) {
		generate := exec.Command("go", "run", "../testdata/generate.go")
		generate.Dir = "../.."
		generate.Run()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("/tmp/"+binaryName, "cat", "../../tests/testdata/large.avro")
		cmd.Stdout = nil // Не выводим в stdout
		cmd.Run()
	}
}

// Тест на утечки памяти при обработке больших файлов
func TestCatMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	// Запускаем несколько раз подряд - не должно быть утечек
	for i := 0; i < 5; i++ {
		_, _, exitCode := runGavro("cat", "../../tests/testdata/large.avro")
		if exitCode != 0 {
			t.Fatalf("Run %d failed with exit code %d", i, exitCode)
		}
	}
}

// Тест проверяет что gavro правильно закрывает файлы
func TestCatFileHandles(t *testing.T) {
	// Открываем файл много раз - не должно быть "too many open files"
	for i := 0; i < 100; i++ {
		_, _, exitCode := runGavro("cat", "../../tests/testdata/users.avro")
		if exitCode != 0 {
			t.Fatalf("Run %d failed", i)
		}
	}
}

// Тест на работу с файлами в разных директориях
func TestCatWithDifferentPaths(t *testing.T) {
	testCases := []string{
		"../../tests/testdata/users.avro",             // relative from test dir
		"../../tests/testdata/../testdata/users.avro", // with ..
	}

	for _, path := range testCases {
		t.Run(path, func(t *testing.T) {
			_, _, exitCode := runGavro("cat", path)
			if exitCode != 0 {
				t.Errorf("Failed to read file with path: %s", path)
			}
		})
	}
}

// Тест --pretty флага
func TestCatPrettyFlag(t *testing.T) {
	stdout, _, exitCode := runGavro("cat", "../../tests/testdata/users.avro", "--pretty")

	if exitCode != 0 {
		t.Fatal("cat --pretty failed")
	}

	// Проверяем что есть отступы (признак pretty-print)
	if !strings.Contains(stdout, "  \"age\"") {
		t.Error("Output should contain indented fields")
	}

	// Проверяем что между записями есть пустые строки
	if !strings.Contains(stdout, "}\n\n{") {
		t.Error("Records should be separated by blank lines in pretty mode")
	}

	// Проверяем что весь вывод - валидный JSON если объединить записи в массив
	// Разделяем по пустым строкам
	records := strings.Split(strings.TrimSpace(stdout), "\n\n")
	if len(records) != 3 {
		t.Errorf("Expected 3 records separated by blank lines, got %d", len(records))
	}

	// Каждая запись должна быть валидным JSON
	for i, record := range records {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(record), &obj); err != nil {
			t.Errorf("Record %d is not valid JSON: %v\nRecord:\n%s", i, err, record)
		}
	}
}

// Тест короткого флага -p
func TestCatPrettyShortFlag(t *testing.T) {
	stdout, _, exitCode := runGavro("cat", "../../tests/testdata/users.avro", "-p")

	if exitCode != 0 {
		t.Fatal("cat -p failed")
	}

	// Проверяем что есть отступы
	if !strings.Contains(stdout, "  \"age\"") {
		t.Error("Output should contain indented fields with -p flag")
	}
}

// Тест что без флага выводится компактный JSON
func TestCatCompactByDefault(t *testing.T) {
	stdout, _, exitCode := runGavro("cat", "../../tests/testdata/users.avro")

	if exitCode != 0 {
		t.Fatal("cat failed")
	}

	// Проверяем что НЕТ отступов (компактный JSON)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	// Каждая строка должна начинаться с {, а не с пробелов
	for i, line := range lines {
		if !strings.HasPrefix(line, "{") {
			t.Errorf("Line %d should start with '{' in compact mode, got: %s", i, line[:min(20, len(line))])
		}

		// Не должно быть двойных переносов строк между записями
		if strings.Contains(line, "\n\n") {
			t.Error("Should not have blank lines between records in compact mode")
		}
	}
}
