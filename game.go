package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

//go:generate stringer -type=GameState -trimprefix GameState
type GameState int

const (
	Handshaking GameState = iota
	Status
	Login
	Play
)

type Game struct {
	hostname    string
	accessToken string
	stream      *PacketStream
	state       GameState
}

type ServerInfo struct {
	Version struct {
		Name     string `json:"name"`
		Protocol int    `json:"protocol"`
	} `json:"version"`
	Description struct {
		Text string `json:"text"`
	} `json:"description"`
	Players struct {
		Max    int `json:"max"`
		Online int `json:"online"`
	} `json:"players"`
}

func NewGame(hostname string, accessToken string) *Game {
	return &Game{hostname, accessToken, nil, Handshaking}
}

func (g *Game) Send(p *Packet) error {
	var compressed bool
	var err error
	p.Direction = ServerBound
	p.State = g.state
	if compressed, err = g.stream.WritePacket(p.PacketID, p.Data); err != nil {
		return err
	}
	flags := ""
	if compressed {
		flags = "Z"
	}
	log.Tracef("game: >>%s %s", flags, p)
	return nil
}

func (g *Game) Receive() (*Packet, error) {
	var p *Packet
	var packetID VarInt
	var data *bytes.Buffer
	var compressed bool
	var err error
	if packetID, data, compressed, err = g.stream.ReadPacket(); err != nil {
		return p, err
	}
	p = &Packet{
		PacketID:  packetID,
		Data:      data,
		Direction: ClientBound,
		State:     g.state,
	}
	flags := ""
	if compressed {
		flags = "Z"
	}
	log.Tracef("game: <<%s %s", flags, p)
	return p, nil
}

func (g *Game) Stop() error {
	return g.stream.Close()
}

func (g *Game) dial() (*net.SRV, *net.TCPConn, error) {
	service, addr, err := ResolveAddr(g.hostname)
	if err != nil {
		return nil, nil, fmt.Errorf("game: resolve addr: %s", err)
	}
	log.Infof("game: connect addr: %s", addr)

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, nil, fmt.Errorf("game: %s", err)
	}

	g.setState(Handshaking)

	return service, conn, nil
}

func (g *Game) setState(state GameState) {
	g.state = state
	log.Warnf("game: \u001b[1;38;5;34mEntering \"%s\" state\u001b[0m", state)
}

func (g *Game) Status() (*ServerInfo, error) {
	var err error
	var resp *Packet

	var service *net.SRV
	var conn *net.TCPConn
	if service, conn, err = g.dial(); err != nil {
		return nil, err
	}

	g.stream = NewPacketStream(conn)

	// https://wiki.vg/Protocol_version_numbers
	// Switch to status state
	if err = g.Send(NewPacket(
		0x00,
		VarInt(340),
		String(service.Target),
		UInt16(service.Port),
		VarInt(1),
	)); err != nil {
		return nil, fmt.Errorf("game: write status state: %s", err)
	}

	// Request status
	// p.Data.Write([]byte{0})
	if err = g.Send(NewPacket(0x00)); err != nil {
		return nil, fmt.Errorf("game: write status request: %s", err)
	}

	g.setState(Status)

	// Receive status response
	log.Infof("game: Waiting for status response...")
	if resp, err = g.Receive(); err != nil {
		return nil, fmt.Errorf("game: read status response: %s", err)
	}
	var statusJSON String
	if err = statusJSON.FromReader(resp.Data); err != nil {
		return nil, fmt.Errorf("game: read status: %s", err)
	}
	serverInfo := &ServerInfo{}
	if err = json.Unmarshal([]byte(statusJSON), serverInfo); err != nil {
		return nil, fmt.Errorf("game: parse status JSON: %s", err)
	}
	log.Infof("game: Server info: %v", serverInfo)

	if err = g.Stop(); err != nil {
		return nil, fmt.Errorf("game: stop game: %s", err)
	}

	return serverInfo, nil
}

func (g *Game) Enter() error {
	var err error
	var req, resp *Packet

	var service *net.SRV
	var conn *net.TCPConn
	if service, conn, err = g.dial(); err != nil {
		return err
	}

	g.stream = NewPacketStream(conn)

	if err = g.Send(NewPacket(
		0x00,
		VarInt(340),
		String(service.Target),
		UInt16(service.Port),
		VarInt(2),
	)); err != nil {
		return fmt.Errorf("game: write login state: %s", err)
	}

	// Login start
	if err = g.Send(NewPacket(0x00, String("TheAnderson"))); err != nil {
		return fmt.Errorf("game: write login start: %s", err)
	}

	g.setState(Login)

	// Receive encryption request
	log.Infof("game: Waiting for encryption request...")
	if resp, err = g.Receive(); err != nil {
		return fmt.Errorf("game: read encryption request: %s", err)
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
		return fmt.Errorf("game: read server ID: %s", err)
	}
	log.Infof("game: Server ID: %s\n", serverID)
	if err = publicKeyBytes.FromReaderWithSize(resp.Data); err != nil {
		return fmt.Errorf("game: read public key: %s", err)
	}
	if err = verifyToken.FromReaderWithSize(resp.Data); err != nil {
		return fmt.Errorf("game: read verify token: %s", err)
	}
	log.Debugf("game: Public key: %v\n", publicKeyBytes)
	log.Debugf("game: Verify token: %v\n", verifyToken)

	var abstractPublicKey interface{}
	var publicKey *rsa.PublicKey
	if abstractPublicKey, err = x509.ParsePKIXPublicKey(publicKeyBytes); err != nil {
		return fmt.Errorf("game: parse public key: %s", err)
	}
	var ok bool
	if publicKey, ok = abstractPublicKey.(*rsa.PublicKey); !ok {
		return fmt.Errorf("game: process public key: %s", err)
	}
	log.Debugf("game: Parsed public key size: %d\n", publicKey.Size())

	// Send encryption response
	sharedSecret := []byte("1234123412341234")
	// accessToken := "eyJhbGciOiJIUzI1NiJ9.eyJ4dWlkIjoiMjUzNTQ1MDQ5NTg4MTM1NCIsImFnZyI6IkFkdWx0Iiwic3ViIjoiMTMwMDQ1MDQtYTMyYS00Njg3LTliMTUtNTA5M2NiNTc1OWQ0IiwibmJmIjoxNjQwMzQ1MzcwLCJhdXRoIjoiWEJPWCIsInJvbGVzIjpbXSwiaXNzIjoiYXV0aGVudGljYXRpb24iLCJleHAiOjE2NDA0MzE3NzAsImlhdCI6MTY0MDM0NTM3MCwicGxhdGZvcm0iOiJQQ19MQVVOQ0hFUiIsInl1aWQiOiI0ZDg1NjM0YjZhOThjYzExMjJhM2M3M2YwMWQ2YmU5NyJ9.nkqxv7Khv719wuxOioPuEJdg1pI2VhvHjnqDXTbj5WM"
	if err = Authenticate(string(serverID), sharedSecret, publicKeyBytes, g.accessToken, "777eb183d1b4463da0fd8f393191599f", "TheAnderson4252"); err != nil {
		return fmt.Errorf("game: authenticate: %s", err)
	}
	var encryptedSharedSecret, encryptedVerifyToken ByteArray
	if encryptedSharedSecret, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, sharedSecret); err != nil {
		return fmt.Errorf("game: encrypt shared secret: %s", err)
	}
	if encryptedVerifyToken, err = rsa.EncryptPKCS1v15(rand.Reader, publicKey, verifyToken); err != nil {
		return fmt.Errorf("game: encrypt verify token: %s", err)
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
	if err = g.Send(req); err != nil {
		return fmt.Errorf("game: write encryption response: %s", err)
	}

	// Switch to encrypted stream
	g.stream = NewPacketStreamWithEncryption(conn, NewAESCFB8(sharedSecret, false), NewAESCFB8(sharedSecret, true))

	// Wait for whatever comes next
	var p *Packet
	for g.state != Play {
		if p, err = g.Receive(); err != nil {
			return fmt.Errorf("game: %s", err)
		}
		if p.PacketID == ILoginSetCompression {
			var compressionThreshold VarInt
			if err = compressionThreshold.FromReader(p.Data); err != nil {
				panic(err)
			}
			log.Infof("game: Compression requested for packages > %d", compressionThreshold)
			g.stream.SetCompressionThreshold(int32(compressionThreshold))
		}
		if p.PacketID == ILoginSuccess {
			g.setState(Play)
		}
		// } else {
		// 	if p.PacketID == 0x18 {
		// 		var channel String
		// 		channel.FromReader(p.Data)
		// 		log.Infof("Plugin message: %s", channel)
		// 	}
	}

	return nil
}
