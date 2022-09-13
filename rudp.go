package rudp

import (
	"errors"
	"net"
	"strconv"

	"github.com/jomstead/go-rudp/client"
	"github.com/jomstead/go-rudp/server"
)

/*
*	RUDP - Reliable UDP
*	Packet Structure  [Reliable Flag][Sequence number][remote ack][remote bitwise][Payload]
*		Reliable flag - 0 for unreliable, 1 for reliable
*		Sequence number - if reliable then a unique sequencial number is added to each packet
*		Remote Ack - the last received sequence number from the remote connection
*		Remote bitwise - acks for the last 32 remote packets
*		Payload - User provided payload
 */

func Listen(network string, host string, port uint16) (*server.RUDPServer, error) {
	switch network {
	case "udp", "udp4", "udp6":
	default:
		return nil, errors.New("only udp, udp4, and udp6 network types accepted")
	}
	address := host + ":" + strconv.Itoa(int(port))
	s, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	c, err := net.ListenUDP(network, s)
	if err != nil {
		return nil, err
	}
	rudpconn := server.RUDPServer{}
	rudpconn.Initialize(c, s)
	return &rudpconn, nil
}

func Dial(network string, host string, port uint16) (*client.RUDPClient, error) {
	switch network {
	case "udp", "udp4", "udp6":
	default:
		return nil, errors.New("only udp, udp4, and udp6 network types accepted")
	}
	address := host + ":" + strconv.Itoa(int(port))
	s, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP(network, nil, s)
	if err != nil {
		return nil, err
	}
	rudpclient := client.RUDPClient{}
	rudpclient.Initialize(c, s)
	return &rudpclient, nil
}
