package server

import (
	"net"
	"testing"

	"github.com/jomstead/go-rudp/client"
)

func TestRUDP_ServerReliablePacketsRemovedFromQueue(t *testing.T) {
	packets := []uint32{0, 1, 2, 3}

	conn := rUDPConnection{
		unverified: packets,
	}
	conn.processAck(2, 0b01)
	if len(conn.unverified) != 2 {
		t.Error("Verified packets are not being removed from unverified list")
	}
}

func TestRUDP_ServerReliablePacketsVerifiedTest(t *testing.T) {
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
	n, _, client_addr, err := server.ReadFromUDP(temp)
	if err != nil {
		t.Error("Failed to receive unreliable packet from client")
	}
	if n != 1 {
		t.Error("ReadFromUDP reported wrong packet size for unreliable packet received")
	}

	// create a fake unverifed list
	packets := []uint32{0, 1}
	server.connections[*client_addr].unverified = packets
	server.connections[*client_addr].seq = 1
	// process acks for that client saying we received sequence 1 but not yet seq 0.  Seq 1 should get sent back as verified
	verified := server.connections[*client_addr].processAck(1, 0)

	if len(verified) != 1 {
		t.Error("Process ack did not generate the correct list.")
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
	_, _, client_addr, _ := server.ReadFromUDP(temp)

	// test reliable packet
	n, seq, err := server.WriteToUDP(&[]byte{1}, *client_addr, true)
	if err != nil {
		t.Error("Failed to send reliable packet")
	}
	if n != 1 {
		t.Error("WriteToUDP reports wrong packet size when sending reliable packet")
	}
	if seq != 0 {
		t.Error("Correct sequence number not returned")
	}

	n, _, _, err = client.ReadFromUDP(make([]byte, 1024))
	if err != nil {
		t.Error("Failed to receive reliable packet from server")
	}
	if n != 1 {
		t.Error("client ReadFromUDP reports wrong packet size received")
	}

	// test unreliable packet
	n, _, err = server.WriteToUDP(&[]byte{2}, *client_addr, false)
	if err != nil {
		t.Error("Failed to send unreliable packet")
	}
	if n != 1 {
		t.Error("WriteToUDP reports wrong packet size when sending unreliable packet")
	}

	n, _, _, err = client.ReadFromUDP(make([]byte, 1024))
	if err != nil {
		t.Error("Failed to receive unreliable packet from server")
	}
	if n != 1 {
		t.Error("client ReadFromUDP reports wrong packet size received")
	}

	// send a packet with an invalid byte[0]
	cc.Write([]byte{2, 0, 0, 0, 0, 0, 0, 0, 1})
	temp = make([]byte, 1024)
	_, _, _, err = server.ReadFromUDP(temp)
	if err == nil {
		t.Error("Didn't throw error for invalid packet header byte[0]")
	}

	// send a reliable packet
	client.Write(&[]byte{2, 0, 0, 0, 0, 0, 0, 0, 1}, true)
	temp = make([]byte, 1024)
	n, _, _, err = server.ReadFromUDP(temp)
	if err != nil {
		t.Error("Failed to receive reliable packet")
	}
	if n != 9 {
		t.Errorf("Returned incorrect packet size: %d should be 9", n)
	}

	client.Write(&[]byte{1, 2, 3}, true)
	_, _, _, err = server.ReadFromUDP(nil)
	if err == nil {
		t.Error("Read from UDP into a nil []byte?")
	}
}
