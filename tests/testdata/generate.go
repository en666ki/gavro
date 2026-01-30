package main

import (
	"log"
	"os"

	"github.com/hamba/avro/v2"
	"github.com/hamba/avro/v2/ocf"
)

// Генерирует различные тестовые Avro файлы
func main() {
	generateSimpleUsers()
	generateComplexSchema()
	generateEmptyFile()
	generateLargeFile()
	generateCorruptedFiles()
	log.Println("All test files generated")
}

// Простая схема с пользователями
func generateSimpleUsers() {
	schema := `{
		"type": "record",
		"name": "User",
		"fields": [
			{"name": "name", "type": "string"},
			{"name": "age", "type": "int"},
			{"name": "email", "type": "string"}
		]
	}`

	s, _ := avro.Parse(schema)
	f, _ := os.Create("tests/testdata/users.avro")
	defer f.Close()

	enc, _ := ocf.NewEncoderWithSchema(s, f, ocf.WithCodec(ocf.Deflate))

	users := []map[string]interface{}{
		{"name": "Alice", "age": int32(30), "email": "alice@example.com"},
		{"name": "Bob", "age": int32(25), "email": "bob@example.com"},
		{"name": "Charlie", "age": int32(35), "email": "charlie@example.com"},
	}

	for _, u := range users {
		enc.Encode(u)
	}
	enc.Flush()
}

// Сложная вложенная схема
func generateComplexSchema() {
	schema := `{
		"type": "record",
		"name": "ComplexRecord",
		"fields": [
			{"name": "id", "type": "long"},
			{"name": "name", "type": "string"},
			{"name": "tags", "type": {"type": "array", "items": "string"}},
			{"name": "metadata", "type": {"type": "map", "values": "string"}},
			{"name": "score", "type": ["null", "double"], "default": null},
			{"name": "nested", "type": {
				"type": "record",
				"name": "NestedRecord",
				"fields": [
					{"name": "field1", "type": "string"},
					{"name": "field2", "type": "int"}
				]
			}}
		]
	}`

	s, _ := avro.Parse(schema)
	f, _ := os.Create("tests/testdata/complex.avro")
	defer f.Close()

	enc, _ := ocf.NewEncoderWithSchema(s, f, ocf.WithCodec(ocf.Null))

	records := []map[string]interface{}{
		{
			"id":   int64(1),
			"name": "First",
			"tags": []interface{}{"tag1", "tag2", "tag3"},
			"metadata": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			"score": map[string]interface{}{"double": 3.14},
			"nested": map[string]interface{}{
				"field1": "nested_value",
				"field2": int32(42),
			},
		},
		{
			"id":       int64(2),
			"name":     "Second",
			"tags":     []interface{}{},
			"metadata": map[string]interface{}{},
			"score":    nil,
			"nested": map[string]interface{}{
				"field1": "another",
				"field2": int32(100),
			},
		},
	}

	for _, r := range records {
		enc.Encode(r)
	}
	enc.Flush()
}

// Пустой файл (только заголовок)
func generateEmptyFile() {
	schema := `{
		"type": "record",
		"name": "Empty",
		"fields": [{"name": "field", "type": "string"}]
	}`

	s, _ := avro.Parse(schema)
	f, _ := os.Create("tests/testdata/empty.avro")
	defer f.Close()

	enc, _ := ocf.NewEncoderWithSchema(s, f, ocf.WithCodec(ocf.Null))
	enc.Flush()
}

// Большой файл для тестирования производительности
func generateLargeFile() {
	schema := `{
		"type": "record",
		"name": "LogEntry",
		"fields": [
			{"name": "timestamp", "type": "long"},
			{"name": "level", "type": "string"},
			{"name": "message", "type": "string"},
			{"name": "data", "type": {"type": "map", "values": "string"}}
		]
	}`

	s, _ := avro.Parse(schema)
	f, _ := os.Create("tests/testdata/large.avro")
	defer f.Close()

	enc, _ := ocf.NewEncoderWithSchema(s, f, ocf.WithCodec(ocf.Snappy))

	// 10000 записей
	for i := 0; i < 10000; i++ {
		record := map[string]interface{}{
			"timestamp": int64(1234567890 + i),
			"level":     "INFO",
			"message":   "This is a log message with some data",
			"data": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		}
		enc.Encode(record)
	}
	enc.Flush()
}

// Поврежденные файлы для fuzzy тестирования
func generateCorruptedFiles() {
	// Файл с неправильным magic header
	os.WriteFile("tests/testdata/bad_magic.avro", []byte("NOT_AVRO_FILE_HEADER"), 0644)

	// Пустой файл
	os.WriteFile("tests/testdata/totally_empty.avro", []byte{}, 0644)

	// Файл с обрезанными данными
	data, _ := os.ReadFile("tests/testdata/users.avro")
	os.WriteFile("tests/testdata/truncated.avro", data[:len(data)/2], 0644)

	// Случайный мусор
	garbage := make([]byte, 1024)
	for i := range garbage {
		garbage[i] = byte(i % 256)
	}
	os.WriteFile("tests/testdata/garbage.avro", garbage, 0644)
}
