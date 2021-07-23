package chord

import (
	"net"
	"net/rpc"
	"strconv"
	"time"
)

const (
	CheckWaitTime = 50 * time.Millisecond
)

type ChordNode struct {
	N Node
	Server *rpc.Server
	Listener net.Listener
}

func (c *ChordNode) Init(port int) {
	c.N.Self.Addr = GetLocalAddress() + ":" + strconv.Itoa(port)
	c.N.Self.HashAddr = SHA1(c.N.Self.Addr)
	c.N.Data = make(map[string]string)
	c.N.PreData = make(map[string]string)
	c.N.Predecessor = nil
	c.Server = rpc.NewServer()
	_ = c.Server.Register(&c.N)
}

func (c *ChordNode) Run() {
	listener , _ := net.Listen("tcp" , c.N.Self.Addr)
	c.Listener = listener
	go c.Server.Accept(c.Listener)
	c.N.Run = true
}

func (c *ChordNode) Create() {
	c.N.SuccessorList[1] = c.N.Self
	c.N.Predecessor = nil
	go func(){
		for c.N.Run {
			c.N.StabilizeLock.Lock()
			c.N.stabilize()
			c.N.StabilizeLock.Unlock()
			time.Sleep(CheckWaitTime)
		}
	}()
	go func(){
		for c.N.Run {
			c.N.checkPre()
			time.Sleep(CheckWaitTime)
		}
	}()
	go func(){
		next := 1
		for c.N.Run {
			c.N.FingerTableLock.Lock()
			c.N.fixFingerTable(&next)
			c.N.FingerTableLock.Unlock()
			time.Sleep(CheckWaitTime)
		}
	}()
}

func (c *ChordNode) Join(addr string) bool { //***只有dial失败时返回false
	c.N.Predecessor = nil
	cli , err := dial(addr)
	if err != nil {
		return false
	}
	var successor IP
	err = cli.Call("Node.GetTargetSuccessor" , c.N.Self.HashAddr , &successor)
	c.N.SuccessorList[1] = successor
	if err != nil {
		return c.Join(addr)
	}
	cli.Close()
	cli , err = dial(c.N.getFirstValidSuccessor().Addr)
	if err != nil {
		return c.Join(addr)
	} else {
		defer cli.Close()
	}
	var Map map[string]string
	err = cli.Call("Node.GetMap" , 0 , &Map)
	if err != nil {
		return c.Join(addr)
	}
	var Pre IP
	err = cli.Call("Node.GetPre" , 0 , &Pre)
	if err != nil {
		return c.Join(addr)
	}
	c.N.DataLock.Lock()
	for k , v := range Map {
		var hash = SHA1(k)
		if between(Pre.HashAddr , hash , c.N.Self.HashAddr , true) {
			c.N.Data[k] = v
		}
	}
	c.N.DataLock.Unlock()
	err = cli.Call("Node.DataMove" , c.N.Self , nil)
	if err != nil {
		return c.Join(addr)
	}
	err = cli.Call("Node.Notify" , c.N.Self , nil)
	if err != nil {
		return c.Join(addr)
	}
	go func(){
		for c.N.Run {
			c.N.StabilizeLock.Lock()
			c.N.stabilize()
			c.N.StabilizeLock.Unlock()
			time.Sleep(CheckWaitTime)
		}
	}()
	go func(){
		for c.N.Run {
			c.N.checkPre()
			time.Sleep(CheckWaitTime)
		}
	}()
	go func(){
		next := 1
		for c.N.Run {
			c.N.FingerTableLock.Lock()
			c.N.fixFingerTable(&next)
			c.N.FingerTableLock.Unlock()
			time.Sleep(CheckWaitTime)
		}
	}()
	return true
}

func (c *ChordNode) Quit() {
	if c.N.Run{
		cli , err := dial(c.N.getFirstValidSuccessor().Addr)
		if err == nil {
			err = cli.Call("Node.AddDataMap" , &c.N.Data , nil)
			cli.Close()
		}
		c.N.Run = false
		c.Listener.Close()
	}
}

func (c *ChordNode) ForceQuit() {
	if c.N.Run {
		c.N.Run = false
		_ = c.Listener.Close()
	}
}

func (c *ChordNode) Ping(addr string) bool {
	return ping(addr)
}

func (c *ChordNode) Put(key string , value string) bool {
	hash := SHA1(key)
	var successor IP
	_err := c.N.GetTargetSuccessor(hash , &successor)
	if _err != nil {
		return false
	}
	cli , err := dial(successor.Addr)
	if err != nil {
		return false
	} else {
		defer cli.Close()
	}
	err = cli.Call("Node.PutData" , &KVpair{key , value} , nil)
	if err != nil {
		return false
	} else {
		return true
	}
}

func (c *ChordNode) Get(key string) (bool , string) {
	hash := SHA1(key)
	var successor IP
	_err := c.N.GetTargetSuccessor(hash , &successor)
	if _err != nil {
		return false , ""
	}
	cli , err := dial(successor.Addr)
	if err != nil {
		return false , ""
	} else {
		defer cli.Close()
	}
	var val string
	err = cli.Call("Node.GetData" , &key , &val)
	if err != nil {
		return false , ""
	}
	return true , val
}

func (c *ChordNode) Delete(key string) bool {
	hash := SHA1(key)
	var successor IP
	_err := c.N.GetTargetSuccessor(hash , &successor)
	if _err != nil {
		return false
	}
	cli , err := dial(successor.Addr)
	if err != nil {
		return false
	} else {
		defer cli.Close()
	}
	err = cli.Call("Node.DeleteData" , &key , nil)
	if err != nil {
		return false
	} else {
		return true
	}
}