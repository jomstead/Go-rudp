package client

import (
	"net"
	"testing"

	"github.com/jomstead/go-rudp/server"
)

func TestRUDP_ClientReliablePacketsRemovedFromQueue(t *testing.T) {
	packets := []uint32{0, 1, 2, 3}
	conn := RUDPClient{
		unverified: packets,
	}
	conn.processAck(2, 0b01)
	if len(conn.unverified) != 2 {
		t.Error("Verified packets are not being removed from unverified list")
	}
}

func TestRUDP_ClientReliablePacketsRetransmissionTest(t *testing.T) {
	// setup the server
	address := "127.0.0.1:8000"
	s, _ := net.ResolveUDPAddr("udp4", address)
	c, _ := net.ListenUDP("udp4", s)
	server := server.RUDPServer{}
	server.Initialize(c, s)
	defer server.Close()

	// setup the client
	cc, _ := net.DialUDP("udp4", nil, s)
	client := RUDPClient{}
	client.Initialize(cc, s)

	// create a fake packet with a timestamp > 200ms ago and add it to the sent packet buffer
	packets := []uint32{0, 1, 2, 3, 4}
	client.unverified = packets

	verified := client.processAck(3, 0b100) // should be 0 and 3
	if len(verified) != 2 {
		t.Error("Process ack did not generate the correct list.")
	}

	verified = client.processAck(4, 0b1101) // should be 1 and 4
	if len(verified) != 2 {
		t.Error("Process ack did not generate the correct list.")
	}

}

func TestRUDP_ClientPacketTest(t *testing.T) {
	// setup the server
	address := "127.0.0.1:8000"
	s, _ := net.ResolveUDPAddr("udp4", address)
	server_conn, _ := net.ListenUDP("udp4", s)
	server := server.RUDPServer{}
	server.Initialize(server_conn, s)
	defer server.Close()

	// setup the client
	cc, _ := net.DialUDP("udp4", nil, s)
	client := RUDPClient{}
	client.Initialize(cc, s)
	defer client.Close()

	if !client.IsConnected() {
		t.Error("IsConnected returned false?")
	}

	// Client send unreliable
	n, err := client.Write(&[]byte{1}, false)
	if err != nil {
		t.Error("Error while sending packet to server")
	}
	if n != 1 {
		t.Error("Client write returned wrong number of bytes sent")
	}

	// read the single packet on the server, now we have the clients address
	temp := make([]byte, 1024)
	n, client_addr, err := server.ReadFromUDP(temp)
	if err != nil {
		t.Error("Error receiving packet from client")
	}
	if n != 1 {
		t.Error("ReadFromUDP reported the incorrect number of bytes received")
	}

	// Client send reliable
	n, err = client.Write(&[]byte{1}, true)
	if err != nil {
		t.Error("Failed to send reliable packet")
	}
	if n != 1 {
		t.Error("WriteToUDP reported incorrect number of bytes sent when sending reliable packet")
	}

	n, _, err = server.ReadFromUDP(make([]byte, 1024))
	if err != nil {
		t.Error("Failed to receive reliable packet from server")
	}
	if n != 1 {
		t.Error("ReadFromUDP reported incorrect number of bytes received from client when receiving reliable packet")
	}

	// send a packet with an invalid byte[0]
	server_conn.WriteToUDPAddrPort([]byte{2, 0, 0, 0, 0, 0, 0, 0, 1}, *client_addr)
	temp = make([]byte, 1024)
	n, _, err = client.ReadFromUDP(temp)
	if err == nil {
		t.Error("Didn't throw error for invalid packet header byte[0]")
	}
	if n != 0 {
		t.Error("When we receive an invalid header byte[0] the packet size should be reported as 0")
	}

	// Test client read with nil buffer
	server.WriteToUDP(&[]byte{1, 2, 3}, *client_addr, true)
	_, _, err = client.ReadFromUDP(nil)
	if err == nil {
		t.Error("Read from UDP into a nil []byte?")
	}

	// Test client receive reliable
	data := make([]byte, 1024)
	n, _, err = client.ReadFromUDP(data)
	if err != nil {
		t.Error("Read from UDP failed")
	}
	if n != 3 {
		t.Error("ReadFromUDP reported wrong number of bytes read for reliable packet")
	}

	// Test client receive unreliable
	server.WriteToUDP(&[]byte{1, 2, 3}, *client_addr, false)
	n, _, err = client.ReadFromUDP(data)
	if err != nil {
		t.Error("Read from UDP failed")
	}
	if n != 3 {
		t.Error("ReadFromUDP reported wrong number of bytes read for reliable packet")
	}

}
