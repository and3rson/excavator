package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
)

func main() {
	hostname := "20b20t.org"
	cname, addrs, err := net.LookupSRV("minecraft", "tcp", hostname)
	if err != nil {
		panic(err)
	}
	fmt.Println(cname, addrs, err)
	service := addrs[0]
	fmt.Println(service.Target, service.Port)

	addr, err := net.ResolveTCPAddr("tcp", service.Target + ":" + fmt.Sprint(service.Port))
	if err != nil {
		panic(err)
	}

	fmt.Println("Connect addr:", addr)

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err)
	}

	stream := NewPacketStream(conn)

	var req, resp *Packet

	// https://wiki.vg/Protocol_version_numbers
	// Switch to status state
	req = NewPacket(0x00)
	req.Data.Write(VarInt(340).ToBytes())
	req.Data.Write(String(service.Target).ToBytes())
	_ = binary.Write(req.Data, binary.BigEndian, service.Port)
	req.Data.Write(VarInt(1).ToBytes())
	if err := stream.WritePacket(req); err != nil {
		panic(err)
	}
	// Request status
	req = NewPacket(0x00)
	// p.Data.Write([]byte{0})
	if err := stream.WritePacket(req); err != nil {
		panic(err)
	}

	// Receive status response
	fmt.Println("Waiting for status response...")
	resp, err = stream.ReadPacket()
	if err != nil {
		panic(err)
	}
	statusData, _ := ioutil.ReadAll(resp.Data)
	fmt.Println(string(statusData))

	// for _, i := range VarInt(340).ToBytes() {
	// 	fmt.Println("XX", i)
	// }

	if err = stream.PacketStreamReader.Close(); err != nil {
		panic(err)
	}

	conn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		panic(err)
	}
	stream = NewPacketStream(conn)

	// Switch to login state
	req = NewPacket(0x00)
	req.Data.Write(VarInt(340).ToBytes())
	req.Data.Write(String(service.Target).ToBytes())
	_ = binary.Write(req.Data, binary.BigEndian, service.Port)
	req.Data.Write(VarInt(2).ToBytes())
	if err := stream.WritePacket(req); err != nil {
		panic(err)
	}

	// Login start
	req = NewPacket(0x00)
	req.Data.Write(String("TheAnderson").ToBytes())
	if err := stream.WritePacket(req); err != nil {
		panic(err)
	}

	// Receive encryption request
	fmt.Println("Waiting for encryption request...")
	resp, err = stream.ReadPacket()
	if err != nil {
		panic(err)
	}
	// loginData, _ := ioutil.ReadAll(resp.Data)
	// fmt.Println(string(loginData))
	var serverID String
	var publicKeyBytes, verifyToken ByteArray
	// fmt.Println("...")
	// buf := []byte{0, 0, 0, 0, 0}
	// resp.Data.Read(buf)
	// fmt.Println("B", buf)

	if err = serverID.FromReader(resp.Data); err != nil {
		panic(err)
	}
	fmt.Printf("Server ID: %s\n", serverID)
	if err = publicKeyBytes.FromReaderWithSize(resp.Data); err != nil {
		panic(err)
	}
	if err = verifyToken.FromReaderWithSize(resp.Data); err != nil {
		panic(err)
	}
	fmt.Printf("Public key: %v\n", publicKeyBytes)
	fmt.Printf("Verify token: %v\n", verifyToken)

	var abstractPublicKey interface{}
	var publicKey *rsa.PublicKey
	if abstractPublicKey, err = x509.ParsePKIXPublicKey(publicKeyBytes); err != nil {
		panic(err)
	}
	var ok bool
	if publicKey, ok = abstractPublicKey.(*rsa.PublicKey); !ok {
		panic(fmt.Errorf("Failed to cast public key to *rsa.PublicKey"))
	}
	fmt.Printf("Parsed public key size: %d\n", publicKey.Size())

	// Send encryption response
	sharedSecret := []byte("1234123412341234")
	var encryptedSharedSecret, encryptedVerifyToken ByteArray
	if encryptedSharedSecret, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, sharedSecret); err != nil {
		panic(err)
	}
	if encryptedVerifyToken, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, verifyToken); err != nil {
		panic(err)
	}
	req = NewPacket(0x01)
	req.Data.Write(encryptedSharedSecret.ToBytesWithSize())
	req.Data.Write(encryptedVerifyToken.ToBytesWithSize())
	if err := stream.WritePacket(req); err != nil {
		panic(err)
	}

	// Switch to encrypted stream
	stream = NewPacketStream(NewAESCFB8ReadWriteCloser(conn, sharedSecret))

	// Wait for whatever comes next
	resp, err = stream.ReadPacket()
	if err != nil {
		panic(err)
	}

	// fmt.Println(packet)
}
