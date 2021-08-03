package chord

import (
	"errors"
	"math/big"
	"sync"
)

const (
	M                  = 160
	successorListLen   = 20
)

type IP struct {
	Addr     string
	HashAddr *big.Int
}

type KVpair struct {
	Key   string
	Value string
}

type Node struct {
	Self              IP
	Data              map[string]string
	SuccessorList     [successorListLen + 1]IP
	FingerTable       [M + 1]IP
	Predecessor       *IP
	InNet             bool
	PreData           map[string]string
	DataLock          sync.Mutex
	PreDataLock       sync.Mutex
	SuccessorListLock sync.Mutex
	StabilizeLock     sync.Mutex
	FingerTableLock   sync.Mutex
}

func (n *Node) AddPreDataMap(Map map[string]string , _ *int) error {
	n.PreDataLock.Lock()
	for k , v := range Map {
		n.PreData[k] = v
	}
	n.PreDataLock.Unlock()
	return nil
}

func (n *Node) AddDataMap(Map *map[string]string , _ *int) error {
	n.DataLock.Lock()
	for k , v := range *Map {
		n.Data[k] = v
	}
	n.DataLock.Unlock()
	return nil
}

func (n *Node) GetMap(_ int , Map *map[string]string) error {
	n.DataLock.Lock()
	*Map = n.Data
	n.DataLock.Unlock()
	return nil
}

func (n *Node) GetPre(_ int , Pre *IP) error {
	if n.Predecessor == nil {
		//fmt.Println("find no pre")
		return errors.New("no find")
	} else {
		*Pre = *n.Predecessor
		return nil
	}
}

func (n *Node) GetSuccessorList(_ int , successorList *[successorListLen + 1]IP) error {
	*successorList = n.SuccessorList
	return nil
}

func (n *Node) fixSuccessorList(pos int) error{
	client , err := dial(n.SuccessorList[pos].Addr)
	if err != nil {
		return err
	} else {
		defer client.Close()
	}
	var newSuccessors [successorListLen + 1]IP
	_ = client.Call("Node.GetSuccessorList" , 0 , &newSuccessors)
	n.SuccessorList[1] = n.SuccessorList[pos]
	for j := 2 ; j <= successorListLen ; j++ {
		n.SuccessorList[j] = newSuccessors[j - 1]
	}
	return nil
}

func (n *Node) getFirstValidSuccessor() IP { //***只有当全部断线时返回false，否则一直查找
	var i int
	n.SuccessorListLock.Lock()
	for i = 1 ; i <= successorListLen ; i++ {
		if ping(n.SuccessorList[i].Addr) {
			break
		}
	}
	if i != 1 { //发现存在未拨通点时，进行successorlist的修复
		if i == successorListLen + 1 {
			n.SuccessorListLock.Unlock()
			return IP{}
		}
		err := n.fixSuccessorList(i)
		if err != nil {
			n.SuccessorListLock.Unlock()
			return n.getFirstValidSuccessor()
		}
	}
	n.SuccessorListLock.Unlock()
	return n.SuccessorList[1]
}

func (n *Node) getFarthestFingerTable(target *big.Int) IP {
	for i := M ; i >= 1 ; i-- {
		if n.FingerTable[i].HashAddr != nil && between(n.Self.HashAddr , n.FingerTable[i].HashAddr , target , true) && ping(n.FingerTable[i].Addr) {
			return n.FingerTable[i]
		}
	}
	return n.SuccessorList[1]
}

func (n *Node) GetTargetSuccessor(target *big.Int , successor *IP) error {
	validSuccessor := n.getFirstValidSuccessor()
	if validSuccessor.HashAddr == nil {
		*successor = validSuccessor
		//fmt.Println("no target successor")
		return errors.New("no find")
	}
	if between(n.Self.HashAddr , target , validSuccessor.HashAddr , true) {
		*successor = validSuccessor
	} else {
		next := n.getFarthestFingerTable(target)
		client , e := dial(next.Addr)
		if e != nil {
			return n.GetTargetSuccessor(target , successor)
		} else {
			defer client.Close()
		}
		var succ IP
		err := client.Call("Node.GetTargetSuccessor" , target , &succ)
		if err != nil {
			return err
		} else {
			*successor = succ
		}
	}
	return nil
}

func (n *Node) stabilize() {
	validSuccessor := n.getFirstValidSuccessor()
	if validSuccessor.HashAddr == nil {
		return
	}
	c , err := dial(validSuccessor.Addr)
	if err != nil {
		return
	} else {
		defer c.Close()
	}
	var pre IP
	err = c.Call("Node.GetPre" , 0 , &pre)
	if err == nil && ping(pre.Addr) {
		n.SuccessorListLock.Lock()
		if (pre.HashAddr != nil) && between(n.Self.HashAddr , pre.HashAddr , n.SuccessorList[1].HashAddr , false) {
			n.SuccessorList[1] = pre
		}
		c , err = dial(n.SuccessorList[1].Addr)
		n.SuccessorListLock.Unlock()
		if err != nil {
			return
		} else {
			defer c.Close()
		}
	}
	err = c.Call("Node.Notify" , &n.Self , nil)
	var succSuccessors [successorListLen + 1]IP
	err = c.Call("Node.GetSuccessorList" , 0 , &succSuccessors)
	if err != nil {
		return
	}
	n.SuccessorListLock.Lock()
	for i := 2 ; i <= successorListLen ; i++ {
		n.SuccessorList[i] = succSuccessors[i - 1]
	}
	n.SuccessorListLock.Unlock()
}

func (n *Node) Notify(pre IP , _ *int) error {
	if n.Predecessor == nil || between(n.Predecessor.HashAddr , pre.HashAddr , n.Self.HashAddr , false) {
		n.Predecessor = new(IP)
		*n.Predecessor = pre
		c , err := dial(pre.Addr)
		if err == nil {
			n.PreDataLock.Lock()
			newMap := make(map[string]string)
			_ = c.Call("Node.GetMap" , 0 , &newMap)
			n.PreData = newMap
			n.PreDataLock.Unlock()
			c.Close()
		}
		return nil
	}
	return nil
}

func (n *Node) checkPre() {
	if n.Predecessor != nil {
		if !ping(n.Predecessor.Addr){
			n.Predecessor = nil
			n.PreDataLock.Lock()
			n.DataLock.Lock()
			for k , v := range n.PreData {
				n.Data[k] = v
			}
			c , err := dial(n.getFirstValidSuccessor().Addr)
			if err != nil {
				return
			} else {
				defer c.Close()
			}
			n.DataLock.Unlock()
			_ = c.Call("Node.AddPreDataMap" , n.PreData , nil)
			n.PreData = make(map[string]string)
			n.PreDataLock.Unlock()
		}
	}
}

func (n *Node) fixFingerTable(next *int) {
	err := n.GetTargetSuccessor(jump(n.Self.HashAddr , *next) , &n.FingerTable[*next])
	if err != nil {
		return
	}
	tmp := n.FingerTable[*next]
	*next++
	if *next > M {
		*next = 1
		return
	}
	for { //***一次更新多个FingerTable
		if between(n.Self.HashAddr , jump(n.Self.HashAddr , *next) , tmp.HashAddr , true) {
			n.FingerTable[*next] = tmp
			*next++
			if *next > M {
				*next = 1
				break
			}
		} else {
			break
		}
	}
}

func (n *Node) CheckRun(_ int , run *bool) error {
	*run = n.InNet
	return nil
}

func (n *Node) DataMove(newNode IP , _ *int) error {
	n.DataLock.Lock()
	for k , v := range n.Data {
		if between(n.Predecessor.HashAddr , SHA1(k) , newNode.HashAddr , true) {
			delete(n.Data , v)
		}
	}
	n.DataLock.Unlock()
	return nil
}

func (n *Node) PutPreData(kvPair *KVpair , _ *int) error {
	n.PreDataLock.Lock()
	n.PreData[kvPair.Key] = kvPair.Value
	n.PreDataLock.Unlock()
	return nil
}

func (n *Node) PutData(kvPair *KVpair , _ *int) error {
	c , err := dial(n.getFirstValidSuccessor().Addr)
	if err != nil {
		return err
	} else {
		defer c.Close()
		_ = c.Call("Node.PutPreData" , kvPair , nil)
		n.DataLock.Lock()
		n.Data[kvPair.Key] = kvPair.Value
		n.DataLock.Unlock()
		return nil
	}
}

func (n *Node) GetData(key *string , value *string) error {
	n.DataLock.Lock()
	tmp , ok := n.Data[*key]
	*value = tmp
	n.DataLock.Unlock()
	if !ok {
		return errors.New("no find")
	}
	return nil
}

func (n *Node) DeletePreData(key *string , _ *int) error {
	n.PreDataLock.Lock()
	delete(n.PreData , *key)
	n.PreDataLock.Unlock()
	return nil
}

func (n *Node) DeleteData(key *string , _ *int) error {
	c , err := dial(n.getFirstValidSuccessor().Addr)
	if err != nil {
		return err
	} else {
		defer c.Close()
		_ = c.Call("Node.DeletePreData" , key , nil)
		n.DataLock.Lock()
		_ , ok := n.Data[*key]
		if !ok {
			n.DataLock.Unlock()
			return errors.New("no find")
		} else {
			delete(n.Data , *key)
			n.DataLock.Unlock()
			return nil
		}
	}
}