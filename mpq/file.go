package mpq

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func (a *Archive) ReadFile(name string) ([]byte, error) {
	block, err := a.Find(name)
	if err != nil {
		return nil, err
	}
	data, err := a.readAt(a.baseOffset+int64(block.Offset), uint64(block.CompressedSize))
	if err != nil {
		return nil, fmt.Errorf("mpq: read %q: %w", name, err)
	}
	if block.Flags&FileExists == 0 {
		return nil, fmt.Errorf("mpq: file %q is marked as absent", name)
	}

	key := uint32(0)
	if block.Flags&FileEncrypted != 0 {
		key = HashString(name, HashFileKey)
		if block.Flags&FileFixKey != 0 {
			key = (key + block.Offset) ^ block.UncompressedSize
		}
	}
	if block.Flags&FileSingleUnit != 0 {
		return readSingleUnit(data, block, key)
	}
	if block.Flags&(FileCompress|FileImplode) == 0 {
		if block.Flags&FileEncrypted != 0 {
			return nil, fmt.Errorf("mpq: encrypted unsectorised file %q is unsupported", name)
		}
		return data, nil
	}
	return a.readSectors(data, block, key, name)
}

func readSingleUnit(data []byte, block BlockEntry, key uint32) ([]byte, error) {
	if key != 0 {
		data = decrypt(data, key)
	}
	if block.Flags&(FileCompress|FileImplode) != 0 && block.CompressedSize < block.UncompressedSize {
		if block.Flags&FileCompress == 0 {
			return decompressImplode(data, block.UncompressedSize)
		}
		return Decompress(data, block.UncompressedSize)
	}
	if uint32(len(data)) != block.UncompressedSize {
		return nil, fmt.Errorf("mpq: file size %d, expected %d", len(data), block.UncompressedSize)
	}
	return data, nil
}

func (a *Archive) readSectors(data []byte, block BlockEntry, key uint32, name string) ([]byte, error) {
	sectorSize := a.Header.SectorSize
	sectorCount := (block.UncompressedSize + sectorSize - 1) / sectorSize
	tableEntries := sectorCount + 1
	if block.Flags&FileSectorCRC != 0 {
		tableEntries++
	}
	tableSize := tableEntries * 4
	if tableSize > uint32(len(data)) {
		return nil, fmt.Errorf("mpq: truncated sector table in %q", name)
	}
	tableData := data[:tableSize]
	if key != 0 {
		tableData = decrypt(tableData, key-1)
	}
	offsets := make([]uint32, tableEntries)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(tableData[i*4:])
	}

	result := make([]byte, 0, block.UncompressedSize)
	for sector := uint32(0); sector < sectorCount; sector++ {
		start, end := offsets[sector], offsets[sector+1]
		if start > end || end > uint32(len(data)) {
			return nil, fmt.Errorf("mpq: invalid sector %d bounds in %q", sector, name)
		}
		chunk := data[start:end]
		if key != 0 {
			chunk = decrypt(chunk, key+sector)
		}
		expectedSize := sectorSize
		if sector == sectorCount-1 {
			expectedSize = block.UncompressedSize - sectorSize*sector
		}
		if block.Flags&(FileCompress|FileImplode) != 0 && uint32(len(chunk)) < expectedSize {
			var err error
			if block.Flags&FileCompress == 0 {
				chunk, err = decompressImplode(chunk, expectedSize)
			} else {
				chunk, err = Decompress(chunk, expectedSize)
			}
			if err != nil {
				return nil, fmt.Errorf("mpq: sector %d of %q: %w", sector, name, err)
			}
		}
		if uint32(len(chunk)) != expectedSize {
			return nil, fmt.Errorf("mpq: sector %d of %q has size %d, expected %d", sector, name, len(chunk), expectedSize)
		}
		result = append(result, chunk...)
	}
	return result, nil
}

func (a *Archive) List() ([]string, error) {
	data, err := a.ReadFile("(listfile)")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	result := lines[:0]
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}
