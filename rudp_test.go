package rudp

import (
	"testing"
)

func TestRUDP_ServerListen(t *testing.T) {
	socket, _ := Listen("udp4", "127.0.0.1", 8000)
	if !socket.IsConnected() {
		t.Error("Expected socket to be listening on 127.0.0.1:8000")
	}
	socket.Close()
	socket, err := Listen("udp4", "127.0.0::", 8000)
	if err == nil {
		t.Error("Expected bad server listen address")
	}
	if socket != nil {
		socket.Close()
	}
}

func TestRUDP_ClientDial(t *testing.T) {
	socket, _ := Dial("udp4", "127.0.0.1", 8000)
	if !socket.IsConnected() {
		t.Error("Expected client to connect")
	}
	socket.Close()
	socket, err := Dial("udp4", "127.0.0::", 8000)
	if err == nil {
		t.Error("Expected bad server dial address")
	}
	if socket != nil {
		socket.Close()
	}
	socket, err = Dial("WHAT", "127.0.0.1", 8000)
	if err == nil {
		t.Error("Expected bad network type")
	}
	if socket != nil {
		socket.Close()
	}
}
