package reader

import (
	"fmt"
	"io"
	"os"

	"github.com/hamba/avro/v2"
	"github.com/hamba/avro/v2/ocf"
)

// AvroReader reads records from an Avro OCF file.
type AvroReader struct {
	file   *os.File
	reader *ocf.Decoder
	schema avro.Schema
}

// NewAvroReader opens an Avro OCF file and returns a reader for it.
// If filePath is "-", it reads from stdin.
func NewAvroReader(filePath string) (*AvroReader, error) {
	var r io.Reader
	var file *os.File

	if filePath == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("cannot open file: %w", err)
		}
		r = f
		file = f
	}

	reader, err := ocf.NewDecoder(r)
	if err != nil {
		if file != nil {
			file.Close()
		}
		return nil, fmt.Errorf("cannot create avro decoder: %w", err)
	}

	return &AvroReader{
		file:   file,
		reader: reader,
		schema: reader.Schema(),
	}, nil
}

// Read returns the next record from the Avro file.
func (r *AvroReader) Read() (Record, error) {
	if !r.reader.HasNext() {
		if err := r.reader.Error(); err != nil {
			return nil, fmt.Errorf("read error: %w", err)
		}
		return nil, io.EOF
	}

	var record map[string]interface{}
	if err := r.reader.Decode(&record); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return Record(record), nil
}

// Close closes the underlying file.
func (r *AvroReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// Schema returns the Avro schema read from the file header.
func (r *AvroReader) Schema() avro.Schema {
	return r.schema
}
