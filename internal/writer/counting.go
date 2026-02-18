package writer

// CountingWriter is a no-op writer for --count mode.
// Records are consumed but not serialized; the count comes from the processor.
type CountingWriter struct{}

func (w *CountingWriter) Write(record interface{}) error { return nil }
func (w *CountingWriter) Flush() error                   { return nil }
