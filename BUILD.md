# Инструкции по сборке URnetwork Client для Windows

## Быстрый старт

### 1. Подготовка окружения

#### Windows

**Установка Go:**
1. Скачайте Go 1.20+ с https://golang.org/dl/
2. Запустите установщик
3. Проверьте установку: `go version`

**Установка GCC (обязательно для Fyne):**
1. Скачайте MSYS2 с https://www.msys2.org/
2. Установите MSYS2
3. Откройте MSYS2 терминал и выполните:
   ```bash
   pacman -Syu
   pacman -S mingw-w64-x86_64-gcc
   ```
4. Добавьте `C:\msys64\mingw64\bin` в переменную окружения PATH:
   - Откройте "Система" → "Дополнительные параметры системы"
   - Нажмите "Переменные среды"
   - В разделе "Системные переменные" найдите PATH
   - Добавьте `C:\msys64\mingw64\bin`

#### Linux (для кросс-компиляции)

```bash
# Ubuntu/Debian
sudo apt-get install gcc-mingw-w64-x86-64

# Arch Linux
sudo pacman -S mingw-w64-gcc
```

#### macOS (для кросс-компиляции)

```bash
brew install mingw-w64
```

### 2. Скачивание зависимостей

```bash
cd /путь/к/URnetworkPC
go mod download
go mod tidy
```

### 3. Сборка приложения

#### Вариант 1: Разработка (с консолью для отладки)
```bash
go build -o URnetworkPC.exe
```

#### Вариант 2: Продакшн (без консоли, оптимизированный)
```bash
go build -ldflags="-H windowsgui -s -w" -o URnetworkPC.exe
```

#### Вариант 3: Кросс-компиляция (с Linux/macOS)
```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags="-H windowsgui -s -w" -o URnetworkPC.exe
```

## Объяснение флагов сборки

### Флаги `-ldflags`

- **`-H windowsgui`** - Скрывает окно консоли при запуске GUI приложения в Windows
- **`-s`** - Удаляет таблицу символов (уменьшает размер)
- **`-w`** - Удаляет отладочную информацию DWARF (уменьшает размер)

Комбинация `-s -w` может уменьшить размер бинарника на 20-30%.

### Переменные окружения для кросс-компиляции

- **`GOOS=windows`** - целевая ОС
- **`GOARCH=amd64`** - архитектура (64-бит)
- **`CGO_ENABLED=1`** - включить CGO (требуется для Fyne)
- **`CC=x86_64-w64-mingw32-gcc`** - кросс-компилятор GCC

## Автоматизация сборки

### PowerShell скрипт (Windows)

Создайте файл `build.ps1`:

```powershell
# build.ps1
Write-Host "Building URnetworkPC..." -ForegroundColor Green

# Разработка
Write-Host "Building development version..." -ForegroundColor Yellow
go build -o URnetworkPC-dev.exe

# Продакшн
Write-Host "Building production version..." -ForegroundColor Yellow
go build -ldflags="-H windowsgui -s -w" -o URnetworkPC.exe

Write-Host "Build complete!" -ForegroundColor Green
Write-Host "Files created:" -ForegroundColor Cyan
Write-Host "  - URnetworkPC-dev.exe (with console)" -ForegroundColor White
Write-Host "  - URnetworkPC.exe (production)" -ForegroundColor White
```

Запуск:
```powershell
.\build.ps1
```

### Bash скрипт (Linux/macOS)

Создайте файл `build.sh`:

```bash
#!/bin/bash
# build.sh

echo "Building URnetworkPC for Windows..."

# Разработка
echo "Building development version..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -o URnetworkPC-dev.exe

# Продакшн
echo "Building production version..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
  go build -ldflags="-H windowsgui -s -w" -o URnetworkPC.exe

echo "Build complete!"
echo "Files created:"
echo "  - URnetworkPC-dev.exe (with console)"
echo "  - URnetworkPC.exe (production)"
```

Запуск:
```bash
chmod +x build.sh
./build.sh
```

### Makefile

Создайте `Makefile`:

```makefile
# Makefile for URnetworkPC

.PHONY: all build build-dev build-prod clean run test

# Переменные
BINARY_NAME=URnetworkPC.exe
BINARY_DEV=URnetworkPC-dev.exe
LDFLAGS=-ldflags="-H windowsgui -s -w"

# По умолчанию - продакшн сборка
all: build-prod

# Разработка (с консолью)
build-dev:
	@echo "Building development version..."
	go build -o $(BINARY_DEV)

# Продакшн (без консоли, оптимизированный)
build-prod:
	@echo "Building production version..."
	go build $(LDFLAGS) -o $(BINARY_NAME)

# Обе версии
build: build-dev build-prod

# Запуск
run:
	go run main.go

# Тесты
test:
	go test -v ./...

# Очистка
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME) $(BINARY_DEV)

# Форматирование
fmt:
	go fmt ./...

# Проверка
vet:
	go vet ./...

# Обновление зависимостей
deps:
	go mod download
	go mod tidy
```

Использование:
```bash
make build-prod  # Продакшн сборка
make build-dev   # Разработка
make build       # Обе версии
make clean       # Очистка
```

## Оптимизация размера бинарника

### 1. Базовая оптимизация (уже включена)
```bash
go build -ldflags="-s -w" -o URnetworkPC.exe
```

### 2. UPX компрессия (дополнительно)

Установите UPX:
- Windows: скачайте с https://upx.github.io/
- Linux: `sudo apt-get install upx`
- macOS: `brew install upx`

Сжатие:
```bash
upx --best --lzma URnetworkPC.exe
```

Это может уменьшить размер еще на 50-70%, но увеличит время запуска.

### 3. Сравнение размеров

| Вариант сборки | Примерный размер |
|----------------|------------------|
| Без оптимизации | ~25-30 MB |
| С `-ldflags="-s -w"` | ~18-22 MB |
| + UPX | ~8-12 MB |

## Проверка сборки

После сборки проверьте, что бинарник работает:

1. **Запустите приложение:**
   ```bash
   .\URnetworkPC.exe
   ```

2. **Проверьте зависимости (Windows):**
   ```powershell
   dumpbin /dependents URnetworkPC.exe
   ```

3. **Проверьте размер:**
   ```bash
   ls -lh URnetworkPC.exe
   ```

## Решение проблем сборки

### Ошибка: "gcc: command not found"

**Причина:** Не установлен GCC компилятор

**Решение:**
1. Установите MinGW-w64 через MSYS2 (см. выше)
2. Добавьте путь к gcc в PATH

### Ошибка: "cgo: C compiler not found"

**Причина:** CGO не может найти компилятор C

**Решение:**
```bash
# Windows
set CGO_ENABLED=1
set CC=gcc

# Linux/macOS
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
```

### Ошибка: "undefined reference to..."

**Причина:** Проблемы с линковкой библиотек

**Решение:**
```bash
# Переустановите зависимости
go clean -cache
go mod tidy
go build -o URnetworkPC.exe
```

### Ошибка при кросс-компиляции: "cannot find package"

**Причина:** Не все зависимости поддерживают кросс-компиляцию с CGO

**Решение:**
1. Соберите на целевой платформе (Windows)
2. Используйте Docker с Windows окружением
3. Используйте GitHub Actions / CI/CD

## CI/CD Pipeline (GitHub Actions)

Создайте `.github/workflows/build.yml`:

```yaml
name: Build URnetworkPC

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: windows-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.20'
    
    - name: Install dependencies
      run: |
        go mod download
        go mod tidy
    
    - name: Build development
      run: go build -o URnetworkPC-dev.exe
    
    - name: Build production
      run: go build -ldflags="-H windowsgui -s -w" -o URnetworkPC.exe
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: URnetworkPC-builds
        path: |
          URnetworkPC.exe
          URnetworkPC-dev.exe
```

## Дополнительные ресурсы

- **Fyne документация:** https://docs.fyne.io/
- **Go cross-compilation:** https://golang.org/doc/install/source#environment
- **URnetwork GitHub:** https://github.com/urnetwork/
- **MinGW-w64:** https://www.mingw-w64.org/
