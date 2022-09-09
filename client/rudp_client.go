package client

import (
	"encoding/binary"
	"errors"
	"net"
	"time"

	"github.com/jomstead/go-rudp/packet"
)

type RUDPClient struct {
	conn               *net.UDPConn
	address            *net.UDPAddr //host:port
	seq                uint32
	isConnected        bool
	remote_seq         uint32
	remote_acks        packet.Ack
	temp               []byte          // temp is used to read in a packet from the remote source and processed for reliable UDP, it is then copied to a new buffer without the RUDP bytes for processing outside the api
	sent_packet_buffer []packet.Packet // keeps a buffer of sent reliable packets, packets get removed from this slice as they are confirm received
}

func (conn *RUDPClient) Close() {
	if conn.conn != nil {
		conn.conn.Close()
	}
	conn.isConnected = false
}

func (conn RUDPClient) IsConnected() bool {
	return conn.isConnected
}
func (conn *RUDPClient) Initialize(c *net.UDPConn, a *net.UDPAddr) {
	conn.isConnected = true                                // is the client 'connected'
	conn.address = a                                       // address of the remote server
	conn.conn = c                                          // connection to the remote server
	conn.seq = ^uint32(0)                                  //seq number
	conn.remote_seq = ^uint32(0)                           // remote seq number
	conn.sent_packet_buffer = make([]packet.Packet, 0, 16) // queue of outbound reliable packets
	conn.temp = make([]byte, 1024)                         // buffer used for receiving packets
	conn.remote_acks = packet.Ack{Data: 0}

}

/* Write sends a packet to the dialed connection */
func (conn *RUDPClient) Write(payload *[]byte, reliable bool) (int, error) {
	// Create the packet [Reliable][Seq][Remote_seq][remote_acks][Payload]
	var data []byte
	var seq uint32
	index := 0
	if reliable {
		data = make([]byte, 13, len(*payload)+13)
		data[0] = 1
		// increase sequence number for reliable packets
		conn.seq += 1
		seq = conn.seq
		binary.BigEndian.PutUint32(data[1:], conn.seq)
		index = 5
	} else {
		data = make([]byte, 9, len(*payload)+9)
		index = 1
	}
	// include the last received sequence number and the sequence history from the remote source
	binary.BigEndian.PutUint32(data[index:], conn.remote_seq)
	binary.BigEndian.PutUint32(data[index+4:], conn.remote_acks.Data)
	index += 8
	data = append(data, *payload...)
	if reliable {
		// save packets incase we need to resend them
		conn.sent_packet_buffer = append(conn.sent_packet_buffer, packet.Packet{Seq: seq, Data: data, Timestamp: time.Now().UnixMilli()})
	}
	n, err := conn.conn.Write(data)
	return n - index, err
}

func (conn RUDPClient) ReadFromUDP(buffer []byte) (n int, addr *net.UDPAddr, err error) {
	if buffer == nil {
		return 0, nil, errors.New("buffer cannot be nil")
	}
	n, addr, err = conn.conn.ReadFromUDP(conn.temp)
	if err != nil {
		return n, addr, err
	}
	if n > 5 && conn.temp[0] == 0 {
		// unreliable packet
		ack := binary.BigEndian.Uint32(conn.temp[1:5])
		ack_bitfield := binary.BigEndian.Uint32(conn.temp[5:9])
		if len(buffer) < n-9 {
			return 0, nil, errors.New("buffer too small for packet")
		}
		conn.processAck(ack, ack_bitfield)
		copy(buffer, conn.temp[9:n])
		return n - 9, addr, err
	}
	if n > 8 && conn.temp[0] == 1 {
		// reliable packet
		seq := binary.BigEndian.Uint32(conn.temp[1:5])
		conn.remote_seq = packet.UpdateAcknowledgements(seq, conn.remote_seq, &conn.remote_acks)
		ack := binary.BigEndian.Uint32(conn.temp[5:9])
		ack_bitfield := binary.BigEndian.Uint32(conn.temp[9:13])
		if len(buffer) < n-13 {
			return 0, nil, errors.New("buffer too small for packet")
		}
		conn.processAck(ack, ack_bitfield)
		copy(buffer, conn.temp[13:n])
		return n - 13, addr, err
	}
	// Not sure what this packet is....
	return 0, addr, errors.New("unexpected RUDP header data")
}

// ProcessAck takes the acknowledgements from the remote resource and removes packets from the local
// reliable packet buffer that have been confirmed as sent
func (conn *RUDPClient) processAck(seq uint32, bitwise uint32) {
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
				conn.conn.Write(p.Data)
			}

			// this packet hasn't been verified, move on to check the next one
			i++
		}

	}
}
