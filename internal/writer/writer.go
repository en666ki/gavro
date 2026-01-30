package writer

// Writer - интерфейс для записи записей в различные форматы
type Writer interface {
	// Write записывает одну запись
	Write(record map[string]interface{}) error

	// Flush сбрасывает буферы (если есть)
	Flush() error
}
