# Go-rudp - Reliable UDP

Go-rudp is a reliable udp implementation for use in multiplayer games written in golang.

## Features
Send reliable and unreliable udp packets.  Client and server provide verification of received reliable packets.

## How does it work?

Packet
[reliable][sequence][remote_ack][remote_bitfield][payload]

Go-rupd adds additional packet information to all outgoing packets.
- reliable[uint8]: 0 or 1 - if set to 1 the packet is reliable.
- seqeunce[uint32]: an incremental sequence number is assigned to each **reliable** packet.
- remote_ack[uint32]: The last reliable packet received from the remote source.
- remote_bitfield[uint32]: A bitfield used to acknowledge the receiption of up to the past 32 packets from the remote source. 1=received, 0=not received.
- payload: the data the user is sending.

When a reliable packet is received, the remote_ack is updated with the sequence number if newer than the current value (sometimes udp receives out of order so it may be an older sequence number).  Then the remote_bitfield is updated using some bit shifting and bit setting.  Then only the payload data is passed through.

## How to use the library

Server.go
```Go
 	server, _ := rudp.Listen("udp4", "127.0.0.1", 8000)
	defer server.Close()

    // receiving a packet
    // n is the length of the received packet (payload only)
    // verified is a list of reliable packets that the client has received since the last read
    temp := make([]byte, 1024)
    n, verified, client_addr, err = server.ReadFromUDP(temp)

    // sending a packet
    // The third argument in WriteToUDP is whether the packet should be reliable 
    // The second return value is the sequence number used for this packet
    payload := []byte{1} // payload to send
    n, sent_seq_number, err := server.WriteToUDP(&payload, *client_addr, true) 

    // When sending and receiving packets, you are responsible for keeping a list of sent packet sequence numbers and removing verified sequences from the list.  See an example implementation in the examples folder.

```

Client.go
```Go
  	client, _ := Dial("udp4", "127.0.0.1", 8000)
	defer client.Close()

    // receiving a packet
    // n is the length of the received packet (payload only)
    // verified is a list of reliable packets that the remote has received since the last read
    temp := make([]byte, 1024)
    n, verified, server_addr, err = client.ReadFromUDP(temp)

    // sending a packet
    // The third argument in WriteToUDP is whether the packet should be reliable 
    // The second return value is the sequence number used for this packet
    payload := []byte{1} // payload to send
    n, sent_seq_number, err := client.Write(&payload, true) 

    // When sending and receiving packets, you are responsible for keeping a list of sent packet sequence numbers and removing verified sequences from the list.  See an example implementation in the examples folder.

```