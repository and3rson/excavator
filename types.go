package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

//go:generate go run tools/genpackets.go

type Direction int

const (
	ClientBound Direction = iota
	ServerBound
)

type VarInt int32
type VarLong int64
type UInt16 uint16
type Long int64
type String string
type ByteArray []byte
type Chat struct {
	Text  string `json:"text"`
	Extra []struct {
		Text string `json:"text"`
	} `json:"extra"`
}
type Packet struct {
	PacketID VarInt
	Data     *bytes.Buffer
	State    GameState
	Direction
}

type Serializable interface {
	ToBytes() []byte
}

func NewPacket(packetID VarInt, args ...Serializable) *Packet {
	p := &Packet{
		PacketID: packetID,
		Data:     bytes.NewBuffer([]byte{}),
	}
	for _, arg := range args {
		p.Data.Write(arg.ToBytes())
	}
	return p
}

func (p *Packet) String() string {
	data := make([]byte, p.Data.Len())
	_, _ = p.Data.Read(data)
	p.Data.Write(data)
	dataRepr := []string{}
	for i, b := range data {
		dataRepr = append(dataRepr, fmt.Sprintf("%02x", b))
		if i > 40 {
			dataRepr = append(dataRepr, fmt.Sprintf("... (+%d)", len(data)-i))
			break
		}
	}
	// io.R p.Data.Peek()
	name := fmt.Sprintf("0x%02x", p.PacketID)
	color := "160"

	if directionMap, ok := PacketNames[p.State]; ok {
		if packetIDMap, ok := directionMap[p.Direction]; ok {
			if packetName, ok := packetIDMap[int32(p.PacketID)]; ok {
				name = fmt.Sprintf("%s (0x%02x)", packetName, p.PacketID)
				color = "190"
			}
		}
	}
	// if foundName, ok := PacketNamesIn[int32(p.PacketID)]; p.In && ok {
	// 	name = foundName
	// 	color = "190"
	// }
	return fmt.Sprintf(
		"{Packet \u001b[38;5;%sm%s\u001b[0m data(%d)=\u001b[38;5;34m%s\u001b[0m}",
		color, name, p.Data.Len(), strings.Join(dataRepr, " "),
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
		bytes = append(bytes, byte((uv&0x7F)|0x80))
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
		*v |= VarInt(buf[0]&0x7F) << (i * 7)
		if buf[0]&0x80 == 0 {
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
		bytes = append(bytes, byte((uv&0x7F)|0x80))
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
		*v |= VarLong(buf[0]&0x7F) << (i * 7)
		if buf[0]&0x80 == 0 {
			return nil
		}
	}
	panic("packetstream: VarInt too long")
}

func (u *UInt16) FromReader(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, u)
}

func (u UInt16) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, u)
	return buf.Bytes()
}

func (l *Long) FromReader(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, l)
}

func (l Long) ToBytes() []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, l)
	return buf.Bytes()
}

func (s String) ToBytes() []byte {
	return append(VarInt(len(s)).ToBytes(), []byte(s)...)
}

func (s *String) FromReader(r io.Reader) error {
	var length VarInt
	if err := length.FromReader(r); err != nil {
		return err
	}
	data := make([]byte, length)
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

func (c *Chat) FromReader(r io.Reader) error {
	var jsonLength VarInt
	if err := jsonLength.FromReader(r); err != nil {
		return err
	}
	jsonData := make([]byte, jsonLength)
	if _, err := r.Read(jsonData); err != nil {
		return err
	}
	if err := json.Unmarshal(jsonData, c); err != nil {
		return err
	}
	return nil
}

func (c *Chat) cleanUp(s string) string {
	s = regexp.MustCompile("\\xA7[\\d\\w]").ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

func (c *Chat) String() string {
	parts := []string{c.cleanUp(c.Text)}
	for _, extra := range c.Extra {
		parts = append(parts, c.cleanUp(extra.Text))
	}
	return strings.Join(parts, "")
}
