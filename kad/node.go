package kad

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

const (
	K = 20
	bucketNum = 160
	WaitTime = 500 * time.Millisecond
	StoreTime = 10 * time.Second
)

type IP struct {
	Addr     string
	HashAddr *big.Int
}

type MergeType struct {
	addr IP
	dis *big.Int
}

type MergeList []MergeType

func (L MergeList) Len() int {
	return len(L)
}

func (L MergeList) Swap(i , j int) {
	L[i] , L[j] = L[j] , L[i]
}

func (L MergeList) Less(i , j int) bool {
	return L[i].dis.Cmp(L[j].dis) < 0
}

type KVpair struct {
	Key string
	Value string
}

type bucket struct {
	Inf [K]IP
	Lock sync.Mutex
}

func (b *bucket) addBucket(n IP) bool {
	b.Lock.Lock()
	var i int
	flag := true
	for i = 0 ; i < K ; i++ {
		if b.Inf[i].Addr == n.Addr {
			flag = false
			break
		}
	}
	if i != K {
		for j := i ; j < K - 1 ; j++ {
			b.Inf[j] = b.Inf[j + 1]
		}
		b.Inf[K - 1] = IP{}
		for j := 0 ; j < K ; j++{
			if b.Inf[j].Addr == "" {
				b.Inf[j] = n
				break
			}
		}
	} else {
		if b.Inf[K - 1].Addr != "" {
			if b.Inf[0].Addr != "" && ping(b.Inf[0].Addr){
				n = b.Inf[0]
			}
			for j := 0 ; j < K - 1 ; j++ {
				b.Inf[j] = b.Inf[j + 1]
			}
			b.Inf[K - 1] = IP{}
			for j := 0 ; j < K ; j++{
				if b.Inf[j].Addr == "" {
					b.Inf[j] = n
					break
				}
			}
		} else {
			for j := 0 ; j < K ; j++{
				if b.Inf[j].Addr == "" {
					b.Inf[j] = n
					break
				}
			}
		}
	}
	b.Lock.Unlock()
	return flag
}

type data struct {
	D map[string]string
	T map[string]time.Time
}

type Node struct {
	Self     IP
	K_bucket [bucketNum]bucket
	K_cnt    int
	Data     data
	DataLock sync.Mutex
	DataTimeLock sync.Mutex
	InNet    bool
}

func (n *Node) CheckRun(_ int , run *bool) error {
	*run = n.InNet
	return nil
}

func (n *Node) Refresh(key string , _ *int) error {
	_ , ok := n.Data.D[key]
	if !ok {
		return errors.New("no find")
	} else {
		n.DataTimeLock.Lock()
		n.Data.T[key] = time.Now()
		n.DataTimeLock.Unlock()
	}
	return nil
}

func (n *Node) AddNodeInBucket(addr IP , _ *int) error {
	if n.Self.Addr == addr.Addr {
		return errors.New("same")
	}
	ok := n.K_bucket[getBucket(n.Self.HashAddr , addr.HashAddr)].addBucket(addr)
	if ok {
		n.K_cnt++
	}
	return nil
}

func (n *Node) nearestNode(target *big.Int) [K]IP {
	var con [K]IP
	var visit map[string]int
	visit = make(map[string]int)
	_ = n.FindNode(target , &con)
	for true {
		var convis = false
		for i := 0 ; i < K && con[i].Addr != "" ; i++ {
			_ , ok := visit[con[i].Addr]
			if !ok {
				visit[con[i].Addr] = 1
				cli , err := dial(con[i].Addr)
				if err != nil {
					continue
				} else {
					var tmp [K]IP
					err = cli.Call("Node.FindNode" , target , &tmp)
					_ = cli.Call("Node.AddNodeInBucket" , n.Self , nil)
					_ = cli.Close()
					if err == nil {
						con = merge(con , tmp , target)
					}
				}
				convis = true
				break
			}
		}
		if !convis {
			break
		}
	}
	return con
}

func (n *Node) query(key string) string {
	target := SHA1(key)
	cli , err := dial(n.Self.Addr)
	if err != nil {
		return ""
	}
	var value string
	err = cli.Call("Node.FindValue" , key , &value)
	if err == nil  {
		n.Refresh(key , nil)
		return value
	}
	value = ""
	con := n.nearestNode(target)
	for i := 0 ; i < K && con[i].Addr != "" ; i++ {
		var val string
		cli , err = dial(con[i].Addr)
		if err != nil {
			continue
		} else {
			err = cli.Call("Node.FindValue" , key , &val)
			if err != nil {
				continue
			} else {
				value = val
				break
			}
		}
	}
	for i := 0 ; i < K && con[i].Addr != "" ; i++ {
		cli , err = dial(con[i].Addr)
		if err != nil {
			continue
		} else {
			err = cli.Call("Node.Refresh" , key , nil)
			if err != nil {
				_ = cli.Call("Node.Store" , KVpair{key , value} , nil)
			}
		}
	}
	return value
}

func (n *Node) StoreData(kvPair KVpair , _ *int) error {
	con := n.nearestNode(SHA1(kvPair.Key))
	for i := 0 ; i < K && con[i].Addr != "" ; i++ {
		cli , err := dial(con[i].Addr)
		if err != nil {
			continue
		} else {
			_ = cli.Call("Node.Store" , kvPair , nil)
			_ = cli.Call("Node.AddNodeInBucket" , n.Self , nil)
			_ = cli.Close()
		}
	}
	return nil
}

func (n *Node) JoinNet(addr string , _ *int) error {
	if !ping(addr) {
		return errors.New("conn fail")
	} else {
		ok := n.K_bucket[getBucket(n.Self.HashAddr , SHA1(addr))].addBucket(IP{addr , SHA1(addr)})
		if ok {
			n.K_cnt++
		}
		container := n.nearestNode(n.Self.HashAddr)
		for i := 0 ; i < K ; i++ {
			if container[i].Addr == "" {
				break
			}
			if container[i].Addr != n.Self.Addr {
				ok = n.K_bucket[getBucket(n.Self.HashAddr , container[i].HashAddr)].addBucket(container[i])
				if ok {
					n.K_cnt++
				}
				cli , err := dial(container[i].Addr)
				if err != nil {
					continue
				} else {
					_ = cli.Call("Node.AddNodeInBucket" , n.Self , nil)
					cli.Close()
				}
			}
		}
	}
	return nil
}

func (n *Node) releaseData() {
	if n.InNet {
		n.DataLock.Lock()
		tmp := n.Data.D
		n.DataLock.Unlock()
		for k , v := range tmp {
			con := n.nearestNode(SHA1(k))
			for i := 0 ; i < K && con[i].Addr != "" ; i++ {
				cli , err := dial(con[i].Addr)
				if err != nil {
					return
				} else {
					_ = cli.Call("Node.Store" , KVpair{k , v} , nil)
					_ = cli.Call("Node.AddNodeInBucket" , n.Self , nil)
					_ = cli.Close()
				}
			}
		}
		time.Sleep(WaitTime)
	}
}

func (n *Node) checkDataTime() {
	ticker := time.Tick(time.Second)
	for range ticker {
		n.DataTimeLock.Lock()
		TimeMap := n.Data.T
		n.DataTimeLock.Unlock()
		for k , t := range TimeMap {
			now := time.Now()
			if t.Add(StoreTime).Before(now) {
				n.DataLock.Lock()
				n.DataTimeLock.Lock()
				delete(n.Data.D , k)
				delete(n.Data.T , k)
				n.DataLock.Unlock()
				n.DataTimeLock.Unlock()
			}
		}
	}
}