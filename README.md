# Heroes of the Storm MPQ Parser

[Русская версия](README.ru.md)

A Go library and CLI for reading `.StormReplay` files. The project parses the MPQ archive and Blizzard's internal binary protocols without being tied to a single game build.

Format definitions live in a separate repository: [hotsbuff/stormprotocol](https://github.com/hotsbuff/stormprotocol). They are available under `schemas/` in this project and are used for code generation and YAML embedding.

## Features

- MPQ user data and archive headers;
- hash and block tables, including file lookup by name;
- table and file decryption;
- uncompressed, sector-based, and single-unit files;
- zlib, bzip2, LZMA, PKWARE, Huffman, Sparse, and mono/stereo IMA ADPCM decompression;
- typed match summaries, player details, Toon handles, colors, and timestamps;
- Blizzard versioned binary and bit-packed formats;
- `replay.header`, `replay.details`, and `replay.initData`;
- game events, messages, tracker events, and attributes;
- preservation of unknown versioned fields;
- portable Kaitai and YAML schemas.

HotS replay streams do not use protobuf: game events and messages use bit-packed encoding, while tracker events use versioned binary encoding.

## Cloning

The schema repository is connected as a Git submodule. Clone the project with submodules:

```bash
git clone --recurse-submodules https://github.com/hotsbuff/stormparser-go.git
cd stormparser-go
```

If the project was cloned without submodules:

```bash
git submodule update --init --recursive
```

To update `schemas/` to the latest commit from `stormprotocol`:

```bash
git submodule update --remote schemas
```

After updating, commit the changed submodule reference in the main repository.

## Requirements

Using the library and CLI requires the Go version specified in `go.mod`.

Additional tools are needed only for regeneration:

- [Kaitai Struct Compiler](https://kaitai.io/) to compile `.ksy` files into Go;
- Python 3 to convert official `protocol<build>.py` files into YAML.

Python is not required at runtime.

## CLI

Print the replay version and selected protocol schema:

```bash
go run ./cmd/hotsreplay path/to/replay.StormReplay
```

Available options:

```text
-summary         Typed match summary with players and duration
-header          Complete replay.header
-details         replay.details
-initdata        replay.initData
-attributes      replay.attributes.events
-gameevents      replay.game.events
-messageevents   replay.message.events
-trackerevents   replay.tracker.events
-extract NAME    Extract an MPQ file to stdout
```

Decoding options can be combined. Output is JSON:

```bash
go run ./cmd/hotsreplay -details -attributes path/to/replay.StormReplay
```

Extract an original binary file:

```bash
go run ./cmd/hotsreplay -extract replay.details path/to/replay.StormReplay > replay.details
```

## Go API

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

For low-level archive access, import `github.com/hotsbuff/stormparser-go/mpq`:

```go
archive, err := mpq.Open("match.StormReplay")
if err != nil {
    // handle error
}
defer archive.Close()

data, err := archive.ReadFile("replay.details")
```

## Protocol Schemas

The `schemas/` directory contains:

```text
schemas/
├── formats/          # Kaitai binary format schemas
├── protocol/builds/  # YAML game build definitions
└── tools/            # YAML generator for official heroprotocol files
```

Regenerate the Go code and current schema:

```bash
make generate
```

Add another build available in the official `Blizzard/heroprotocol` repository:

```bash
make generate PROTOCOL_BUILD=96477
```

The replay header contains `m_baseBuild`, which selects the YAML schema. If the replay is newer than the available schemas, the latest embedded schema is used in compatibility mode. In this case the CLI prints `"exact_protocol": false`: versioned files may remain readable through numeric tags, but bit-packed decoding correctness is not guaranteed for the newer build.

## Go Code Layout

```text
mpq/               MPQ archive, hashing, decryption, and decompression
protocol/          Blizzard protocol types and decoders
replay/            High-level replay API
internal/formats/  Generated Kaitai Go code
cmd/hotsreplay/    CLI
schemas.go         YAML embedding from the schema repository
```

## Verification

```bash
go test ./...
go build ./cmd/hotsreplay
```
