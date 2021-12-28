package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	// log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		// DisableColors: false,
		FullTimestamp:          false,
		DisableLevelTruncation: true,
		PadLevelText:           true,
	})

	var isDebug bool
	flag.BoolVar(&isDebug, "v", false, "Verbose debug logs")
	flag.Parse()
	if isDebug {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	var accessToken string
	if accessToken = os.Getenv("ACCESS_TOKEN"); accessToken == "" {
		panic("Missing ACCESS_TOKEN env var")
	}

	// hostname := "2b2t.org"
	// hostname := "test.2b2t.org"
	// hostname := "9b9t.com"
	hostname := "20b20t.org"

	// stream := NewPacketStream(conn)

	game := NewGame(hostname, accessToken)
	if status, err := game.Status(); err != nil {
		panic(err)
	} else {
		log.Infof("main: Server status: %s", status)
	}
	if err := game.Enter(); err != nil {
		panic(err)
	}

	// Wait for whatever comes next
	var p *Packet
	for {
		var err error
		if p, err = game.Receive(); err != nil {
			panic(err)
		}
		if p.PacketID == IPlayCustomPayload {
			var channel String
			channel.FromReader(p.Data)
			data := string(p.Data.Bytes())
			log.Infof("main: Plugin message: channel=%s, data=%s", channel, data)
		} else if p.PacketID == IPlayChat {
			var chat Chat
			chat.FromReader(p.Data)
			log.Infof("main: Chat message: \u001b[38;5;127m%s\u001b[0m", chat.String())
		} else if p.PacketID == IPlayKickDisconnect {
			var chat Chat
			chat.FromReader(p.Data)
			log.Errorf("main: Disconnected: \u001b[38;5;160m%s\u001b[0m", chat.String())
			game.Stop()
			break
		} else if p.PacketID == IPlayAbilities {
			b, _ := p.Data.ReadByte()
			log.Infof("main: Invulnerable: %v", b&0x01 > 0)
			log.Infof("main: Flying: %v", b&0x02 > 0)
			log.Infof("main: Allow Flying: %v", b&0x04 > 0)
			log.Infof("main: Creative Mode: %v", b&0x08 > 0)
		} else if p.PacketID == IPlayPlayerListHeaderFooter {
			var header, footer Chat
			header.FromReader(p.Data)
			footer.FromReader(p.Data)
			log.Infof("main: Player list header: %s", header.String())
			log.Infof("main: Player list footer: %s", footer.String())
		} else if p.PacketID == IPlayKeepAlive {
			log.Infof("main: Keeping alive")
			var id Long
			id.FromReader(p.Data)
			game.Send(NewPacket(0x0b, id))
		} else {
			// game.Stream.WritePacket(NewPacket(0x01, Long(42)))
		}
	}

	// fmt.Println(packet)
}
