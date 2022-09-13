package main

import (
	"log"
	"net/netip"
	"time"

	"github.com/jomstead/go-rudp"
)

type packet struct {
	seq  uint32
	data []byte
	sent int64
}

type connection struct {
	unverified []packet
}

func (conn *connection) verify(verified uint32) {
	for i, p := range conn.unverified {
		if p.seq == verified {
			// remove the verified packet
			conn.unverified[i] = conn.unverified[len(conn.unverified)-1]
			conn.unverified = conn.unverified[:len(conn.unverified)-1]
			log.Printf("Verified: %d", verified)
			return
		}
	}
}

func main() {
	finished := make(chan bool)
	log.Println("Starting Server....")
	go Server()
	log.Println("Starting Clients...")
	Clients(1)
	<-finished //just run forever....

}

func Clients(num int) {
	for i := num; i > 0; i-- {
		go func(i int) {
			conn := connection{
				unverified: []packet{},
			}
			client, _ := rudp.Dial("udp4", "127.0.0.1", 8000)
			defer client.Close()
			log.Printf("[C%d] Connected to server", i)
			for i := 10; i > 0; i-- {
				// create and send a packet
				data := []byte{uint8(i)}
				n, seq, _ := client.Write(&data, true)
				// add the packet to the unverified list since it is a relaible packet
				conn.unverified = append(conn.unverified, packet{
					seq:  seq,
					data: data[:n],
					sent: time.Now().UnixMilli(),
				})
			}
			for {
				temp := make([]byte, 0, 1500)
				_, verified, _, _ := client.ReadFromUDP(temp)
				// remove verified packets from the list
				for _, v := range verified {
					conn.verify(v)
				}
			}
		}(i)
	}

}

func Server() {
	// store a map of all 'connected' clients
	clients := make(map[netip.AddrPort]connection)

	socket, _ := rudp.Listen("udp4", "127.0.0.1", 8000)
	defer socket.Close()

	func() {
		for {
			// create a buffer for the packet and read from socket
			temp := make([]byte, 1500)
			log.Println("[S] Waiting for packet")
			n, verified, client_addr, _ := socket.ReadFromUDP(temp)
			log.Printf("Received: %v", temp[:n])
			// if this is a new connection add it to the client list, otherwise get the client info from the list
			var client connection
			var ok bool
			if client, ok = clients[*client_addr]; !ok {
				client = connection{
					unverified: make([]packet, 0, 32),
				}
				clients[*client_addr] = client
			}

			// remove verified sequence numbers from the unverified list
			for _, ack := range verified {
				client.verify(ack)
			}
			//

			// send an Echo to the client
			response := temp[:n]
			_, seq, _ := socket.WriteToUDP(&response, *client_addr, true)
			//since this a reliable packet add it to the unverified list
			client.unverified = append(client.unverified, packet{
				seq:  seq,
				data: response,
				sent: time.Now().UnixMilli(),
			})
		}
	}()

}
