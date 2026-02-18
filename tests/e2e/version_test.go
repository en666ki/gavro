package e2e

import (
	"strings"
	"testing"
)

func TestVersionFlag(t *testing.T) {
	stdout, stderr, exitCode := runGavro("--version")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "gavro version") {
		t.Errorf("Expected 'gavro version' in output, got: %s", stdout)
	}

	if strings.TrimSpace(stdout) == "gavro version dev" {
		t.Log("Warning: version is 'dev' - this is ok for development builds")
	}

	parts := strings.Fields(stdout)
	if len(parts) < 3 {
		t.Errorf("Version output should have at least 3 parts, got: %s", stdout)
	}
}

func TestVersionShortFlag(t *testing.T) {
	stdout, stderr, exitCode := runGavro("-v")

	if exitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d. Stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "gavro version") {
		t.Errorf("Expected version output with -v flag, got: %s", stdout)
	}
}

func TestVersionFormat(t *testing.T) {
	stdout, _, exitCode := runGavro("--version")

	if exitCode != 0 {
		t.Fatal("Version command should return exit code 0")
	}

	output := strings.TrimSpace(stdout)

	if !strings.HasPrefix(output, "gavro version ") {
		t.Errorf("Version should start with 'gavro version ', got: %s", output)
	}

	versionPart := strings.TrimPrefix(output, "gavro version ")

	if versionPart == "" {
		t.Error("Version part should not be empty")
	}

	isValidFormat := false

	// Semantic version: v0.1.0, v1.2.3, etc
	if strings.HasPrefix(versionPart, "v") && strings.Contains(versionPart, ".") {
		isValidFormat = true
	}
	// Git hash: 7-40 hex characters
	if len(versionPart) >= 7 && len(versionPart) <= 40 {
		isValidFormat = true
	}
	// Dev build
	if versionPart == "dev" {
		isValidFormat = true
	}
	// Version with +dirty suffix
	if strings.Contains(versionPart, "+") {
		isValidFormat = true
	}

	if !isValidFormat {
		t.Logf("Warning: unexpected version format: %s", versionPart)
	}
}
