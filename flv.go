package flv

import (
	"encoding/binary"
	"errors"
	"io"
)

type HeaderFlag byte

func IsHeaderFlag(c byte) bool {
	return c == byte(HeaderFlagAudio) ||
		c == byte(HeaderFlagVideo) ||
		c == byte(HeaderFlagAudioVideo)
}

const (
	HeaderFlagAudio      HeaderFlag = 0x04
	HeaderFlagVideo      HeaderFlag = 0x01
	HeaderFlagAudioVideo HeaderFlag = HeaderFlagAudio | HeaderFlagVideo
)

type TagFlag byte

func IsTagFlag(c byte) bool {
	return c == byte(TagFlagAudio) ||
		c == byte(TagFlagVideo) ||
		c == byte(TagFlagScript)
}

const (
	TagFlagAudio  TagFlag = 0x08
	TagFlagVideo  TagFlag = 0x09
	TagFlagScript TagFlag = 0x12
)

var (
	errInvalidHeaderFormat = errors.New("invalid header format")
	errInvalidTagFormat    = errors.New("invalid tag format")
)

func putUint24(b []byte, m uint32) {
	b[0] = byte(m >> 16)
	b[1] = byte(m >> 8)
	b[2] = byte(m)
}

func uint24(b []byte) uint32 {
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

type Header struct {
	buff    [13]byte
	Version byte
	Flag    HeaderFlag
}

func (h *Header) WriteTo(writer io.Writer) (int64, error) {
	h.buff[0] = 'F'
	h.buff[1] = 'L'
	h.buff[2] = 'V'
	h.buff[3] = h.Version
	h.buff[4] = byte(h.Flag)
	binary.BigEndian.PutUint32(h.buff[5:], 9)
	binary.BigEndian.PutUint32(h.buff[9:], 0)
	n, err := writer.Write(h.buff[:])
	return int64(n), err
}

func (h *Header) ReadFrom(reader io.Reader) (int64, error) {
	n, err := io.ReadFull(reader, h.buff[:])
	if err != nil {
		return int64(n), err
	}
	if h.buff[0] != 'F' && h.buff[1] != 'L' && h.buff[2] != 'V' && !IsHeaderFlag(h.buff[4]) {
		return int64(n), errInvalidHeaderFormat
	}
	h.Version = h.buff[3]
	h.Flag = HeaderFlag(h.buff[4])
	binary.BigEndian.Uint32(h.buff[5:])
	// first previous size
	binary.BigEndian.Uint32(h.buff[9:])
	return int64(n), err
}

type Tag struct {
	buff      [15]byte
	Flag      TagFlag
	Timestamp uint32
	StreamID  uint32
	Data      []byte
}

func (t *Tag) WriteTo(writer io.Writer) (int64, error) {
	// flag
	t.buff[0] = byte(t.Flag)
	// data size
	putUint24(t.buff[1:], uint32(len(t.Data)))
	// timestamp
	binary.BigEndian.PutUint32(t.buff[4:], t.Timestamp)
	// stream id
	putUint24(t.buff[8:], t.StreamID)
	// data
	binary.BigEndian.PutUint32(t.buff[11:], uint32(len(t.Data)))
	// write header
	n, err := writer.Write(t.buff[:])
	if err != nil {
		return int64(n), err
	}
	// write data
	n, err = writer.Write(t.Data)
	return int64(n + len(t.buff)), err
}

func (t *Tag) ReadFrom(reader io.Reader) (int64, error) {
	n, err := io.ReadFull(reader, t.buff[:11])
	if err != nil {
		return int64(n), err
	}
	// flag
	if !IsTagFlag(t.buff[0]) {
		return int64(n), errInvalidTagFormat
	}
	t.Flag = TagFlag(t.buff[0])
	// data size
	dataSize := int(uint24(t.buff[1:]))
	// timestamp
	t.Timestamp = binary.BigEndian.Uint32(t.buff[4:])
	// stream id
	t.StreamID = uint24(t.buff[8:])
	// data
	if cap(t.Data) < dataSize {
		t.Data = make([]byte, dataSize)
	} else {
		t.Data = t.Data[:dataSize]
	}
	n, err = io.ReadFull(reader, t.Data)
	if err != nil {
		return int64(n + 11), err
	}
	// first previous size
	_, err = io.ReadFull(reader, t.buff[:4])
	return int64(n + 15), err
}
