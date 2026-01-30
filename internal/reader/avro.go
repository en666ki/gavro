package reader

import (
	"fmt"
	"io"
	"os"

	"github.com/hamba/avro/v2"
	"github.com/hamba/avro/v2/ocf"
)

// AvroReader читает записи из Avro OCF файла
type AvroReader struct {
	file   *os.File
	reader *ocf.Decoder
	schema avro.Schema
}

// NewAvroReader создает новый reader для Avro файла
func NewAvroReader(filePath string) (*AvroReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}

	// Создаем OCF decoder который читает схему автоматически
	reader, err := ocf.NewDecoder(file)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("cannot create avro decoder: %w", err)
	}

	return &AvroReader{
		file:   file,
		reader: reader,
		schema: reader.Schema(),
	}, nil
}

// Read читает следующую запись из Avro файла
func (r *AvroReader) Read() (Record, error) {
	if !r.reader.HasNext() {
		return nil, io.EOF
	}

	var record map[string]interface{}
	if err := r.reader.Decode(&record); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return Record(record), nil
}

// Close закрывает файл
func (r *AvroReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// Schema возвращает схему Avro (может быть полезно для будущих команд)
func (r *AvroReader) Schema() avro.Schema {
	return r.schema
}
