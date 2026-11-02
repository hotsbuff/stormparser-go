package mpq

import "errors"

const (
	UserDataMagic = "MPQ\x1b"
	ArchiveMagic  = "MPQ\x1a"
)

type HashType uint32

const (
	HashTableOffset HashType = 0
	HashNameA       HashType = 1
	HashNameB       HashType = 2
	HashFileKey     HashType = 3
)

type FileFlags uint32

const (
	FileImplode      FileFlags = 0x00000100
	FileCompress     FileFlags = 0x00000200
	FileEncrypted    FileFlags = 0x00010000
	FileFixKey       FileFlags = 0x00020000
	FilePatch        FileFlags = 0x00100000
	FileSingleUnit   FileFlags = 0x01000000
	FileDeleteMarker FileFlags = 0x02000000
	FileSectorCRC    FileFlags = 0x04000000
	FileExists       FileFlags = 0x80000000
)

type Compression uint8

const (
	CompressionHuffman   Compression = 0x01
	CompressionZlib      Compression = 0x02
	CompressionPKWARE    Compression = 0x08
	CompressionBzip2     Compression = 0x10
	CompressionLZMA      Compression = 0x12
	CompressionSparse    Compression = 0x20
	CompressionADPCMMono Compression = 0x40
	CompressionADPCM     Compression = 0x80
)

type UserDataHeader struct {
	UserDataSize        uint32
	ArchiveHeaderOffset uint32
	Content             []byte
}

type Header struct {
	FormatVersion     uint16
	SectorSize        uint32
	HashTableOffset   uint64
	BlockTableOffset  uint64
	HashTableEntries  uint32
	BlockTableEntries uint32
}

type HashEntry struct {
	NameA      uint32
	NameB      uint32
	Locale     uint16
	Platform   uint16
	BlockIndex uint32
}

type BlockEntry struct {
	Offset           uint32
	CompressedSize   uint32
	UncompressedSize uint32
	Flags            FileFlags
}

var ErrFileNotFound = errors.New("mpq: file not found")
