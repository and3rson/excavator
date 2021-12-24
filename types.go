package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type VarInt int32
type VarLong int64
type String string
type ByteArray []byte
type Packet struct {
	Length VarInt
	PacketID VarInt
	Data *bytes.Buffer
}

func NewPacket(packetID VarInt) *Packet {
	return &Packet{
		PacketID: packetID,
		Data: bytes.NewBuffer([]byte{}),
	}
}

func (p *Packet) String() string {
	data := make([]byte, p.Data.Len())
	_, _ = p.Data.Read(data)
	p.Data.Write(data)
	dataRepr := []string{}
	for i, b := range data {
		dataRepr = append(dataRepr, fmt.Sprintf("%02x", b))
		if i > 50 {
			dataRepr = append(dataRepr, fmt.Sprintf("... (+%d)", len(data) - i))
			break
		}
	}
	// io.R p.Data.Peek()
	return fmt.Sprintf(
		"{Packet length=%d id=%02x data=%s}",
		p.Length, p.PacketID, strings.Join(dataRepr, " "),
	)
}

func (v VarInt) ToBytes() []byte {
	uv := uint32(v)
	bytes := []byte{}
	for {
		if uv & ^uint32(0x7F) == 0 {
			bytes = append(bytes, byte(uv))
			break
		}
		bytes = append(bytes, byte((uv & 0x7F) | 0x80))
		uv = uv >> 7
	}
	return bytes
}

func (v *VarInt) FromReader(r io.Reader) error {
	*v = 0
	buf := []byte{0}
	for i := 0; i < 5; i++ {
		_, err := r.Read(buf) // TODO
		if err != nil {
			return err
		}
		*v |= VarInt(buf[0] & 0x7F) << (i * 7)
		if buf[0] & 0x80 == 0 {
			return nil
		}
	}
	panic("packetstream: VarInt too long")
}

func (v VarLong) ToBytes() []byte {
	uv := uint64(v)
	bytes := []byte{}
	for {
		if uv & ^uint64(0x7F) == 0 {
			bytes = append(bytes, byte(uv))
			break
		}
		bytes = append(bytes, byte((uv & 0x7F) | 0x80))
		uv = uv >> 7
	}
	return bytes
}

func (v *VarLong) FromReader(r io.Reader) error {
	*v = 0
	buf := []byte{0}
	for i := 0; i < 10; i++ {
		_, err := r.Read(buf) // TODO
		if err != nil {
			return err
		}
		*v |= VarLong(buf[0] & 0x7F) << (i * 7)
		if buf[0] & 0x80 == 0 {
			return nil
		}
	}
	panic("packetstream: VarInt too long")
}

func (s String) ToBytes() []byte {
	return append(VarInt(len(s)).ToBytes(), []byte(s)...)
}

func (s *String) FromReader(r io.Reader) error {
	length := []byte{0}
	if _, err := r.Read(length); err != nil {
		return err
	}
	data := make([]byte, length[0])
	if _, err := r.Read(data); err != nil {
		return err
	}
	*s = String(data)
	return nil
}

func (b ByteArray) ToBytes() []byte {
	return []byte(b)
}

func (b ByteArray) ToBytesWithSize() []byte {
	return append(VarInt(len(b)).ToBytes(), b.ToBytes()...)
}

func (b *ByteArray) FromReader(r io.Reader, length int32) error {
	buf := make([]byte, length)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	*b = buf
	return nil
}

func (b *ByteArray) FromReaderWithSize(r io.Reader) error {
	var length VarInt
	if err := length.FromReader(r); err != nil {
		return err
	}
	if err := b.FromReader(r, int32(length)); err != nil {
		return err
	}
	return nil
}
