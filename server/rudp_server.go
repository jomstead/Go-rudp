package server

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"net/netip"
	"time"

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
	addr               netip.AddrPort
	isConnected        bool
	seq                uint32
	remote_seq         uint32
	remote_acks        packet.Ack
	sent_packet_buffer []packet.Packet // keeps a buffer of sent reliable packets, packets get removed from this slice as they are confirm received
	server             *RUDPServer
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
		// save packets incase we need to resend them
		client.sent_packet_buffer = append(client.sent_packet_buffer, packet.Packet{Seq: seq, Data: data, Timestamp: time.Now().UnixMilli()})
		log.Printf("Sent to client: %d %d %b", client.seq, client.remote_seq, client.remote_acks.Data)
	} else {
		log.Printf("Sent to client: %d %b", client.remote_seq, client.remote_acks.Data)

	}
	log.Printf("[S] Sent %d", seq)
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

func (conn *RUDPServer) ReadFromUDP(buffer []byte) (n int, addr *netip.AddrPort, err error) {
	// use a temp buffer to read a packet from that client
	if buffer == nil {
		return 0, nil, errors.New("buffer not initialized")
	}
	n, client_addr, err := conn.conn.ReadFromUDPAddrPort(conn.temp)
	addr = &client_addr
	if err != nil {
		return n, addr, err
	}
	// create a new rUDPConnection for each new addr
	var client *rUDPConnection
	client = conn.connections[*addr]
	if client == nil {
		client = &rUDPConnection{
			isConnected:        true,
			seq:                ^uint32(0),
			remote_seq:         ^uint32(0), // remote seq number
			server:             conn,
			sent_packet_buffer: make([]packet.Packet, 0, 16), // queue of outbound reliable packets
			remote_acks:        packet.Ack{Data: 0},
			addr:               *addr,
		}
		conn.connections[*addr] = client
	}

	if n > 5 && conn.temp[0] == 0 {
		// unreliable packet
		ack := binary.BigEndian.Uint32(conn.temp[1:5])
		ack_bitfield := binary.BigEndian.Uint32(conn.temp[5:9])
		if len(buffer) < n-9 {
			return 0, nil, errors.New("buffer too small for packet")
		}
		client.processAck(ack, ack_bitfield)
		copy(buffer, conn.temp[9:n])
		log.Printf("[S] %d %b", ack, ack_bitfield)
		return n - 9, addr, err
	}
	if n > 8 && conn.temp[0] == 1 {
		// reliable packet
		seq := binary.BigEndian.Uint32(conn.temp[1:5])
		client.remote_seq = packet.UpdateAcknowledgements(seq, client.remote_seq, &client.remote_acks)
		ack := binary.BigEndian.Uint32(conn.temp[5:9])
		ack_bitfield := binary.BigEndian.Uint32(conn.temp[9:13])
		if len(buffer) < n-13 {
			return 0, nil, errors.New("buffer too small for packet")
		}
		client.processAck(ack, ack_bitfield)
		copy(buffer, conn.temp[13:n])

		log.Printf("[S] %d %d %b", seq, ack, ack_bitfield)
		return n - 13, addr, err
	}
	// Not sure what this is....
	return n, addr, errors.New("unexpected RUDP header data")
}

// ProcessAck takes the acknowledgements from the remote resource and removes packets from the local
// reliable packet buffer that have been confirmed as sent
func (conn *rUDPConnection) processAck(seq uint32, bitwise uint32) {
	bits := packet.Ack{Data: bitwise}
	count := len(conn.sent_packet_buffer)
	i := 0
	for i < count {
		p := conn.sent_packet_buffer[i]
		// check if this packet in the buffer has been verified as delivered.
		// It is verified if either the sequence number is the same as the received sequence number, or
		// if the bitwise bit for that packet is set in the bitwise field(which holds the last 32 acknowledgements)
		if p.Seq == seq || bits.Has(seq-p.Seq-1) {
			//overwrite the packet in the buffer with the last packet in the buffer list
			conn.sent_packet_buffer[i] = conn.sent_packet_buffer[count-1]
			// then remove the last packet in the list since we moved it to a new spot in the list
			count--
			conn.sent_packet_buffer = conn.sent_packet_buffer[0:count]
		} else {
			// check if the packet has taken too long to be verified, which means it was probably lost and
			// will need to be retransmitted
			if time.Now().UnixMilli()-p.Timestamp > 200 {
				// it has been 200ms, retransmit the packet and reset the timestamp (use the direct write to the UDP socket)
				conn.server.conn.WriteToUDPAddrPort(p.Data, conn.addr)
			}

			// this packet hasn't been verified, move on to check the next one
			i++
		}

	}
}
