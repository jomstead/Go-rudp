## Go-rudp - Reliable UDP

## Features
Send reliable and unreliable udp packets.  Client and server provide verification of received reliable packets.

## How does it work?

---
reliable: [0 or 1]
sequence: [uint32]
remote_ack: [uint32]
remote_bitfield: [uint32]
payload: [user data]
---