package main_test

// import (
// 	"bytes"
// 	"io"
// 	"io/ioutil"
// 	"testing"

// 	x "github.com/and3rson/excavator"
// )

// // NopWriteCloser returns a WriteCloser with a no-op Close method wrapping
// // the provided Writer w.
// func NopWriteCloser(w io.Writer) io.WriteCloser {
// 	return nopWriteCloser{w}
// }

// type nopWriteCloser struct {
// 	io.Writer
// }

// func (nopWriteCloser) Close() error { return nil }

// func TestReadVarInt(t *testing.T) {
// 	buff := bytes.NewBuffer([]byte{})
// 	r := x.NewPacketStreamReader(io.NopCloser(buff))

// 	buff.ReadByte()

// 	for expected, input := range map[x.VarInt][]byte{
// 		x.VarInt(127): {0x7f},
// 		x.VarInt(128): {0x80, 0x01},
// 		x.VarInt(25565): {0xdd, 0xc7, 0x01},
// 		x.VarInt(2147483647): {0xff, 0xff, 0xff, 0xff, 0x07},
// 		x.VarInt(-1): {0xff, 0xff, 0xff, 0xff, 0x0f},
// 		x.VarInt(-2147483648): {0x80, 0x80, 0x80, 0x80, 0x08},
// 	} {

// 		buff.Write(input)
// 		actual := r.ReadVarInt()
// 		if expected != actual {
// 			t.Errorf("Expected %d, got %d", expected, actual)
// 		}
// 	}
// }

// func TestWriteVarInt(t *testing.T) {
// 	buff := bytes.NewBuffer([]byte{})
// 	w := x.NewPacketStreamWriter(NopWriteCloser(buff))

// 	buff.ReadByte()

// 	for input, expected := range map[x.VarInt][]byte{
// 		x.VarInt(127): {0x7f},
// 		x.VarInt(128): {0x80, 0x01},
// 		x.VarInt(25565): {0xdd, 0xc7, 0x01},
// 		x.VarInt(2147483647): {0xff, 0xff, 0xff, 0xff, 0x07},
// 		x.VarInt(-1): {0xff, 0xff, 0xff, 0xff, 0x0f},
// 		x.VarInt(-2147483648): {0x80, 0x80, 0x80, 0x80, 0x08},
// 	} {
// 		w.WriteVarInt(input)
// 		actual, _ := ioutil.ReadAll(buff)
// 		if bytes.Compare(expected, actual) != 0 {
// 			t.Errorf("Expected %v, got %v", expected, actual)
// 		}
// 	}
// }
