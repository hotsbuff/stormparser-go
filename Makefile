PROTOCOL_BUILD ?= 96477

.PHONY: generate test build

generate:
	kaitai-struct-compiler -t go --go-package formats --outdir internal schemas/formats/mpq.ksy schemas/formats/storm_value.ksy schemas/formats/attributes_events.ksy
	python3 schemas/tools/export_protocol.py https://raw.githubusercontent.com/Blizzard/heroprotocol/master/heroprotocol/versions/protocol$(PROTOCOL_BUILD).py schemas/protocol/builds/$(PROTOCOL_BUILD).yaml

test:
	go test ./...

build:
	go build ./cmd/hotsreplay
