package main

import (
	"time"

	"github.com/jomstead/wildspace/server/wildnet"
	"github.com/jomstead/wildspace/server/wildworld"
)

var net wildnet.Net
var world wildworld.World

func main() {
	println("Starting wildnet networking")
	net = wildnet.Net{}
	go net.Init() // initializes the net listener AND creates a go routine listening for packets

	println("Starting wildnet world")
	world = wildworld.World{}
	world.Init(&net) // initialize the world

	// start the game loop
	println("Game loop starting")
	for range time.Tick(time.Millisecond * 500) { // 20 updates per second
		world.Update()
	}
}
