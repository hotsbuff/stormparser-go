package replay

import (
	"fmt"

	"github.com/hotsbuff/stormparser-go/mpq"
	"github.com/hotsbuff/stormparser-go/protocol"
)

type Replay struct {
	Archive *mpq.Archive
	Schema  *protocol.Schema
	Header  *protocol.Value
}

func Open(path string) (*Replay, error) {
	archive, err := mpq.Open(path)
	if err != nil {
		return nil, err
	}
	replay, err := New(archive)
	if err != nil {
		archive.Close()
		return nil, err
	}
	return replay, nil
}

func New(archive *mpq.Archive) (*Replay, error) {
	latest, err := protocol.Latest()
	if err != nil {
		return nil, err
	}
	header, err := protocol.DecodeVersioned(archive.UserDataHeader.Content, latest, latest.HeaderType)
	if err != nil {
		return nil, fmt.Errorf("decode replay header: %w", err)
	}
	version := header.Field("m_version")
	if version == nil || version.Field("m_baseBuild") == nil {
		return nil, fmt.Errorf("replay header has no m_version.m_baseBuild")
	}
	baseBuild := uint32(version.Field("m_baseBuild").Int)
	schema, _, err := protocol.LoadCompatible(baseBuild)
	if err != nil {
		return nil, err
	}
	return &Replay{Archive: archive, Schema: schema, Header: header}, nil
}

func (r *Replay) Close() error {
	return r.Archive.Close()
}

func (r *Replay) Version() Version {
	version := r.Header.Field("m_version")
	return Version{
		Flags:     version.Field("m_flags").Int,
		Major:     version.Field("m_major").Int,
		Minor:     version.Field("m_minor").Int,
		Revision:  version.Field("m_revision").Int,
		Build:     version.Field("m_build").Int,
		BaseBuild: version.Field("m_baseBuild").Int,
	}
}

func (r *Replay) ReadFile(name string) ([]byte, error) {
	return r.Archive.ReadFile(name)
}

func (r *Replay) Details() (*protocol.Value, error) {
	data, err := r.ReadFile(DetailsFile)
	if err != nil {
		return nil, err
	}
	return protocol.DecodeVersioned(data, r.Schema, r.Schema.DetailsType)
}

func (r *Replay) InitData() (*protocol.Value, error) {
	data, err := r.ReadFile(InitDataFile)
	if err != nil {
		return nil, err
	}
	return protocol.DecodeBitPacked(data, r.Schema, r.Schema.InitDataType)
}

func (r *Replay) Attributes() (*protocol.Attributes, error) {
	data, err := r.ReadFile(AttributesFile)
	if err != nil {
		return nil, err
	}
	return protocol.DecodeAttributes(data)
}
