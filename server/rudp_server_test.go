package server

import (
	"net"
	"testing"
	"time"

	"github.com/jomstead/go-rudp/client"
	"github.com/jomstead/go-rudp/packet"
)

func TestRUDP_ServerReliablePacketsRemovedFromQueue(t *testing.T) {
	packets := make([]packet.Packet, 3)
	packets[0] = packet.Packet{Seq: 0, Data: []byte{1}, Timestamp: time.Now().UnixMilli()}
	packets[1] = packet.Packet{Seq: 1, Data: []byte{2}, Timestamp: time.Now().UnixMilli()}
	packets[2] = packet.Packet{Seq: 2, Data: []byte{3}, Timestamp: time.Now().UnixMilli()}
	conn := rUDPConnection{
		sent_packet_buffer: packets,
	}
	conn.processAck(2, 0b01)
	if len(conn.sent_packet_buffer) != 1 {
		t.Error("Verified packets are not being removed from sent packet buffer")
	}
}

func TestRUDP_ServerReliablePacketsRetransmissionTest(t *testing.T) {
	// setup the server
	address := "127.0.0.1:8000"
	s, _ := net.ResolveUDPAddr("udp4", address)
	c, _ := net.ListenUDP("udp4", s)
	server := RUDPServer{}
	server.Initialize(c, s)
	defer server.Close()

	// setup the client
	cc, _ := net.DialUDP("udp4", nil, s)
	client := client.RUDPClient{}
	client.Initialize(cc, s)

	// send a single packet to server
	client.Write(&[]byte{1}, false)

	// read the single packet on the server, now we have the clients address
	temp := make([]byte, 1024)
	n, client_addr, err := server.ReadFromUDP(temp)
	if err != nil {
		t.Error("Failed to receive unreliable packet from client")
	}
	if n != 1 {
		t.Error("ReadFromUDP reported wrong packet size for unreliable packet received")
	}

	// create a fake packet with a timestamp > 200ms ago and add it to the sent packet buffer
	packets := []packet.Packet{
		{Seq: 0, Data: []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, Timestamp: time.Now().UnixMilli() - 201},
		{Seq: 1, Data: []byte{1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1}, Timestamp: time.Now().UnixMilli() - 201}}
	server.connections[*client_addr].sent_packet_buffer = packets
	// process acks for that client saying we received sequence 1 but not yet seq 0.  Seq 0 should get retransmitted
	server.connections[*client_addr].processAck(1, 0)

	b := make([]byte, 1024)
	n, _, _ = client.ReadFromUDP(b)
	if n != 1 {
		t.Error("Server did not retransmit lost packet.")
	}

}

func TestRUDP_ServerPacketTest(t *testing.T) {
	// setup the server
	address := "127.0.0.1:8000"
	s, _ := net.ResolveUDPAddr("udp4", address)
	c, _ := net.ListenUDP("udp4", s)
	server := RUDPServer{}
	server.Initialize(c, s)
	defer server.Close()

	if !server.IsConnected() {
		t.Error("IsConnected returned false?")
	}

	// setup the client
	cc, _ := net.DialUDP("udp4", nil, s)
	client := client.RUDPClient{}
	client.Initialize(cc, s)

	// send a single packet to server
	client.Write(&[]byte{1}, false)

	// read the single packet on the server, now we have the clients address
	temp := make([]byte, 1024)
	_, client_addr, _ := server.ReadFromUDP(temp)

	// test reliable packet
	n, err := server.WriteToUDP(&[]byte{1}, *client_addr, true)
	if err != nil {
		t.Error("Failed to send reliable packet")
	}
	if n != 1 {
		t.Error("WriteToUDP reports wrong packet size when sending reliable packet")
	}

	n, _, err = client.ReadFromUDP(make([]byte, 1024))
	if err != nil {
		t.Error("Failed to receive reliable packet from server")
	}
	if n != 1 {
		t.Error("client ReadFromUDP reports wrong packet size received")
	}

	// test unreliable packet
	n, err = server.WriteToUDP(&[]byte{2}, *client_addr, false)
	if err != nil {
		t.Error("Failed to send unreliable packet")
	}
	if n != 1 {
		t.Error("WriteToUDP reports wrong packet size when sending unreliable packet")
	}

	n, _, err = client.ReadFromUDP(make([]byte, 1024))
	if err != nil {
		t.Error("Failed to receive unreliable packet from server")
	}
	if n != 1 {
		t.Error("client ReadFromUDP reports wrong packet size received")
	}

	// send a packet with an invalid byte[0]
	cc.Write([]byte{2, 0, 0, 0, 0, 0, 0, 0, 1})
	temp = make([]byte, 1024)
	_, _, err = server.ReadFromUDP(temp)
	if err == nil {
		t.Error("Didn't throw error for invalid packet header byte[0]")
	}

	// send a reliable packet
	client.Write(&[]byte{2, 0, 0, 0, 0, 0, 0, 0, 1}, true)
	temp = make([]byte, 1024)
	n, _, err = server.ReadFromUDP(temp)
	if err != nil {
		t.Error("Failed to receive reliable packet")
	}
	if n != 9 {
		t.Errorf("Returned incorrect packet size: %d should be 9", n)
	}

	client.Write(&[]byte{1, 2, 3}, true)
	_, _, err = server.ReadFromUDP(nil)
	if err == nil {
		t.Error("Read from UDP into a nil []byte?")
	}
}
