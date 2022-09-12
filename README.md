# Go-rudp
 Reliable UDP

Features:
Send unreliable packets
Send reliable packets
Retransmit lost reliable packets

Work in progress:
Use average verify time to determine retransmit time
Detect disconnections (no packets received within timeout period, or reliable packet not being verified after several attempts)