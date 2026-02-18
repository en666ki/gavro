package writer

// Writer writes records to an output destination.
type Writer interface {
	// Write writes a single record.
	Write(record interface{}) error

	// Flush flushes any buffered output.
	Flush() error
}
