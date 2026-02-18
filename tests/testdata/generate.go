package main

import (
	"log"
	"os"

	"github.com/hamba/avro/v2"
	"github.com/hamba/avro/v2/ocf"
)

func main() {
	generateSimpleUsers()
	generateComplexSchema()
	generateEmptyFile()
	generateLargeFile()
	generateCorruptedFiles()
	log.Println("All test files generated")
}

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

	s, err := avro.Parse(schema)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("tests/testdata/users.avro")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc, err := ocf.NewEncoderWithSchema(s, f, ocf.WithCodec(ocf.Deflate))
	if err != nil {
		log.Fatal(err)
	}

	users := []map[string]interface{}{
		{"name": "Alice", "age": int32(30), "email": "alice@example.com"},
		{"name": "Bob", "age": int32(25), "email": "bob@example.com"},
		{"name": "Charlie", "age": int32(35), "email": "charlie@example.com"},
	}

	for _, u := range users {
		if err := enc.Encode(u); err != nil {
			log.Fatal(err)
		}
	}
	if err := enc.Flush(); err != nil {
		log.Fatal(err)
	}
}

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

	s, err := avro.Parse(schema)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("tests/testdata/complex.avro")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc, err := ocf.NewEncoderWithSchema(s, f, ocf.WithCodec(ocf.Null))
	if err != nil {
		log.Fatal(err)
	}

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
		if err := enc.Encode(r); err != nil {
			log.Fatal(err)
		}
	}
	if err := enc.Flush(); err != nil {
		log.Fatal(err)
	}
}

func generateEmptyFile() {
	schema := `{
		"type": "record",
		"name": "Empty",
		"fields": [{"name": "field", "type": "string"}]
	}`

	s, err := avro.Parse(schema)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("tests/testdata/empty.avro")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc, err := ocf.NewEncoderWithSchema(s, f, ocf.WithCodec(ocf.Null))
	if err != nil {
		log.Fatal(err)
	}
	if err := enc.Flush(); err != nil {
		log.Fatal(err)
	}
}

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

	s, err := avro.Parse(schema)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("tests/testdata/large.avro")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc, err := ocf.NewEncoderWithSchema(s, f, ocf.WithCodec(ocf.Snappy))
	if err != nil {
		log.Fatal(err)
	}

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
		if err := enc.Encode(record); err != nil {
			log.Fatal(err)
		}
	}
	if err := enc.Flush(); err != nil {
		log.Fatal(err)
	}
}

func generateCorruptedFiles() {
	if err := os.WriteFile("tests/testdata/bad_magic.avro", []byte("NOT_AVRO_FILE_HEADER"), 0644); err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("tests/testdata/totally_empty.avro", []byte{}, 0644); err != nil {
		log.Fatal(err)
	}

	data, err := os.ReadFile("tests/testdata/users.avro")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("tests/testdata/truncated.avro", data[:len(data)/2], 0644); err != nil {
		log.Fatal(err)
	}

	garbage := make([]byte, 1024)
	for i := range garbage {
		garbage[i] = byte(i % 256)
	}
	if err := os.WriteFile("tests/testdata/garbage.avro", garbage, 0644); err != nil {
		log.Fatal(err)
	}
}
