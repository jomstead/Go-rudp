package server

import (
	"encoding/binary"
	"errors"
	"net"
	"net/netip"

	"github.com/jomstead/go-rudp/packet"
)

type RUDPServer struct {
	conn        *net.UDPConn
	address     *net.UDPAddr //host:port
	isConnected bool
	connections map[netip.AddrPort]*rUDPConnection
	temp        []byte
}

type rUDPConnection struct {
	addr        netip.AddrPort
	isConnected bool
	seq         uint32
	remote_seq  uint32
	remote_acks packet.Ack
	unverified  []uint32 // keeps a list of unverified sequence numbers
	server      *RUDPServer
}

func (conn *RUDPServer) Initialize(c *net.UDPConn, s *net.UDPAddr) {
	conn.isConnected = true // is the server running
	conn.address = s        // address for the server (this machine)
	conn.conn = c           // connection for the server
	conn.temp = make([]byte, 1024)
	conn.connections = make(map[netip.AddrPort]*rUDPConnection)

}

/* WriteToUDP acts like Write but sends the packet to an UDPAddr */
func (conn *RUDPServer) WriteToUDP(payload *[]byte, addr netip.AddrPort, reliable bool) (int, error) {
	// TODO: Get the next sequence number FOR THAT CONNECTION (UDPAddr)
	client := conn.connections[addr]

	var data []byte
	var seq uint32
	index := 0
	if reliable {
		data = make([]byte, 13, len(*payload)+13)
		data[0] = 1
		// increase sequence number for reliable packets
		client.seq += 1
		seq = client.seq
		binary.BigEndian.PutUint32(data[1:], client.seq)
		index = 5
	} else {
		data = make([]byte, 9, len(*payload)+9)
		data[0] = 0
		index = 1
	}
	// include the last received sequence number and the sequence history from the remote source
	binary.BigEndian.PutUint32(data[index:], client.remote_seq)
	binary.BigEndian.PutUint32(data[index+4:], client.remote_acks.Data)
	index += 8
	data = append(data, *payload...)
	if reliable {
		// keep a list of unverified sequence numbers
		client.unverified = append(client.unverified, seq)
	}

	n, err := conn.conn.WriteToUDPAddrPort(data, addr)
	return n - index, err
}

func (conn *RUDPServer) Close() {
	if conn.conn != nil {
		conn.conn.Close()
	}
	conn.isConnected = false
}

func (conn RUDPServer) IsConnected() bool {
	return conn.isConnected
}

func (conn *RUDPServer) ReadFromUDP(buffer []byte) (n int, verified []uint32, addr *netip.AddrPort, err error) {
	// use a temp buffer to read a packet from that client
	if buffer == nil {
		return 0, []uint32{}, nil, errors.New("buffer not initialized")
	}
	n, client_addr, err := conn.conn.ReadFromUDPAddrPort(conn.temp)
	addr = &client_addr
	if err != nil {
		return n, []uint32{}, addr, err
	}
	// create a new rUDPConnection for each new addr
	var client *rUDPConnection
	client = conn.connections[*addr]
	if client == nil {
		client = &rUDPConnection{
			isConnected: true,
			seq:         ^uint32(0),
			remote_seq:  ^uint32(0), // remote seq number
			server:      conn,
			unverified:  make([]uint32, 0, 16), // queue of unverified seuquence numbers
			remote_acks: packet.Ack{Data: 0},
			addr:        *addr,
		}
		conn.connections[*addr] = client
	}

	if n > 5 && conn.temp[0] == 0 {
		// unreliable packet
		ack := binary.BigEndian.Uint32(conn.temp[1:5])
		ack_bitfield := binary.BigEndian.Uint32(conn.temp[5:9])
		verified = client.processAck(ack, ack_bitfield)
		copy(buffer, conn.temp[9:n])
		return n - 9, verified, addr, err
	}
	if n > 8 && conn.temp[0] == 1 {
		// reliable packet
		seq := binary.BigEndian.Uint32(conn.temp[1:5])
		client.remote_seq = packet.UpdateAcknowledgements(seq, client.remote_seq, &client.remote_acks)
		ack := binary.BigEndian.Uint32(conn.temp[5:9])
		ack_bitfield := binary.BigEndian.Uint32(conn.temp[9:13])
		verified = client.processAck(ack, ack_bitfield)
		copy(buffer, conn.temp[13:n])

		return n - 13, verified, addr, err
	}
	// Not sure what this is....
	return n, []uint32{}, addr, errors.New("unexpected RUDP header data")
}

// ProcessAck takes the acknowledgements from the remote resource and removes packets from the local
// reliable packet buffer that have been confirmed as sent
func (conn *rUDPConnection) processAck(seq uint32, bitwise uint32) []uint32 {
	bits := packet.Ack{Data: bitwise}
	count := len(conn.unverified)
	i := 0
	verified := make([]uint32, 0)

	for i < count {
		unver_seq := conn.unverified[i]
		// check if this packet in the buffer has been verified as delivered.
		// It is verified if either the sequence number is the same as the received sequence number, or
		// if the bitwise bit for that packet is set in the bitwise field(which holds the last 32 acknowledgements)
		if unver_seq == seq || bits.Has(seq-unver_seq-1) {
			//overwrite the unverified in the buffer with the last unverifed in the buffer list
			conn.unverified[i] = conn.unverified[count-1]
			// then remove the last packet in the list since we moved it to a new spot in the list
			count--
			conn.unverified = conn.unverified[0:count]
			// add the verified packet to the verified list to return
			verified = append(verified, unver_seq)
		} else {
			// this packet hasn't been verified, move on to check the next one
			i++
		}

	}
	return verified
}
