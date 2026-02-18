package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestQuerySimpleFilter(t *testing.T) {
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 28")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 matching records (Alice and Charlie), got %d", len(lines))
	}

	var record1, record2 map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &record1); err != nil {
		t.Fatalf("Failed to parse line 0 as JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &record2); err != nil {
		t.Fatalf("Failed to parse line 1 as JSON: %v", err)
	}

	if record1["name"] != "Alice" || record2["name"] != "Charlie" {
		t.Errorf("Expected Alice and Charlie, got %v and %v", record1["name"], record2["name"])
	}
}

func TestQueryComplexFilter(t *testing.T) {
	testCases := []struct {
		name          string
		expression    string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "age >= 30 && name startsWith A",
			expression:    "record.age >= 30 && record.name.startsWith('A')",
			expectedCount: 1,
			expectedNames: []string{"Alice"},
		},
		{
			name:          "email contains example",
			expression:    "record.email.contains('example')",
			expectedCount: 3,
			expectedNames: []string{"Alice", "Bob", "Charlie"},
		},
		{
			name:          "age != 25",
			expression:    "record.age != 25",
			expectedCount: 2,
			expectedNames: []string{"Alice", "Charlie"},
		},
		{
			name:          "email endsWith .com",
			expression:    "record.email.endsWith('.com')",
			expectedCount: 3,
			expectedNames: []string{"Alice", "Bob", "Charlie"},
		},
		{
			name:          "name startsWith B",
			expression:    "record.name.startsWith('B')",
			expectedCount: 1,
			expectedNames: []string{"Bob"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", tc.expression)

			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}

			lines := strings.Split(strings.TrimSpace(stdout), "\n")
			if len(lines) != tc.expectedCount {
				t.Errorf("Expected %d records, got %d", tc.expectedCount, len(lines))
			}

			for i, line := range lines {
				if line == "" {
					continue
				}
				var record map[string]interface{}
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if i < len(tc.expectedNames) {
					if record["name"] != tc.expectedNames[i] {
						t.Errorf("Expected name %s at position %d, got %v", tc.expectedNames[i], i, record["name"])
					}
				}
			}
		})
	}
}

func TestQueryAliasQ(t *testing.T) {
	stdout, _, exitCode := runGavro("q", "../../tests/testdata/users.avro", "record.age == 25")

	if exitCode != 0 {
		t.Fatal("q alias should work")
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 record (Bob), got %d", len(lines))
	}

	var record map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatalf("Failed to parse line 0 as JSON: %v", err)
	}
	if record["name"] != "Bob" {
		t.Errorf("Expected Bob, got %v", record["name"])
	}
}

func TestQueryNoMatches(t *testing.T) {
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 100")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if strings.TrimSpace(stdout) != "" {
		t.Errorf("Expected empty output for no matches, got: %s", stdout)
	}
}

func TestQueryAllMatch(t *testing.T) {
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 0")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected all 3 records, got %d", len(lines))
	}
}

func TestQueryInvalidExpression(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
	}{
		{"syntax error", "record.age >"},
		{"invalid field", "record..age > 18"},
		{"unclosed paren", "record.age > 18 && (record.name == 'Alice'"},
		{"invalid operator", "record.age === 18"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", tc.expression)

			if exitCode == 0 {
				t.Error("Expected non-zero exit code for invalid expression")
			}

			if !strings.Contains(stderr, "invalid expression") && !strings.Contains(stderr, "failed to compile") {
				t.Logf("Stderr: %s", stderr)
			}
		})
	}
}

func TestQueryMissingArgs(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{"no args", []string{"query"}},
		{"only file", []string{"query", "users.avro"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, exitCode := runGavro(tc.args...)

			if exitCode == 0 {
				t.Error("Expected non-zero exit code")
			}

			if !strings.Contains(stderr, "accepts 2 arg(s)") {
				t.Errorf("Expected argument error, got: %s", stderr)
			}
		})
	}
}

func TestQueryNonExistentFile(t *testing.T) {
	_, stderr, exitCode := runGavro("query", "nonexistent.avro", "record.age > 18")

	if exitCode == 0 {
		t.Fatal("Expected non-zero exit code for missing file")
	}

	if !strings.Contains(stderr, "no such file or directory") {
		t.Errorf("Expected file not found error, got: %s", stderr)
	}
}

func TestQueryOnComplexSchema(t *testing.T) {
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/complex.avro", "record.id > 0")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 records, got %d", len(lines))
	}
}

func TestQueryJSONLinesFormat(t *testing.T) {
	stdout, _, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 28")

	if exitCode != 0 {
		t.Fatal("query failed")
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for i, line := range lines {
		var record map[string]interface{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}

		if strings.HasSuffix(line, ",") {
			t.Errorf("Line %d should not end with comma", i)
		}
	}

	if strings.HasPrefix(stdout, "[") {
		t.Error("Output should not be a JSON array")
	}
}

func TestQueryHelp(t *testing.T) {
	stdout, _, exitCode := runGavro("query", "--help")

	if exitCode != 0 {
		t.Fatal("query help should return exit code 0")
	}

	requiredStrings := []string{
		"CEL",
		"expression",
		"record.age",
		"startsWith",
		"endsWith",
	}

	for _, str := range requiredStrings {
		if !strings.Contains(stdout, str) {
			t.Errorf("Help should contain '%s'", str)
		}
	}
}

func TestQueryBooleanLogic(t *testing.T) {
	testCases := []struct {
		name          string
		expression    string
		expectedCount int
	}{
		{"AND both true", "record.age > 25 && record.age < 35", 1},                     // Alice (30)
		{"OR", "record.age < 26 || record.age > 34", 2},                                // Bob (25), Charlie (35)
		{"NOT", "!(record.age == 25)", 2},                                              // Alice, Charlie
		{"Complex", "(record.age > 25 && record.age < 35) || record.name == 'Bob'", 2}, // Alice (30), Bob (25)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", tc.expression)

			if exitCode != 0 {
				t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
			}

			lines := strings.Split(strings.TrimSpace(stdout), "\n")
			actualCount := len(lines)
			if lines[0] == "" {
				actualCount = 0
			}

			if actualCount != tc.expectedCount {
				t.Errorf("Expected %d records, got %d. Output: %s", tc.expectedCount, actualCount, stdout)
			}
		})
	}
}

// BenchmarkQueryVsCat compares query and cat throughput on large files.
func BenchmarkQueryVsCat(b *testing.B) {
	b.Run("cat_large_file", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("cat", "../../tests/testdata/large.avro")
		}
	})

	b.Run("query_match_all", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/large.avro", "record.timestamp > 0")
		}
	})

	b.Run("query_match_half", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/large.avro", "record.timestamp > 1234572890")
		}
	})

	b.Run("query_match_few", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/large.avro", "record.timestamp > 1234577890")
		}
	})
}

// BenchmarkQueryFilterTypes benchmarks different filter types.
func BenchmarkQueryFilterTypes(b *testing.B) {
	b.Run("numeric_comparison", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/large.avro", "record.timestamp > 1234567890")
		}
	})

	b.Run("string_contains", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/large.avro", "record.level.contains('INFO')")
		}
	})

	b.Run("complex_logic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/large.avro", "record.timestamp > 1234567890 && record.level == 'INFO'")
		}
	})
}

// BenchmarkQuerySelectivity benchmarks query at various filter selectivities.
func BenchmarkQuerySelectivity(b *testing.B) {
	b.Run("select_100_percent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/users.avro", "record.age > 0")
		}
	})

	b.Run("select_66_percent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/users.avro", "record.age > 25")
		}
	})

	b.Run("select_33_percent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/users.avro", "record.age > 30")
		}
	})

	b.Run("select_0_percent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runGavro("query", "../../tests/testdata/users.avro", "record.age > 100")
		}
	})
}

// TestQueryLimitFlag verifies the --limit flag limits output records.
func TestQueryLimitFlag(t *testing.T) {
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 0", "--limit", "1")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 line with --limit 1, got %d", len(lines))
	}
}

// TestQueryLimitWithFilter verifies --limit applies after filtering.
func TestQueryLimitWithFilter(t *testing.T) {
	// age > 28 matches Alice(30) and Charlie(35) — limit 1 should return only Alice
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 28", "--limit", "1")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	var record map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if record["name"] != "Alice" {
		t.Errorf("Expected Alice, got %v", record["name"])
	}
}

// TestQueryCountFlag verifies the --count flag outputs only the match count.
func TestQueryCountFlag(t *testing.T) {
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 28", "--count")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if strings.TrimSpace(stdout) != "2" {
		t.Errorf("Expected 2, got %s", strings.TrimSpace(stdout))
	}
}

// TestQueryCountNoMatches verifies --count returns 0 when no records match.
func TestQueryCountNoMatches(t *testing.T) {
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 100", "--count")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if strings.TrimSpace(stdout) != "0" {
		t.Errorf("Expected 0, got %s", strings.TrimSpace(stdout))
	}
}

// TestQueryCountWithLimit verifies --count and --limit work together.
func TestQueryCountWithLimit(t *testing.T) {
	stdout, stderr, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 28", "--count", "--limit", "1")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if strings.TrimSpace(stdout) != "1" {
		t.Errorf("Expected 1, got %s", strings.TrimSpace(stdout))
	}
}

// TestQueryPrettyFlag verifies the --pretty flag produces indented output.
func TestQueryPrettyFlag(t *testing.T) {
	stdout, _, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 28", "--pretty")

	if exitCode != 0 {
		t.Fatal("query --pretty failed")
	}

	if !strings.Contains(stdout, "  \"age\"") {
		t.Error("Output should contain indented fields")
	}

	if !strings.Contains(stdout, "}\n\n{") {
		t.Error("Records should be separated by blank lines in pretty mode")
	}

	records := strings.Split(strings.TrimSpace(stdout), "\n\n")
	if len(records) != 2 {
		t.Errorf("Expected 2 matching records separated by blank lines, got %d", len(records))
	}

	for i, record := range records {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(record), &obj); err != nil {
			t.Errorf("Record %d is not valid JSON: %v", i, err)
		}
	}
}

// TestQueryPrettyShortFlag verifies the -p short flag produces indented output.
func TestQueryPrettyShortFlag(t *testing.T) {
	stdout, _, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 28", "-p")

	if exitCode != 0 {
		t.Fatal("query -p failed")
	}

	if !strings.Contains(stdout, "  \"age\"") {
		t.Error("Output should contain indented fields with -p flag")
	}
}

// TestQueryCompactByDefault verifies compact JSON output is the default.
func TestQueryCompactByDefault(t *testing.T) {
	stdout, _, exitCode := runGavro("query", "../../tests/testdata/users.avro", "record.age > 28")

	if exitCode != 0 {
		t.Fatal("query failed")
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	for i, line := range lines {
		if !strings.HasPrefix(line, "{") {
			t.Errorf("Line %d should start with '{' in compact mode", i)
		}
	}

	if strings.Contains(stdout, "\n\n") {
		t.Error("Should not have blank lines between records in compact mode")
	}
}
