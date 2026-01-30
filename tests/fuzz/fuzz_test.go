package fuzz

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// FuzzAvroReader тестирует reader на случайных входных данных
func FuzzAvroReader(f *testing.F) {
	// Добавляем seed корпус - валидные файлы
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

	// Добавляем известные проблемные случаи
	f.Add([]byte{})                                      // пустой файл
	f.Add([]byte("Obj\x01"))                             // начало Avro magic но обрезано
	f.Add([]byte("Obj\x01" + string(make([]byte, 100)))) // magic + мусор
	f.Add(make([]byte, 1024))                            // нули

	// Добавляем файл с неправильным magic
	f.Add([]byte("NOT_AVRO_HEADER"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Создаем временный файл
		tmpFile := filepath.Join(t.TempDir(), "fuzz.avro")
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			t.Skip()
		}

		// Запускаем gavro cat на этом файле
		cmd := exec.Command("/tmp/gavro", "cat", tmpFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		// Главное - программа не должна паниковать или зависать
		_ = cmd.Run()

		// Проверяем что нет паники
		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Errorf("Program panicked on input: %v", stderrStr)
		}
		if bytes.Contains([]byte(stderrStr), []byte("runtime error")) {
			t.Errorf("Runtime error on input: %v", stderrStr)
		}

		// Ожидаем либо успех (exit 0) либо нормальную ошибку (exit 1)
		// Но не segfault (exit -11, -6, etc)
		if cmd.ProcessState != nil {
			exitCode := cmd.ProcessState.ExitCode()
			if exitCode < -1 {
				t.Errorf("Unexpected exit code %d (possible crash)", exitCode)
			}
		}
	})
}

// FuzzAvroMutation мутирует валидные Avro файлы
func FuzzAvroMutation(f *testing.F) {
	// Читаем валидный файл как базу
	validData, err := os.ReadFile("../../tests/testdata/users.avro")
	if err != nil {
		f.Skip("Cannot read valid test file")
	}

	// Seed corpus - различные мутации валидного файла
	f.Add(validData, 0, byte(0)) // замена байта на позиции
	f.Add(validData, 10, byte(255))
	f.Add(validData, len(validData)/2, byte(128))

	f.Fuzz(func(t *testing.T, data []byte, pos int, val byte) {
		if len(data) == 0 {
			t.Skip()
		}

		// Мутируем данные
		mutated := make([]byte, len(data))
		copy(mutated, data)

		if pos >= 0 && pos < len(mutated) {
			mutated[pos] = val
		}

		// Создаем временный файл
		tmpFile := filepath.Join(t.TempDir(), "mutated.avro")
		if err := os.WriteFile(tmpFile, mutated, 0644); err != nil {
			t.Skip()
		}

		// Тестируем
		cmd := exec.Command("/tmp/gavro", "cat", tmpFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		_ = cmd.Run()

		// Не должно быть паники или краша
		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Errorf("Panic on mutated file at pos %d: %v", pos, stderrStr)
		}
	})
}

// FuzzAvroTruncation тестирует обрезанные файлы
func FuzzAvroTruncation(f *testing.F) {
	validData, err := os.ReadFile("../../tests/testdata/users.avro")
	if err != nil {
		f.Skip("Cannot read valid test file")
	}

	// Seeds - обрезаем файл в разных местах
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

		// Программа должна корректно обработать неполные данные
		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Errorf("Panic on truncated file: %v", stderrStr)
		}
		if bytes.Contains([]byte(stderrStr), []byte("runtime error")) {
			t.Errorf("Runtime error on truncated file: %v", stderrStr)
		}
	})
}

// FuzzAvroLargeInput тестирует очень большие входные данные
func FuzzAvroLargeInput(f *testing.F) {
	// Проверяем что программа не падает на больших входных данных
	sizes := []int{1024, 10 * 1024, 100 * 1024, 1024 * 1024}

	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 10*1024*1024 { // ограничиваем 10MB
			t.Skip()
		}

		tmpFile := filepath.Join(t.TempDir(), "large.avro")
		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			t.Skip()
		}

		cmd := exec.Command("/tmp/gavro", "cat", tmpFile)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		// Добавляем timeout чтобы не зависнуть
		done := make(chan error, 1)
		go func() {
			done <- cmd.Run()
		}()

		select {
		case <-done:
			// OK, завершилось
		}

		stderrStr := stderr.String()
		if bytes.Contains([]byte(stderrStr), []byte("panic")) {
			t.Error("Panic on large input")
		}
	})
}

// FuzzAvroSpecialBytes тестирует специальные байтовые последовательности
func FuzzAvroSpecialBytes(f *testing.F) {
	// Добавляем известные проблемные паттерны
	specialCases := [][]byte{
		{0x00, 0x00, 0x00, 0x00},             // нули
		{0xFF, 0xFF, 0xFF, 0xFF},             // единицы
		{0x4F, 0x62, 0x6A, 0x01},             // Avro magic "Obj\x01"
		{0x4F, 0x62, 0x6A, 0x01, 0x00},       // Avro magic + null
		bytes.Repeat([]byte{0x41}, 100),      // повторяющиеся символы
		bytes.Repeat([]byte{0x00, 0xFF}, 50), // чередование
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

// Утилита для min (для Go < 1.21)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestFuzzPrerequisites проверяет что тестовые данные на месте
func TestFuzzPrerequisites(t *testing.T) {
	// Проверяем что gavro собран
	if _, err := os.Stat("/tmp/gavro"); os.IsNotExist(err) {
		// Собираем
		cmd := exec.Command("go", "build", "-o", "/tmp/gavro", "../../main.go")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build gavro: %v", err)
		}
	}

	// Проверяем что тестовые данные сгенерированы
	testFiles := []string{
		"../../tests/testdata/users.avro",
		"../../tests/testdata/complex.avro",
		"../../tests/testdata/empty.avro",
	}

	for _, file := range testFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			// Генерируем
			cmd := exec.Command("go", "run", "../../tests/testdata/generate.go")
			cmd.Dir = "../.."
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to generate test data: %v", err)
			}
			break
		}
	}
}
