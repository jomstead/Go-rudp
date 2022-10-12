package wildnet

import (
	"crypto/sha1"
	"errors"
	"log"
	"math/rand"

	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

type Net struct {
	Packets  chan *Packet
	Presence map[uint32]*kcp.UDPSession
}

type Packet struct {
	PlayerID uint32
	Data     []byte
}

func (server *Net) Init() {
	server.Presence = make(map[uint32]*kcp.UDPSession)
	server.Packets = make(chan *Packet, 32768)
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	if listener, err := kcp.ListenWithOptions("127.0.0.1:12345", block, 10, 3); err == nil {
		for {
			s, err := listener.AcceptKCP()
			s.SetACKNoDelay(true)
			s.SetStreamMode(true)
			s.SetWriteDelay(false)
			s.SetNoDelay(1, 10, 2, 1)
			s.SetMtu(1024)
			if err != nil {
				log.Fatal(err)
			}
			go server.handle(s)
		}
	} else {
		log.Fatal(err)
	}
}

func (server *Net) SendTo(playerID uint32, packet []byte) (int, error) {
	if addr, ok := server.Presence[playerID]; ok {
		return addr.Write(packet)
	}
	return 0, errors.New("presence with that id not found")
}

func (server *Net) handle(conn *kcp.UDPSession) {
	// TODO: check the redis server for an entry with that id and the unique passcode
	// TODO: then remove the entry from redis and add the id and connection to the presences map
	var id uint32 = uint32(rand.Int31())
	server.Presence[id] = conn
	for {
		buf := make([]byte, 1024) // TODO: use a ring buffer or something...
		n, err := conn.Read(buf)
		if err != nil || n < 1 {
			log.Println(err)
			return
		}
		server.process(id, buf[:n])
	}
}

func (server *Net) process(playerID uint32, buf []byte) {
	p := Packet{
		PlayerID: playerID,
		Data:     buf,
	}
	// add every packet received to the packets list to be processed during the world update
	server.Packets <- &p
}
