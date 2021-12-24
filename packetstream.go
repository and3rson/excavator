package main

import (
	"bytes"
	"io/ioutil"
	// "compress/flate"
	"compress/zlib"
	"crypto/cipher"
	"fmt"
	"io"

	// "io/ioutil"

	log "github.com/sirupsen/logrus"
)

type PacketStream struct {
	*PacketStreamReader
	*PacketStreamWriter
}

type PacketStreamReader struct {
	io.Reader
	compressionThreshold int32
}

type PacketStreamWriter struct {
	io.WriteCloser
	compressionThreshold int32
}

func NewPacketStream(stream io.ReadWriteCloser) *PacketStream {
	return &PacketStream{
		NewPacketStreamReader(stream),
		NewPacketStreamWriter(stream),
	}
}

func (s *PacketStream) SetCompressionThreshold(compressionThreshold int32) {
	s.PacketStreamReader.compressionThreshold = compressionThreshold
	s.PacketStreamWriter.compressionThreshold = compressionThreshold
}

func NewPacketStreamWithEncryption(
	stream io.ReadWriteCloser, encrypter cipher.Stream, decrypter cipher.Stream,
) *PacketStream {
	return &PacketStream{
		&PacketStreamReader{
			cipher.StreamReader{
				S: decrypter,
				R: stream,
			},
			-1,
		},
		&PacketStreamWriter{
			cipher.StreamWriter{
				S: encrypter,
				W: stream,
			},
			-1,
		},
	}
}

func NewPacketStreamReader(stream io.ReadCloser) *PacketStreamReader {
	return &PacketStreamReader{stream, -1}
}

func NewPacketStreamWriter(stream io.WriteCloser) *PacketStreamWriter {
	return &PacketStreamWriter{stream, -1}
}

func (r *PacketStreamReader) ReadPacket() (*Packet, error) {
	length := VarInt(0)
	packetID := VarInt(0)
	var packetData []byte
	var p *Packet
	var err error
	var reader io.Reader = r
	var bodyLength VarInt
	compressed := false
	// data := bytes.NewBuffer([]byte{})
	if err := length.FromReader(r); err != nil {
		return nil, fmt.Errorf("packetstream: read length: %s", err)
	}
	log.Tracef("Incoming length: %d", length)
	if r.compressionThreshold != -1 {
		// Compression enabled
		uncompressedLength := VarInt(0)
		if err = uncompressedLength.FromReader(r); err != nil {
			return nil, fmt.Errorf("packetstream: read body length: %s", err)
		}
		if uncompressedLength != 0 {
			log.Tracef("Uncompressed length: %d", uncompressedLength)
			compressed = true
			if reader, err = zlib.NewReader(r); err != nil {
				return nil, fmt.Errorf("packetstream: init zlib reader: %s", err)
			}
		}
		bodyLength = length - VarInt(len(uncompressedLength.ToBytes()))
		// bodyLength = length
	} else {
		// Compression disabled
		bodyLength = length
	}

	if err = packetID.FromReader(reader); err != nil {
		return nil, fmt.Errorf("packetstream: read id: %s", err)
	}
	// fmt.Println("len", length, "packet id", packetID)
	packetData = make([]byte, int(bodyLength) - len(packetID.ToBytes()))
	if _, err = io.ReadFull(reader, packetData); err != nil {
		return nil, fmt.Errorf("packetstream: read body: %s", err)
	}
	p = &Packet{
		Length:   bodyLength,
		PacketID: packetID,
		Data:     bytes.NewBuffer(packetData),
	}
	flags := ""
	if r.compressionThreshold != -1 {
		if compressed {
			flags = "ZZ"
		} else {
			flags = "Z"
		}
	}
	log.Tracef("<<%s %s", flags, p)
	return p, nil
}

func (w *PacketStreamWriter) WritePacket(p *Packet) error {
	packetIDBytes := p.PacketID.ToBytes()
	p.Length = VarInt(len(packetIDBytes) + p.Data.Len())
	var packetBody = append(p.PacketID.ToBytes(), p.Data.Bytes()...)
	// var writer io.Writer = w
	compressed := false
	if w.compressionThreshold >= 0 {
		// Compression enabled
		uncompressedLength := VarInt(0)
		if len(packetBody) >= int(w.compressionThreshold) {
			// Compress
			var err error
			compressedBuffer := bytes.NewBuffer([]byte{})
			compressor := zlib.NewWriter(compressedBuffer)
			if _, err = compressor.Write(packetBody); err != nil {
				return err
			}
			if packetBody, err = ioutil.ReadAll(compressedBuffer); err != nil {
				return err
			}
			compressed = true
			uncompressedLength = p.Length
		}
		if _, err := w.Write(VarInt(len(uncompressedLength.ToBytes()) + len(packetBody)).ToBytes()); err != nil {
			return fmt.Errorf("packetstream: write length (1)")
		}
		if _, err := w.Write(uncompressedLength.ToBytes()); err != nil {
			return fmt.Errorf("packetstream: write length (2)")
		}
		if _, err := w.Write(packetBody); err != nil {
			return fmt.Errorf("packetstream: write body")
		}
	} else {
		// Compression disabled
		if _, err := w.Write(append(p.Length.ToBytes(), packetBody...)); err != nil {
			return fmt.Errorf("packetstream: write packet: %s", err)
		}
		// if _, err := w.Write(append(p.Length.ToBytes(), packetIDBytes...)); err != nil {
		// 	return fmt.Errorf("packetstream: write length: %s", err)
		// }
		// if _, err := io.Copy(w, p.Data); err != nil {
		// 	return fmt.Errorf("packetstream: write body: %s", err)
		// }
	}
	flags := ""
	if w.compressionThreshold != -1 {
		if compressed {
			flags = "ZZ"
		} else {
			flags = "Z"
		}
	}

	log.Tracef(">>%s %s", flags, p)
	return nil
}

// func (s *PacketStreamReader) Close() error {
// 	return s.Close()
// }
