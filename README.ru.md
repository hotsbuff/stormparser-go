# MPQ Parser для Heroes of the Storm

[English version](README.md)

Библиотека и CLI на Go для чтения файлов `.StormReplay`. Проект разбирает MPQ-архив и внутренние бинарные протоколы Blizzard без привязки к одной сборке игры.

Описания форматов вынесены в отдельный репозиторий: [hotsbuff/stormprotocol](https://github.com/hotsbuff/stormprotocol). В этом проекте они находятся в каталоге `schemas/` и используются при генерации кода и встраивании YAML-схем.

## Возможности

- MPQ user data и заголовки архива;
- хеш-таблицы, таблицы блоков и поиск файлов по имени;
- расшифровка таблиц и зашифрованных файлов;
- несжатые, sector-based и single-unit файлы;
- распаковка zlib, bzip2, LZMA, PKWARE, Huffman, Sparse и mono/stereo IMA ADPCM;
- типизированные сводки матчей, игроки, Toon, цвета и временные метки;
- versioned binary и bit-packed форматы Blizzard;
- `replay.header`, `replay.details`, `replay.initData`;
- игровые события, сообщения, события трекера и атрибуты;
- сохранение неизвестных versioned-полей;
- переносимые Kaitai- и YAML-схемы.

Внутренние потоки HotS не используют protobuf: игровые события и сообщения записаны в bit-packed формате, а события трекера — в versioned binary.

## Клонирование

Репозиторий схем подключён как Git submodule. Клонируйте проект вместе с submodules:

```bash
git clone --recurse-submodules https://github.com/hotsbuff/stormparser-go.git
cd stormparser-go
```

Если проект уже склонирован без submodules:

```bash
git submodule update --init --recursive
```

Чтобы обновить `schemas/` до последнего коммита `stormprotocol`:

```bash
git submodule update --remote schemas
```

После обновления зафиксируйте изменённую ссылку на submodule коммитом в основном репозитории.

## Требования

Для использования библиотеки и CLI нужен Go, указанный в `go.mod`.

Дополнительные инструменты нужны только для повторной генерации:

- [Kaitai Struct Compiler](https://kaitai.io/) — компиляция `.ksy` в Go;
- Python 3 — преобразование официальных `protocol<build>.py` в YAML.

Python не требуется во время выполнения готовой программы.

## CLI

Вывести версию реплея и используемую схему:

```bash
go run ./cmd/hotsreplay path/to/replay.StormReplay
```

Доступные параметры:

```text
-summary         Типизированная сводка матча с игроками и длительностью
-header          Полный replay.header
-details         replay.details
-initdata        replay.initData
-attributes      replay.attributes.events
-gameevents      replay.game.events
-messageevents   replay.message.events
-trackerevents   replay.tracker.events
-extract NAME    Извлечь файл из MPQ и записать его в stdout
```

Параметры декодирования можно комбинировать. Результат выводится в JSON:

```bash
go run ./cmd/hotsreplay -details -attributes path/to/replay.StormReplay
```

Извлечение исходного бинарного файла:

```bash
go run ./cmd/hotsreplay -extract replay.details path/to/replay.StormReplay > replay.details
```

## Использование из Go

```go
package main

import (
    "fmt"
    "log"

    "github.com/hotsbuff/stormparser-go/replay"
)

func main() {
    r, err := replay.Open("match.StormReplay")
    if err != nil {
        log.Fatal(err)
    }
    defer r.Close()

    fmt.Printf("game version: %+v\n", r.Version())

    summary, err := r.Summary()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("map: %s, players: %d, duration: %s\n",
        summary.Map, len(summary.Players), summary.Duration)

    events, err := r.TrackerEvents()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("tracker events: %d\n", len(events))
}
```

Для низкоуровневого доступа импортируется пакет `github.com/hotsbuff/stormparser-go/mpq`:

```go
archive, err := mpq.Open("match.StormReplay")
if err != nil {
    // обработать ошибку
}
defer archive.Close()

data, err := archive.ReadFile("replay.details")
```

## Схемы протокола

В каталоге `schemas/` находятся:

```text
schemas/
├── formats/          # Kaitai-схемы бинарных форматов
├── protocol/builds/  # YAML-описания сборок игры
└── tools/            # Генератор YAML из официального heroprotocol
```

Сгенерировать Go-код и обновить текущую схему:

```bash
make generate
```

Добавить другую сборку, доступную в официальном `Blizzard/heroprotocol`:

```bash
make generate PROTOCOL_BUILD=96477
```

Заголовок реплея содержит `m_baseBuild`, по которому выбирается YAML-схема. Если реплей новее доступных схем, используется последняя встроенная версия в режиме совместимости. В этом случае CLI выводит `"exact_protocol": false`: versioned-файлы могут читаться благодаря числовым тегам, но корректность bit-packed данных новой сборки не гарантируется.

## Структура Go-кода

```text
mpq/               MPQ-архив, хеши, шифрование и распаковка
protocol/          Типы и декодеры протокола Blizzard
replay/            Высокоуровневый API реплея
internal/formats/  Сгенерированный Kaitai Go-код
cmd/hotsreplay/    CLI
schemas.go         Встраивание YAML из репозитория схем
```

## Проверка

```bash
go test ./...
go build ./cmd/hotsreplay
```
