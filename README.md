# Go-rudp - Reliable UDP

Go-rudp is a reliable udp implementation for use in multiplayer games written in golang.

## Features
Send reliable and unreliable udp packets.  Client and server provide verification of received reliable packets.

## How does it work?

Packet
[reliable][sequence][remote_ack][remote_bitfield][payload]

Go-rupd adds additional packet information to all outgoing packets.
reliable[uint8]: 0 or 1 - if set to 1 the packet is reliable
seqeunce[uint32]: an incremental sequence number is assigned to each **reliable** packet
remote_ack[uint32]: The last reliable packet received from the remote source
remote_bitfield[uint32]: A bitfield used to acknowledge the receiption of up to the past 32 packets from the remote source. 1=received, 0=not received
payload: the data the user is sending

When a reliable packet is received, the remote_ack is updated with the sequence number if newer than the current value (sometimes udp receives out of order so it may be an older sequence number).  Then the remote_bitfield is updated using some bit shifting and bit setting.  Then only the payload data is passed through.

## How to use the library

```Go
Test
```

