package reader

// Record represents a single record from a data source.
type Record map[string]interface{}

// Reader reads records from a data source.
type Reader interface {
	// Read returns the next record. Returns io.EOF when all records are exhausted.
	Read() (Record, error)

	// Close releases any resources held by the reader.
	Close() error
}
