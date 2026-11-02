package mpq

import (
	"bytes"
	"compress/bzip2"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hotsbuff/stormparser-go/internal/compress/adpcm"
	"github.com/hotsbuff/stormparser-go/internal/compress/huffman"
	"github.com/hotsbuff/stormparser-go/internal/compress/pkware"
	"github.com/ulikunitz/xz/lzma"
)

func Decompress(data []byte, uncompressedSize uint32) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("mpq: empty compressed block")
	}
	mask := Compression(data[0])
	if mask == CompressionLZMA {
		return decompressLZMA(data[1:], uncompressedSize)
	}
	known := CompressionHuffman | CompressionZlib | CompressionPKWARE | CompressionBzip2 | CompressionSparse | CompressionADPCMMono | CompressionADPCM
	if unknown := mask &^ known; unknown != 0 {
		return nil, fmt.Errorf("mpq: unsupported compression mask 0x%02x", uint8(unknown))
	}

	var err error
	result := data[1:]
	intermediateSize := int(uncompressedSize)
	if mask&(CompressionADPCM|CompressionADPCMMono) != 0 {
		intermediateSize = int(uncompressedSize)/2 + 6
	}
	if mask&CompressionBzip2 != 0 {
		result, err = readStream(bzip2.NewReader(bytes.NewReader(result)))
		if err != nil {
			return nil, fmt.Errorf("mpq: bzip2: %w", err)
		}
	}
	if mask&CompressionPKWARE != 0 {
		result, err = pkware.Explode(result, intermediateSize)
		if err != nil {
			return nil, fmt.Errorf("mpq: pkware: %w", err)
		}
	}
	if mask&CompressionZlib != 0 {
		result, err = decompressZlib(result)
		if err != nil {
			return nil, err
		}
	}
	if mask&CompressionHuffman != 0 {
		result, err = huffman.Decompress(result, intermediateSize)
		if err != nil {
			return nil, fmt.Errorf("mpq: huffman: %w", err)
		}
	}
	if mask&CompressionADPCM != 0 {
		result, err = adpcm.Decompress(result, 2)
	} else if mask&CompressionADPCMMono != 0 {
		result, err = adpcm.Decompress(result, 1)
	}
	if err != nil {
		return nil, fmt.Errorf("mpq: adpcm: %w", err)
	}
	if mask&CompressionSparse != 0 {
		result, err = decompressSparse(result, uncompressedSize)
		if err != nil {
			return nil, err
		}
	}
	return checkDecompressedSize(result, uncompressedSize)
}

func decompressImplode(data []byte, uncompressedSize uint32) ([]byte, error) {
	result, err := pkware.Explode(data, int(uncompressedSize))
	if err != nil {
		return nil, fmt.Errorf("mpq: implode: %w", err)
	}
	return checkDecompressedSize(result, uncompressedSize)
}

func decompressZlib(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("mpq: open zlib stream: %w", err)
	}
	defer reader.Close()
	result, err := readStream(reader)
	if err != nil {
		return nil, fmt.Errorf("mpq: zlib: %w", err)
	}
	return result, nil
}

func decompressLZMA(data []byte, uncompressedSize uint32) ([]byte, error) {
	if len(data) < 2 || data[0] != 0 {
		return nil, fmt.Errorf("mpq: invalid lzma filter header")
	}
	reader, err := lzma.NewReader(bytes.NewReader(data[1:]))
	if err != nil {
		return nil, fmt.Errorf("mpq: open lzma stream: %w", err)
	}
	result, err := readStream(reader)
	if err != nil {
		return nil, fmt.Errorf("mpq: lzma: %w", err)
	}
	return checkDecompressedSize(result, uncompressedSize)
}

func decompressSparse(data []byte, expectedSize uint32) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("mpq: sparse payload is truncated")
	}
	outputSize := binary.BigEndian.Uint32(data[:4])
	if outputSize > expectedSize {
		return nil, fmt.Errorf("mpq: sparse size %d exceeds expected %d", outputSize, expectedSize)
	}
	result := make([]byte, outputSize)
	inputPosition := 4
	outputPosition := uint32(0)
	for inputPosition < len(data) && outputPosition < outputSize {
		tag := data[inputPosition]
		inputPosition++
		if tag&0x80 != 0 {
			count := uint32(tag&0x7f) + 1
			if count > uint32(len(data)-inputPosition) || count > outputSize-outputPosition {
				return nil, fmt.Errorf("mpq: sparse literal exceeds block bounds")
			}
			copy(result[outputPosition:outputPosition+count], data[inputPosition:inputPosition+int(count)])
			inputPosition += int(count)
			outputPosition += count
			continue
		}
		count := uint32(tag&0x7f) + 3
		if count > outputSize-outputPosition {
			return nil, fmt.Errorf("mpq: sparse zero run exceeds block bounds")
		}
		outputPosition += count
	}
	if outputPosition != outputSize {
		return nil, fmt.Errorf("mpq: sparse output is truncated")
	}
	return result, nil
}

func readStream(reader io.Reader) ([]byte, error) {
	return io.ReadAll(reader)
}

func checkDecompressedSize(data []byte, expectedSize uint32) ([]byte, error) {
	if uint32(len(data)) != expectedSize {
		return nil, fmt.Errorf("mpq: decompressed size %d, expected %d", len(data), expectedSize)
	}
	return data, nil
}
