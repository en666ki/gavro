# Tests

Этот проект содержит комплексное тестовое покрытие с e2e и fuzzy тестами.

## Структура

```
tests/
├── e2e/              # End-to-end тесты CLI
│   └── cat_test.go   # Тесты команды cat
├── fuzz/             # Fuzzy тесты
│   └── fuzz_test.go  # Тестирование на случайных и поврежденных данных
└── testdata/         # Тестовые Avro файлы
    └── generate.go   # Генератор тестовых данных
```

## Запуск тестов

### E2E тесты

```bash
# Полный прогон
go test -v ./tests/e2e/...

# Быстрый прогон (без длительных тестов)
go test -v ./tests/e2e/... -short

# Конкретный тест
go test -v ./tests/e2e/... -run TestCatSimpleUsers

# С бенчмарками
go test -v ./tests/e2e/... -bench=.
```

### Fuzzy тесты

```bash
# Запуск всех fuzzy тестов (seed only)
go test -v ./tests/fuzz/...

# Fuzzing на 30 секунд
go test -v ./tests/fuzz/... -fuzz=FuzzAvroReader -fuzztime=30s

# Fuzzing с конкретным тестом
go test -v ./tests/fuzz/... -fuzz=FuzzAvroMutation -fuzztime=1m

# Fuzzing до первой ошибки
go test -v ./tests/fuzz/... -fuzz=FuzzAvroTruncation
```

### Все тесты

```bash
# Запустить все тесты
go test -v ./tests/...

# С покрытием
go test -v ./tests/... -cover

# С race detector
go test -v ./tests/... -race
```

## E2E тесты (tests/e2e/)

Тестируют полный флоу работы CLI через exec:

### Покрытие

- ✅ **Чтение валидных файлов** - простые и сложные схемы
- ✅ **JSON Lines формат** - корректность вывода
- ✅ **Интеграция с jq** - pipe и фильтрация
- ✅ **Обработка ошибок** - файл не найден, невалидный формат
- ✅ **Поврежденные файлы** - битые заголовки, обрезанные данные, мусор
- ✅ **Большие файлы** - производительность на 10000 записей
- ✅ **Утечки ресурсов** - memory leaks и file handles
- ✅ **Help команды** - корректность вывода справки
- ✅ **Различные пути** - относительные, с .., и т.д.

### Тестовые данные

Автоматически генерируются через `go run tests/testdata/generate.go`:

- `users.avro` - простая схема с 3 пользователями
- `complex.avro` - сложная схема с массивами, map, вложенными records, nullable
- `empty.avro` - пустой файл (только заголовок)
- `large.avro` - 10000 записей для тестов производительности
- `bad_magic.avro` - неправильный magic header
- `totally_empty.avro` - полностью пустой файл
- `truncated.avro` - обрезанный файл
- `garbage.avro` - случайный мусор

## Fuzzy тесты (tests/fuzz/)

Тестируют устойчивость к некорректным входным данным:

### Стратегии

1. **FuzzAvroReader** - полностью случайные данные
   - Пустые файлы
   - Частичные magic headers
   - Случайные байты
   - Проверяет что программа не падает с паникой

2. **FuzzAvroMutation** - мутации валидных файлов
   - Замена байтов на случайных позициях
   - Тестирует boundary conditions

3. **FuzzAvroTruncation** - обрезанные файлы
   - Файлы обрезанные в разных местах
   - Проверяет обработку EOF

4. **FuzzAvroLargeInput** - очень большие данные
   - Файлы до 10MB
   - Проверяет отсутствие зависаний

5. **FuzzAvroSpecialBytes** - специальные последовательности
   - Нули, единицы, чередования
   - Avro magic + мусор
   - Повторяющиеся паттерны

### Гарантии

Fuzzy тесты гарантируют что gavro:
- ✅ Не падает с panic на любых входных данных
- ✅ Не имеет runtime errors (nil dereference, index out of bounds)
- ✅ Не зависает на больших или специальных входных данных
- ✅ Возвращает нормальные exit codes (0 или 1, не segfault)
- ✅ Выдает читаемые ошибки вместо крашей

## Continuous Integration

Рекомендуемый CI pipeline:

```bash
# Quick check
go test ./tests/... -short -race

# Full suite
go test ./tests/... -race -cover

# Fuzzing (длительный)
go test ./tests/fuzz/... -fuzz=FuzzAvroReader -fuzztime=5m
```

## Добавление новых тестов

### E2E тесты

1. Создайте тестовые данные в `tests/testdata/generate.go`
2. Добавьте тест функцию в `tests/e2e/cat_test.go`
3. Используйте `runGavro()` helper для запуска команд

### Fuzzy тесты

1. Добавьте новую `FuzzXxx` функцию в `tests/fuzz/fuzz_test.go`
2. Определите seed corpus через `f.Add()`
3. Проверяйте отсутствие паник и runtime errors

## Производительность

Текущие бенчмарки (примерно):

- **Чтение 10000 записей**: ~50-100ms
- **JSON Lines вывод**: streaming, O(1) память
- **Fuzzing speed**: ~600-1000 exec/sec

Запустить бенчмарки:

```bash
go test ./tests/e2e/... -bench=BenchmarkCatLargeFile -benchmem
```
