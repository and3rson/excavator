package main

import (
	"bytes"
	"fmt"
	"io"
)

type PacketStream struct {
	*PacketStreamReader
	*PacketStreamWriter
}

type PacketStreamReader struct {
	io.ReadCloser
}

type PacketStreamWriter struct {
	io.WriteCloser
}

func NewPacketStream(stream io.ReadWriteCloser) *PacketStream {
	return &PacketStream{
		&PacketStreamReader{stream},
		&PacketStreamWriter{stream},
	}
}

func NewPacketStreamReader(stream io.ReadCloser) *PacketStreamReader {
	return &PacketStreamReader{stream}
}

func NewPacketStreamWriter(stream io.WriteCloser) *PacketStreamWriter {
	return &PacketStreamWriter{stream}
}

func (r *PacketStreamReader) ReadPacket() (*Packet, error) {
	length := VarInt(0)
	packetID := VarInt(0)
	// data := bytes.NewBuffer([]byte{})
	if err := length.FromReader(r); err != nil {
		return nil, fmt.Errorf("packetstream: read length: %s", err)
	}
	if err := packetID.FromReader(r); err != nil {
		return nil, fmt.Errorf("packetstream: read id: %s", err)
	}
	fmt.Println("len", length, "packet id", packetID)
	dataBytes := make([]byte, int(length) - len(packetID.ToBytes()))
	if _, err := io.ReadFull(r, dataBytes); err != nil {
		return nil, fmt.Errorf("packetstream: read body: %s", err)
	}
	p := &Packet{
		Length:   length,
		PacketID: packetID,
		Data:     bytes.NewBuffer(dataBytes),
	}
	fmt.Println("<<", p)
	return p, nil
}

func (w *PacketStreamWriter) WritePacket(p *Packet) error {
	packetIDBytes := p.PacketID.ToBytes()
	p.Length = VarInt(len(packetIDBytes) + p.Data.Len())
	fmt.Println(">>", p)
	if _, err := w.Write(append(p.Length.ToBytes(), packetIDBytes...)); err != nil {
		return fmt.Errorf("packetstream: write header: %s", err)
	}
	if _, err := io.Copy(w, p.Data); err != nil {
		return fmt.Errorf("packetstream: write body: %s", err)
	}
	return nil
}

// func (s *PacketStreamReader) Close() error {
// 	return s.Close()
// }
