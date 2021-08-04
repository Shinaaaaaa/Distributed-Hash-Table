package kad

import (
	"crypto/sha1"
	"math/big"
	"net"
	"net/rpc"
	"sort"
	"time"
)

const (
	tryTime  = 3
	waitTime = 500 * time.Millisecond
)
var hashMod = new(big.Int).Exp(big.NewInt(2) , big.NewInt(160) , nil)

func SHA1(addr string) *big.Int {
	hash := sha1.New()
	hash.Write([]byte(addr))
	return new(big.Int).SetBytes(hash.Sum(nil))
}

func GetLocalAddress() string {
	var localaddress string
	ifaces, err := net.Interfaces()
	if err != nil {
		panic("init: failed to find network interfaces")
	}
	for _, elt := range ifaces {
		if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
			addrs, err := elt.Addrs()
			if err != nil {
				panic("init: failed to get addresses for network interface")
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
						localaddress = ip4.String()
						break
					}
				}
			}
		}
	}
	if localaddress == "" {
		panic("init: failed to find non-loopback interface with valid address on this node")
	}
	return localaddress
}

func getDis(a *big.Int , b *big.Int) *big.Int {
	tmp := new(big.Int)
	tmp.Xor(a , b)
	return tmp
}

func getBucket(a *big.Int , b *big.Int) int {
	tmp := new(big.Int)
	tmp.Xor(a , b)
	return tmp.BitLen() - 1;
}

func merge(a [K]IP , b [K]IP , target *big.Int) [K]IP{
	var con [K]IP
	var tmp [K]IP
	cnt := 0
	for i := 0 ; i < K ; i++ {
		flag := false
		for j := 0 ; j < K ; j++ {
			if a[j].Addr == b[i].Addr {
				flag = true
				break
			}
		}
		if !flag {
			tmp[cnt] = b[i]
			cnt++
		}
	}
	b = tmp

	var list MergeList
	for i := 0 ; i < K ; i++ {
		if a[i].Addr != "" {
			list = append(list , MergeType{a[i] , getDis(a[i].HashAddr , target)})
		}
		if b[i].Addr != "" {
			list = append(list , MergeType{b[i] , getDis(b[i].HashAddr , target)})
		}
	}
	sort.Sort(list)
	for i := 0 ; i < len(list) && i < K ; i++ {
		con[i] = list[i].addr
	}
	return con
}

func dial(addr string) (*rpc.Client , error) { //***多次dial
	var err error
	var c *rpc.Client
	for i := 0 ; i < tryTime; i++{
		c , err = rpc.Dial("tcp" , addr)
		if err == nil{
			return c , err
		} else {
			time.Sleep(waitTime)
		}
	}
	return nil , err
}