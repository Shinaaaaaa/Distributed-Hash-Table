package main

import (
	"DHT/chord"
)

func NewNode(port int) dhtNode {
	var client chord.ChordNode
	client.Init(port)
	var res = &client
	return res
}
