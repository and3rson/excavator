package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"net"
	log "github.com/sirupsen/logrus"
)

func main() {
	// log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		// DisableColors: false,
		FullTimestamp: true,
		DisableLevelTruncation: true,
		PadLevelText: true,
	})
	log.SetLevel(log.TraceLevel)

	hostname := "20b20t.org"
	cname, addrs, err := net.LookupSRV("minecraft", "tcp", hostname)
	if err != nil {
		panic(err)
	}
	log.Debugf("CNAME: %s, addrs: %v", cname, addrs)
	service := addrs[0]
	log.Infof("Target/port: %s:%d", service.Target, service.Port)

	addr, err := net.ResolveTCPAddr("tcp", service.Target + ":" + fmt.Sprint(service.Port))
	if err != nil {
		panic(err)
	}

	log.Infof("Connect addr: %s", addr)

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
	log.Infof("Waiting for status response...")
	resp, err = stream.ReadPacket()
	if err != nil {
		panic(err)
	}
	var statusJSON String
	if err = statusJSON.FromReader(resp.Data); err != nil {
		panic(err)
	}
	log.Infof("Status data: %v", string(statusJSON))

	// for _, i := range VarInt(340).ToBytes() {
	// 	fmt.Println("XX", i)
	// }

	if err = stream.Close(); err != nil {
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
	log.Infof("Waiting for encryption request...")
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
	log.Infof("Server ID: %s\n", serverID)
	if err = publicKeyBytes.FromReaderWithSize(resp.Data); err != nil {
		panic(err)
	}
	if err = verifyToken.FromReaderWithSize(resp.Data); err != nil {
		panic(err)
	}
	log.Debugf("Public key: %v\n", publicKeyBytes)
	log.Debugf("Verify token: %v\n", verifyToken)

	var abstractPublicKey interface{}
	var publicKey *rsa.PublicKey
	if abstractPublicKey, err = x509.ParsePKIXPublicKey(publicKeyBytes); err != nil {
		panic(err)
	}
	var ok bool
	if publicKey, ok = abstractPublicKey.(*rsa.PublicKey); !ok {
		panic(fmt.Errorf("Failed to cast public key to *rsa.PublicKey"))
	}
	log.Debugf("Parsed public key size: %d\n", publicKey.Size())

	// Send encryption response
	sharedSecret := []byte("1234123412341234")
	accessToken := "eyJhbGciOiJIUzI1NiJ9.eyJ4dWlkIjoiMjUzNTQ1MDQ5NTg4MTM1NCIsImFnZyI6IkFkdWx0Iiwic3ViIjoiMTMwMDQ1MDQtYTMyYS00Njg3LTliMTUtNTA5M2NiNTc1OWQ0IiwibmJmIjoxNjQwMzQ1MzcwLCJhdXRoIjoiWEJPWCIsInJvbGVzIjpbXSwiaXNzIjoiYXV0aGVudGljYXRpb24iLCJleHAiOjE2NDA0MzE3NzAsImlhdCI6MTY0MDM0NTM3MCwicGxhdGZvcm0iOiJQQ19MQVVOQ0hFUiIsInl1aWQiOiI0ZDg1NjM0YjZhOThjYzExMjJhM2M3M2YwMWQ2YmU5NyJ9.nkqxv7Khv719wuxOioPuEJdg1pI2VhvHjnqDXTbj5WM"
	if err = Authenticate(string(serverID), sharedSecret, publicKeyBytes, accessToken, "777eb183d1b4463da0fd8f393191599f", "TheAnderson4252"); err != nil {
		panic(err)
	}
	var encryptedSharedSecret, encryptedVerifyToken ByteArray
	if encryptedSharedSecret, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, sharedSecret); err != nil {
		panic(err)
	}
	if encryptedVerifyToken, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, verifyToken); err != nil {
		panic(err)
	}
	req = NewPacket(0x01)
	// fmt.Println("!", encryptedSharedSecret)
	// for i := 0; i < 100; i++ {
	// 	encryptedVerifyToken[i] = 0
	// }
	// fmt.Println("!!", encryptedSharedSecret)
	req.Data.Write(encryptedSharedSecret.ToBytesWithSize())
	req.Data.Write(encryptedVerifyToken.ToBytesWithSize())
	// fmt.Println("!!!", req.String())
	if err := stream.WritePacket(req); err != nil {
		panic(err)
	}

	// Switch to encrypted stream
	stream = NewPacketStreamWithEncryption(conn, NewAESCFB8(sharedSecret, false), NewAESCFB8(sharedSecret, true))

	inGame := false

	// Wait for whatever comes next
	var p *Packet
	for {
		p, err = stream.ReadPacket()
		if err != nil {
			panic(err)
		}
		if !inGame {
			if p.PacketID == 0x03 {
				var compressionThreshold VarInt
				if err = compressionThreshold.FromReader(p.Data); err != nil {
					panic(err)
				}
				log.Infof("Compression requested for packages > %d", compressionThreshold)
				stream.SetCompressionThreshold(int32(compressionThreshold))
			}
			if p.PacketID == 0x02 {
				log.Warn("===========================")
				log.Warn("=== Entering play state ===")
				log.Warn("===========================")
				inGame = true
			}
		} else {
			if p.PacketID == 0x18 {
				var channel String
				channel.FromReader(p.Data)
				log.Infof("Plugin message: %s", channel)
			}
		}
	}

	// fmt.Println(packet)
}
