package mpq

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/hotsbuff/stormparser-go/internal/formats"
	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
)

type Archive struct {
	reader         io.ReaderAt
	size           int64
	closer         io.Closer
	baseOffset     int64
	UserDataHeader UserDataHeader
	Header         Header
	hashTable      []HashEntry
	blockTable     []BlockEntry
}

func Open(path string) (*Archive, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	archive, err := New(file, info.Size())
	if err != nil {
		file.Close()
		return nil, err
	}
	archive.closer = file
	return archive, nil
}

func New(reader io.ReaderAt, size int64) (*Archive, error) {
	stream := kaitai.NewStream(io.NewSectionReader(reader, 0, size))
	decoded := formats.NewMpq()
	if err := decoded.Read(stream, nil, decoded); err != nil {
		return nil, fmt.Errorf("mpq: read user data header: %w", err)
	}
	archiveHeader, err := decoded.ArchiveHeader()
	if err != nil {
		return nil, fmt.Errorf("mpq: read archive header: %w", err)
	}
	archive := &Archive{
		reader:     reader,
		size:       size,
		baseOffset: int64(decoded.UserDataHeader.ArchiveHeaderOffset),
		UserDataHeader: UserDataHeader{
			UserDataSize:        decoded.UserDataHeader.UserDataSize,
			ArchiveHeaderOffset: decoded.UserDataHeader.ArchiveHeaderOffset,
			Content:             decoded.UserDataHeader.Content,
		},
		Header: Header{
			FormatVersion:     archiveHeader.FormatVersion,
			SectorSize:        uint32(512) << archiveHeader.SectorSizeShift,
			HashTableOffset:   uint64(archiveHeader.HashTableOffset) | uint64(archiveHeader.HashTableOffsetHigh)<<32,
			BlockTableOffset:  uint64(archiveHeader.BlockTableOffset) | uint64(archiveHeader.BlockTableOffsetHigh)<<32,
			HashTableEntries:  archiveHeader.HashTableEntries,
			BlockTableEntries: archiveHeader.BlockTableEntries,
		},
	}
	if err := archive.readTables(); err != nil {
		return nil, err
	}
	return archive, nil
}

func (a *Archive) Close() error {
	if a.closer != nil {
		return a.closer.Close()
	}
	return nil
}

func (a *Archive) readTables() error {
	hashData, err := a.readAt(a.baseOffset+int64(a.Header.HashTableOffset), uint64(a.Header.HashTableEntries)*16)
	if err != nil {
		return fmt.Errorf("mpq: read hash table: %w", err)
	}
	hashData = decrypt(hashData, HashString("(hash table)", HashFileKey))
	a.hashTable = make([]HashEntry, a.Header.HashTableEntries)
	for i := range a.hashTable {
		offset := i * 16
		a.hashTable[i] = HashEntry{
			NameA:      binary.LittleEndian.Uint32(hashData[offset:]),
			NameB:      binary.LittleEndian.Uint32(hashData[offset+4:]),
			Locale:     binary.LittleEndian.Uint16(hashData[offset+8:]),
			Platform:   binary.LittleEndian.Uint16(hashData[offset+10:]),
			BlockIndex: binary.LittleEndian.Uint32(hashData[offset+12:]),
		}
	}

	blockData, err := a.readAt(a.baseOffset+int64(a.Header.BlockTableOffset), uint64(a.Header.BlockTableEntries)*16)
	if err != nil {
		return fmt.Errorf("mpq: read block table: %w", err)
	}
	blockData = decrypt(blockData, HashString("(block table)", HashFileKey))
	a.blockTable = make([]BlockEntry, a.Header.BlockTableEntries)
	for i := range a.blockTable {
		offset := i * 16
		a.blockTable[i] = BlockEntry{
			Offset:           binary.LittleEndian.Uint32(blockData[offset:]),
			CompressedSize:   binary.LittleEndian.Uint32(blockData[offset+4:]),
			UncompressedSize: binary.LittleEndian.Uint32(blockData[offset+8:]),
			Flags:            FileFlags(binary.LittleEndian.Uint32(blockData[offset+12:])),
		}
	}
	return nil
}

func (a *Archive) readAt(offset int64, size uint64) ([]byte, error) {
	if offset < 0 || size > uint64(a.size) || offset > a.size-int64(size) {
		return nil, fmt.Errorf("range [%d, %d) exceeds archive size %d", offset, uint64(offset)+size, a.size)
	}
	if size > uint64(^uint(0)>>1) {
		return nil, fmt.Errorf("allocation of %d bytes is too large", size)
	}
	data := make([]byte, int(size))
	_, err := io.ReadFull(io.NewSectionReader(a.reader, offset, int64(size)), data)
	return data, err
}

func (a *Archive) Find(name string) (BlockEntry, error) {
	if len(a.hashTable) == 0 {
		return BlockEntry{}, ErrFileNotFound
	}
	nameA := HashString(name, HashNameA)
	nameB := HashString(name, HashNameB)
	start := HashString(name, HashTableOffset) % uint32(len(a.hashTable))
	for i := uint32(0); i < uint32(len(a.hashTable)); i++ {
		entry := a.hashTable[(start+i)%uint32(len(a.hashTable))]
		if entry.BlockIndex == 0xffffffff {
			break
		}
		if entry.BlockIndex != 0xfffffffe && entry.NameA == nameA && entry.NameB == nameB {
			if uint64(entry.BlockIndex) >= uint64(len(a.blockTable)) {
				return BlockEntry{}, fmt.Errorf("mpq: invalid block index %d for %q", entry.BlockIndex, name)
			}
			return a.blockTable[entry.BlockIndex], nil
		}
	}
	return BlockEntry{}, fmt.Errorf("%w: %s", ErrFileNotFound, name)
}
