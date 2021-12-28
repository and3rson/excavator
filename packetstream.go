package main

import (
	"bytes"
	"compress/zlib"
	"crypto/cipher"
	"fmt"
	"io"
	"io/ioutil"
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

func (r *PacketStreamReader) ReadPacket() (VarInt, *bytes.Buffer, bool, error) {
	length := VarInt(0)
	packetID := VarInt(0)
	var packetData []byte
	var err error
	var reader io.Reader = r
	var bodyLength VarInt
	compressed := false
	// data := bytes.NewBuffer([]byte{})
	if err := length.FromReader(r); err != nil {
		return 0, nil, false, fmt.Errorf("packetstream: read length: %s", err)
	}
	// log.Tracef("packetstream: Incoming length: %d", length)
	if r.compressionThreshold != -1 {
		// Compression enabled
		uncompressedLength := VarInt(0)
		if err = uncompressedLength.FromReader(r); err != nil {
			return 0, nil, false, fmt.Errorf("packetstream: read body length: %s", err)
		}
		bodyLength = length - VarInt(len(uncompressedLength.ToBytes()))
		if uncompressedLength != 0 {
			// log.Tracef("packetstream: Uncompressed length: %d", uncompressedLength)
			compressed = true
			// Pre-read data into buffer so that zlib does not mess up our stream
			buf := bytes.NewBuffer([]byte{})
			if _, err = io.CopyN(buf, r, int64(bodyLength)); err != nil {
				return 0, nil, false, fmt.Errorf("packetstream: pre-read compressed data: %s", err)
			}
			if reader, err = zlib.NewReader(buf); err != nil {
				return 0, nil, false, fmt.Errorf("packetstream: init zlib reader: %s", err)
			}
		}
		// bodyLength = length
	} else {
		// Compression disabled
		bodyLength = length
	}

	// log.Infof("will read %d bytes", bodyLength)

	if err = packetID.FromReader(reader); err != nil {
		return 0, nil, false, fmt.Errorf("packetstream: read id: %s", err)
	}
	// fmt.Println("len", length, "packet id", packetID)
	packetData = make([]byte, int(bodyLength)-len(packetID.ToBytes()))
	if _, err = io.ReadFull(reader, packetData); err != nil {
		return 0, nil, false, fmt.Errorf("packetstream: read body: %s", err)
	}

	return packetID, bytes.NewBuffer(packetData), compressed, nil
}

func (w *PacketStreamWriter) WritePacket(packetID VarInt, data *bytes.Buffer) (bool, error) {
	packetIDBytes := packetID.ToBytes()
	length := VarInt(len(packetIDBytes) + data.Len())
	var packetBody = append(packetID.ToBytes(), data.Bytes()...)
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
				return false, err
			}
			if packetBody, err = ioutil.ReadAll(compressedBuffer); err != nil {
				return false, err
			}
			compressed = true
			uncompressedLength = length
		}
		if _, err := w.Write(VarInt(len(uncompressedLength.ToBytes()) + len(packetBody)).ToBytes()); err != nil {
			return false, fmt.Errorf("packetstream: write length (1)")
		}
		if _, err := w.Write(uncompressedLength.ToBytes()); err != nil {
			return false, fmt.Errorf("packetstream: write length (2)")
		}
		if _, err := w.Write(packetBody); err != nil {
			return false, fmt.Errorf("packetstream: write body")
		}
	} else {
		// Compression disabled
		if _, err := w.Write(append(length.ToBytes(), packetBody...)); err != nil {
			return false, fmt.Errorf("packetstream: write packet: %s", err)
		}
		// if _, err := w.Write(append(p.Length.ToBytes(), packetIDBytes...)); err != nil {
		// 	return fmt.Errorf("packetstream: write length: %s", err)
		// }
		// if _, err := io.Copy(w, p.Data); err != nil {
		// 	return fmt.Errorf("packetstream: write body: %s", err)
		// }
	}

	return compressed, nil
}

// func (s *PacketStreamReader) Close() error {
// 	return s.Close()
// }
