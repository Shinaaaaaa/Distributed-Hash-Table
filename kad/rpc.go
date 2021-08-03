package kad

import (
	"errors"
	"math/big"
	"net/rpc"
	"sort"
)

func ping(addr string) bool { //***多次ping确保
	for i := 0 ; i < tryTime ; i++ {
		c , err := rpc.Dial("tcp" , addr)
		if err != nil {
			continue
		} else{
			defer c.Close()
			var run bool
			_ = c.Call("Node.CheckRun" , 0 , &run)
			if run {
				return true
			} else {
				return false
			}
		}
	}
	//fmt.Println(addr + "ping fail")
	return false
}

func (n *Node) Store(kvPair KVpair , _ *int) error {
	n.DataLock.Lock()
	n.Data[kvPair.Key] = kvPair.Value
	n.DataLock.Unlock()
	return nil
}

func (n *Node) FindNode(target *big.Int , container *[K]IP) error{
	var con [K]IP
	pos := getBucket(n.Self.HashAddr , target)
	if pos == -1 {
		pos = 0
	}
	var cnt = 0

	if n.K_cnt <= K {
		for i := 0 ; i < bucketNum ; i++ {
			n.K_bucket[i].Lock.Lock()
			for j := 0 ; j < K ; j++ {
				if n.K_bucket[i].Inf[j].Addr == "" {
					break
				}
				if n.K_bucket[i].Inf[j].Addr != "" && ping(n.K_bucket[i].Inf[j].Addr) {
					con[cnt] = n.K_bucket[i].Inf[j]
					cnt++
				}
			}
			n.K_bucket[i].Lock.Unlock()
		}
	} else {
		n.K_bucket[pos].Lock.Lock()
		for i := 0 ; i < K && cnt < K ; i++{
			if n.K_bucket[pos].Inf[i].Addr != "" && ping(n.K_bucket[pos].Inf[i].Addr) {
				con[cnt] = n.K_bucket[pos].Inf[i]
				cnt++
			}
		}
		n.K_bucket[pos].Lock.Unlock()
		if cnt < K {
			var i int
			for i = 1 ; pos + i < bucketNum && pos - i >= 0  && cnt < K ; i++ {
				n.K_bucket[pos + i].Lock.Lock()
				n.K_bucket[pos - i].Lock.Lock()
				var list MergeList
				for j := 0 ; j < K ; j++ {
					if n.K_bucket[pos + i].Inf[j].Addr != "" && ping(n.K_bucket[pos + i].Inf[j].Addr) {
						list = append(list , MergeType{n.K_bucket[pos + i].Inf[j] , getDis(n.K_bucket[pos + i].Inf[j].HashAddr , target)})
					}
					if n.K_bucket[pos - i].Inf[j].Addr != "" && ping(n.K_bucket[pos - i].Inf[j].Addr)  {
						list = append(list , MergeType{n.K_bucket[pos - i].Inf[j] , getDis(n.K_bucket[pos - i].Inf[j].HashAddr , target)})
					}
				}
				sort.Sort(list)
				for j := 0 ; j < len(list) && cnt < K ; j++ {
					con[cnt] = list[j].addr
					cnt++
				}
				n.K_bucket[pos + i].Lock.Unlock()
				n.K_bucket[pos - i].Lock.Unlock()
			}
			for ; pos + i < bucketNum && cnt < K ; i++ {
				n.K_bucket[pos + i].Lock.Lock()
				var list MergeList
				for j := 0 ; j < K ; j++ {
					if n.K_bucket[pos + i].Inf[j].Addr != "" && ping(n.K_bucket[pos + i].Inf[j].Addr) {
						list = append(list , MergeType{n.K_bucket[pos + i].Inf[j] , getDis(n.K_bucket[pos + i].Inf[j].HashAddr , target)})
					}
				}
				sort.Sort(list)
				for j := 0 ; j < len(list) && cnt < K ; j++ {
					con[cnt] = list[j].addr
					cnt++
				}
				n.K_bucket[pos + i].Lock.Unlock()
			}
			for ; pos - i >= 0 && cnt < K ; i++ {
				n.K_bucket[pos - i].Lock.Lock()
				var list MergeList
				for j := 0 ; j < K ; j++ {
					if n.K_bucket[pos - i].Inf[j].Addr != ""  && ping(n.K_bucket[pos - i].Inf[j].Addr) {
						list = append(list , MergeType{n.K_bucket[pos - i].Inf[j] , getDis(n.K_bucket[pos - i].Inf[j].HashAddr , target)})
					}
				}
				sort.Sort(list)
				for j := 0 ; j < len(list) && cnt < K ; j++ {
					con[cnt] = list[j].addr
					cnt++
				}
				n.K_bucket[pos - i].Lock.Unlock()
			}
		}
	}
	*container = con
	return nil
}

func (n *Node) FindValue(key string , value *string) error {
	val , ok := n.Data[key]
	if !ok {
		*value = ""
		return errors.New("no find")
	} else {
		*value = val
		return nil
	}
}