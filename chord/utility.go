package chord

import (
	"crypto/sha1"
	"math/big"
	"net"
	"net/rpc"
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

func between(begin *big.Int , mid *big.Int , end *big.Int , in bool) bool {
	if end.Cmp(begin) > 0 {
		return (begin.Cmp(mid) < 0 && mid.Cmp(end) < 0) || (in && mid.Cmp(end) == 0)
	} else {
		return begin.Cmp(mid) < 0 || mid.Cmp(end) < 0 || (in && mid.Cmp(end) == 0)
	}
}

func jump(addr *big.Int , next int) *big.Int {
	exp := new(big.Int).Exp(big.NewInt(2) , big.NewInt(int64(next) - 1) , nil)
	ans := new(big.Int).Add(addr , exp)
	return new(big.Int).Mod(ans , hashMod)
}

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