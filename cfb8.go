package main

import (
	"crypto/aes"
	"crypto/cipher"
)

// CFB stream with 8 bit segment size
// See http://csrc.nist.gov/publications/nistpubs/800-38a/sp800-38a.pdf
type CFB8 struct {
	b         cipher.Block
	blockSize int
	in        []byte
	out       []byte

	decrypt bool
}

func (x *CFB8) XORKeyStream(dst, src []byte) {
	for i := range src {
		x.b.Encrypt(x.out, x.in)
		copy(x.in[:x.blockSize-1], x.in[1:])
		if x.decrypt {
			x.in[x.blockSize-1] = src[i]
		}
		dst[i] = src[i] ^ x.out[0]
		if !x.decrypt {
			x.in[x.blockSize-1] = dst[i]
		}
	}
}

// NewCFB8Encrypter returns a Stream which encrypts with cipher feedback mode
// (segment size = 8), using the given Block. The iv must be the same length as
// the Block's block size.
// func NewCFB8Encrypter(block cipher.Block, iv []byte) cipher.Stream {
// 	return NewCFB8(block, iv, false)
// }

// NewCFB8Decrypter returns a Stream which decrypts with cipher feedback mode
// (segment size = 8), using the given Block. The iv must be the same length as
// the Block's block size.
// func NewCFB8Decrypter(block cipher.Block, iv []byte) cipher.Stream {
// 	return NewCFB8(block, iv, true)
// }

func NewCFB8(block cipher.Block, iv []byte, decrypt bool) cipher.Stream {
	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		// stack trace will indicate whether it was de or encryption
		panic("cipher.newCFB: IV length must equal block size")
	}
	x := &CFB8{
		b:         block,
		blockSize: blockSize,
		out:       make([]byte, blockSize),
		in:        make([]byte, blockSize),
		decrypt:   decrypt,
	}
	copy(x.in, iv)

	return x
}

func NewAESCFB8(key []byte, decrypt bool) cipher.Stream {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	return NewCFB8(block, key, decrypt)
}

// type AESCFB8ReadWriteCloser struct {
// 	stream io.ReadWriteCloser
// 	encrypter *CFB8
// 	decrypter *CFB8
// }

// func NewAESCFB8ReadWriteCloser(stream io.ReadWriteCloser, key []byte) *AESCFB8ReadWriteCloser {
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	cipher.StreamReader{
// 		cipher.NewCFBEncrypter(block, key),
// 		stream,
// 	}
// 	return &AESCFB8ReadWriteCloser{
// 		stream,
// 		NewCFB8(block, key, false),
// 		NewCFB8(block, key, true),
// 	}
// }

// func (c *AESCFB8ReadWriteCloser) Read(b []byte) (int, error) {
// 	src := make([]byte, len(b))
// 	n, err := c.stream.Read(b)
// 	if err != nil {
// 		return n, err
// 	}
// 	c.decrypter.XORKeyStream(b, src)
// 	return n, nil
// }

// func (c *AESCFB8ReadWriteCloser) Write(p []byte) (n int, err error) {
// 	dst := make([]byte, len(p))
// 	c.encrypter.XORKeyStream(dst, p)
// 	return c.stream.Write(dst)
// }

// func (c *AESCFB8ReadWriteCloser) Close() error {
// 	return c.stream.Close()
// }

// func (c *CFB8) Encrypt(ciphertext, plaintext, nonce []byte) ([]byte, error) {
// 	NewCFB8(c.block, c.iv, false)
// }
