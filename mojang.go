package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func AuthDigest(serverID string, sharedSecret, publicKey []byte) string {
	h := sha1.New()
	h.Write([]byte(serverID))
	h.Write(sharedSecret)
	h.Write(publicKey)
	hash := h.Sum(nil)

	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = TwosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(fmt.Sprintf("%x", hash), "0")
	if negative {
		res = "-" + res
	}

	return res
}

func TwosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = ^p[i]
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}
type Profile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AuthRequest struct {
	AccessToken     string  `json:"accessToken"`
	SelectedProfile Profile `json:"selectedProfile"`
	ServerID        string  `json:"serverId"`
}

func Authenticate(serverID string, sharedSecret, publicKey []byte, accessToken, uuid, name string) error {
	var requestPacket []byte
	var err error
	var req *http.Request
	var resp *http.Response
	client := http.Client{}
	if requestPacket, err = json.Marshal(
		AuthRequest{
			AccessToken: accessToken,
			SelectedProfile: Profile{
				ID:   uuid,
				Name: name,
			},
			ServerID: AuthDigest(serverID, sharedSecret, publicKey),
		},
	); err != nil {
		return err
	}
	if req, err = http.NewRequest(
		http.MethodPost,
		"https://sessionserver.mojang.com/session/minecraft/join",
		bytes.NewReader(requestPacket),
	); err != nil {
		return err
	}
	req.Header.Set("User-agent", "go-mc")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	if resp, err = client.Do(req); err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("auth fail: %s", string(body))
	}
	return nil
}
