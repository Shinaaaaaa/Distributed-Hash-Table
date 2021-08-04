package kad

import (
	"net"
	"net/rpc"
	"strconv"
	"time"
)

type KadNode struct {
	N Node
	Server *rpc.Server
	Listener net.Listener
}

func (k *KadNode) Init(port int) {
	k.N.Self.Addr = GetLocalAddress() + ":" + strconv.Itoa(port)
	k.N.Self.HashAddr = SHA1(k.N.Self.Addr)
	k.N.Data.D = make(map[string]string)
	k.N.Data.T = make(map[string]time.Time)
	k.Server = rpc.NewServer()
	_ = k.Server.Register(&k.N)
}

func (k *KadNode) Run() {
	listener , _ := net.Listen("tcp" , k.N.Self.Addr)
	k.Listener = listener
	go k.Server.Accept(k.Listener)
}

func (k *KadNode) Create() {
	//go k.N.releaseData()
	go k.N.checkDataTime()
	k.N.InNet = true
}

func (k *KadNode) Join(addr string) bool {
	err := k.N.JoinNet(addr , nil)
	//go k.N.releaseData()
	go k.N.checkDataTime()
	k.N.InNet = true
	if err != nil {
		return false
	} else {
		return true
	}
}

func (k *KadNode) Quit() {
	if k.N.InNet {
		k.N.InNet = false
		_ = k.Listener.Close()
	}
}

func (k *KadNode) ForceQuit() {
	if k.N.InNet {
		k.N.InNet = false
		_ = k.Listener.Close()
	}
}

func (k *KadNode) Ping(addr string) bool {
	return ping(addr)
}

func (k *KadNode) Put(key string , value string) bool {
	err := k.N.StoreData(KVpair{key , value} , nil)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (k *KadNode) Get(key string) (bool , string) {
	val := k.N.query(key)
	if val == "" {
		return false , ""
	} else {
		return true , val
	}
}