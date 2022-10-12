package wildworld

import (
	"fmt"
	"time"

	"github.com/jomstead/wildspace/common/math"
	"github.com/jomstead/wildspace/server/wildnet"
)

type World struct {
	network *wildnet.Net
	players map[uint32]Player
}

type Player struct {
	Position math.Vector2
}

func (world *World) Init(net *wildnet.Net) {
	world.network = net
	world.players = make(map[uint32]Player)
}

func (world *World) GetPlayer(playerID uint32) Player {
	return world.players[playerID]
}

func (world *World) SetPlayer(playerID uint32, player Player) {
	world.players[playerID] = player
}

func (world *World) Update() {
	start := time.Now()
	// TODO: World update function
	// process network packets
	numPackets := len(world.network.Packets)
	for i := 0; i < numPackets; i++ {
		packet := <-world.network.Packets
		world.processPacket(packet)
	}
	packetTime := time.Since(start)

	// update positions
	nano := '\u03BC'
	fmt.Printf("Update - Processed %d packets(%v%cs)\n", numPackets, packetTime.Nanoseconds(), nano)

}

func (world *World) processPacket(packet *wildnet.Packet) {
	command := wildnet.Command(packet.Data[0])
	switch command {
	case wildnet.SCAN:
	case wildnet.SCAN_TARGET:
	case wildnet.MOVE:
	default:
	}
}
