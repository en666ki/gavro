# Release Checklist

## Подготовка к первому релизу

### 1. Инициализация git репозитория (если еще не сделано)

```bash
git init
git add .
git commit -m "Initial commit: gavro CLI tool for Avro files"
```

### 2. Добавление remote и push

```bash
git remote add origin https://github.com/en666ki/gavro.git
git branch -M main
git push -u origin main
```

### 3. Создание первого релиза

```bash
# Создать и запушить тег
git tag -a v0.1.0 -m "Initial release v0.1.0"
git push origin v0.1.0
```

### 4. (Опционально) Создать GitHub Release

На GitHub:
1. Перейти в Releases → Create a new release
2. Выбрать тег `v0.1.0`
3. Название: `v0.1.0 - Initial Release`
4. Описание:

```markdown
## gavro v0.1.0 - Initial Release

First stable release of gavro - a fast CLI tool for working with Apache Avro files.

### Features
- ✅ `gavro cat` command for outputting Avro files as JSON Lines
- ✅ Compatible with jq and standard UNIX tools
- ✅ Streaming processing with minimal memory usage
- ✅ Comprehensive error handling
- ✅ Full test coverage (e2e + fuzzing)

### Installation
```bash
go install github.com/en666ki/gavro@v0.1.0
```

### Usage
```bash
# Output Avro file as JSON Lines
gavro cat file.avro

# Pipe to jq
gavro cat file.avro | jq 'select(.age > 18)'
```

### What's Next
- Schema inspection command
- Query/filter capabilities
- Format conversion tools
```

## После релиза

Пользователи смогут установить gavro командой:

```bash
# Последняя версия
go install github.com/en666ki/gavro@latest

# Конкретная версия
go install github.com/en666ki/gavro@v0.1.0
```

## Версионирование

Проект использует [Semantic Versioning](https://semver.org/):

- **v0.1.x** - Patch releases (bug fixes)
- **v0.x.0** - Minor releases (new features, backward compatible)
- **vx.0.0** - Major releases (breaking changes)

## Создание новых релизов

### Patch release (v0.1.1)

```bash
# Фиксы багов, без новых фич
git tag -a v0.1.1 -m "Bug fixes"
git push origin v0.1.1
```

### Minor release (v0.2.0)

```bash
# Новые фичи, обратная совместимость
git tag -a v0.2.0 -m "Add schema command"
git push origin v0.2.0
```

### Major release (v1.0.0)

```bash
# Breaking changes или stable API
git tag -a v1.0.0 -m "First stable release"
git push origin v1.0.0
```

## Build с версией

Для сборки с правильной версией:

```bash
# Вручную указать версию
go build -ldflags "-X github.com/en666ki/gavro/cmd.Version=v0.1.0" -o gavro

# Автоматически из git tag
VERSION=$(git describe --tags --always)
go build -ldflags "-X github.com/en666ki/gavro/cmd.Version=$VERSION" -o gavro
```

## Проверка перед релизом

```bash
# Запустить все тесты
go test ./... -race -cover

# Fuzzing (минимум 1 минута)
go test ./tests/fuzz/... -fuzz=FuzzAvroReader -fuzztime=1m

# Проверить что билдится
go build

# Проверить версию
./gavro --version

# Проверить основной функционал
./gavro cat tests/testdata/users.avro | jq '.'
```

## CI/CD (опционально)

Создать `.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.21'

    - name: Run tests
      run: go test ./... -race -cover

    - name: Run fuzzing (short)
      run: go test ./tests/fuzz/... -fuzz=FuzzAvroReader -fuzztime=30s

    - name: Build
      run: go build -o gavro
```

## Go Module Proxy

После push тега, go module proxy автоматически индексирует новую версию.
Может занять несколько минут. Проверить статус:

```
https://proxy.golang.org/github.com/en666ki/gavro/@v/list
```
