package reader

// Record представляет одну запись из источника данных
type Record map[string]interface{}

// Reader - интерфейс для чтения записей из различных источников
type Reader interface {
	// Read читает следующую запись. Возвращает io.EOF когда данные закончились
	Read() (Record, error)

	// Close закрывает ресурсы
	Close() error
}
