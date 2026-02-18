package fuzz

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// FuzzAvroReader tests the reader against random input data.
func FuzzAvroReader(f *testing.F) {
	seedFiles := []string{
		"../../tests/testdata/users.avro",
		"../../tests/testdata/complex.avro",
		"../../tests/testdata/empty.avro",
	}

	for _, file := range seedFiles {
		if data, err := os.ReadFile(file); err == nil {
			f.Add(data)
		}
	}

	f.Add([]byte{})                                      // empty file
	f.Add([]byte("Obj\x01"))                             // truncated Avro magic
	f.Add([]byte("Obj\x01" + string(make([]byte, 100)))) // magic + garbage
	f.Add(make([]byte, 1024))                            // all zeros

	f.Add([]byte("NOT_AVRO_HEADER"))

	f.Fuzz(func(t *testing.T, data []byte) {
		tmpFile := filepath.Join(t.TempDir(), "fuzz.avro")
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			t.Skip()
		}

		cmd := exec.Command("/tmp/gavro", "cat", tmpFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Errorf("Program panicked on input: %v", stderrStr)
		}
		if bytes.Contains([]byte(stderrStr), []byte("runtime error")) {
			t.Errorf("Runtime error on input: %v", stderrStr)
		}

		if cmd.ProcessState != nil {
			exitCode := cmd.ProcessState.ExitCode()
			if exitCode < -1 {
				t.Errorf("Unexpected exit code %d (possible crash)", exitCode)
			}
		}
	})
}

// FuzzAvroMutation mutates valid Avro files and tests robustness.
func FuzzAvroMutation(f *testing.F) {
	validData, err := os.ReadFile("../../tests/testdata/users.avro")
	if err != nil {
		f.Skip("Cannot read valid test file")
	}

	f.Add(validData, 0, byte(0))
	f.Add(validData, 10, byte(255))
	f.Add(validData, len(validData)/2, byte(128))

	f.Fuzz(func(t *testing.T, data []byte, pos int, val byte) {
		if len(data) == 0 {
			t.Skip()
		}

		mutated := make([]byte, len(data))
		copy(mutated, data)

		if pos >= 0 && pos < len(mutated) {
			mutated[pos] = val
		}

		tmpFile := filepath.Join(t.TempDir(), "mutated.avro")
		if err := os.WriteFile(tmpFile, mutated, 0644); err != nil {
			t.Skip()
		}

		cmd := exec.Command("/tmp/gavro", "cat", tmpFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Errorf("Panic on mutated file at pos %d: %v", pos, stderrStr)
		}
	})
}

// FuzzAvroTruncation tests truncated Avro files at various cut points.
func FuzzAvroTruncation(f *testing.F) {
	validData, err := os.ReadFile("../../tests/testdata/users.avro")
	if err != nil {
		f.Skip("Cannot read valid test file")
	}

	for i := 0; i <= 100; i += 10 {
		cutPos := len(validData) * i / 100
		f.Add(validData[:cutPos])
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		tmpFile := filepath.Join(t.TempDir(), "truncated.avro")
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			t.Skip()
		}

		cmd := exec.Command("/tmp/gavro", "cat", tmpFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Errorf("Panic on truncated file: %v", stderrStr)
		}
		if bytes.Contains([]byte(stderrStr), []byte("runtime error")) {
			t.Errorf("Runtime error on truncated file: %v", stderrStr)
		}
	})
}

// FuzzAvroLargeInput tests that large inputs do not cause crashes or hangs.
func FuzzAvroLargeInput(f *testing.F) {
	sizes := []int{1024, 10 * 1024, 100 * 1024, 1024 * 1024}

	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 10*1024*1024 { // limit to 10MB
			t.Skip()
		}

		tmpFile := filepath.Join(t.TempDir(), "large.avro")
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			t.Skip()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "/tmp/gavro", "cat", tmpFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Error("Panic on large input")
		}
	})
}

// FuzzAvroSpecialBytes tests known-problematic byte patterns.
func FuzzAvroSpecialBytes(f *testing.F) {
	specialCases := [][]byte{
		{0x00, 0x00, 0x00, 0x00},             // all zeros
		{0xFF, 0xFF, 0xFF, 0xFF},             // all ones
		{0x4F, 0x62, 0x6A, 0x01},             // Avro magic "Obj\x01"
		{0x4F, 0x62, 0x6A, 0x01, 0x00},       // Avro magic + null
		bytes.Repeat([]byte{0x41}, 100),      // repeated bytes
		bytes.Repeat([]byte{0x00, 0xFF}, 50), // alternating bytes
	}

	for _, data := range specialCases {
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		tmpFile := filepath.Join(t.TempDir(), "special.avro")
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			t.Skip()
		}

		cmd := exec.Command("/tmp/gavro", "cat", tmpFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Errorf("Panic on special bytes: %x", data[:min(len(data), 20)])
		}
	})
}

// TestFuzzPrerequisites checks that the binary and test data are present.
func TestFuzzPrerequisites(t *testing.T) {
	if _, err := os.Stat("/tmp/gavro"); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", "/tmp/gavro", "../../main.go")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build gavro: %v", err)
		}
	}

	testFiles := []string{
		"../../tests/testdata/users.avro",
		"../../tests/testdata/complex.avro",
		"../../tests/testdata/empty.avro",
	}

	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			cmd := exec.Command("go", "run", "../../tests/testdata/generate.go")
			cmd.Dir = "../.."
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to generate test data: %v", err)
			}
			break
		}
	}
}
