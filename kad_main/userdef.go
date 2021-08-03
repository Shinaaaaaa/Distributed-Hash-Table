package main

import (
	"kad"
)

func NewNode(port int) dhtNode {
	var client kad.KadNode
	client.Init(port)
	var res = &client
	return res
}
