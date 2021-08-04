package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func forceQuitTest() (bool, int, int) {
	_, _ = yellow.Println("Start Force Quit Test")

	forceQuitFailedCnt, forceQuitTotalCnt, panicked := 0, 0, false

	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("Program panicked with", r)
		}
		panicked = true
	}()

	nodes := new([NodeSize + 1]dhtNode)
	nodeAddresses := new([NodeSize + 1]string)
	kvMap := make(map[string]string)
	nodesInNetwork := make([]int, 0, TestNodeSize+1)

	/* InNet all nodes. */
	wg = new(sync.WaitGroup)
	for i := 0; i <= NodeSize; i++ {
		nodes[i] = NewNode(firstPort + i)
		nodeAddresses[i] = portToAddr(localAddress, firstPort+i)

		wg.Add(1)
		go nodes[i].Run()
	}
	time.Sleep(AfterRunSleepTime)

	/* Node 0 creates a new network. All notes join the network. */
	joinInfo := testInfo{
		msg:       "Force quit join",
		failedCnt: 0,
		totalCnt:  0,
	}
	nodes[0].Create()
	nodesInNetwork = append(nodesInNetwork, 0)
	_, _ = cyan.Printf("Start joining\n")
	for i := 1; i <= NodeSize; i++ {
		addr := nodeAddresses[rand.Intn(i)]
		if !nodes[i].Join(addr) {
			joinInfo.fail()
		} else {
			joinInfo.success()
		}
		nodesInNetwork = append(nodesInNetwork, i)

		time.Sleep(JoinSleepTime)
	}
	joinInfo.finish(&forceQuitFailedCnt, &forceQuitTotalCnt)

	time.Sleep(AfterJoinSleepTime)

	/* Put. */
	putInfo := testInfo{
		msg:       "Force quit put",
		failedCnt: 0,
		totalCnt:  0,
	}
	_, _ = cyan.Printf("Start putting\n")
	for i := 0; i < 50; i++ {
		key := randString(lengthOfKeyValue)
		value := randString(lengthOfKeyValue)
		kvMap[key] = value

		if !nodes[rand.Intn(NodeSize+1)].Put(key, value) {
			putInfo.fail()
		} else {
			putInfo.success()
		}
	}
	putInfo.finish(&forceQuitFailedCnt, &forceQuitTotalCnt)

	/* 10 - 1 = 9 rounds in total. */
	for t := 1; t <= RoundNum- 1; t++ {
		_, _ = cyan.Printf("Force Quit Round %d\n", t)

		/* Force quit. */
		_, _ = cyan.Printf("Start force quitting (round %d)\n", t)
		for i := 1; i <= forceQuitRoundQuitNodeSize; i++ {
			idxInArray := rand.Intn(len(nodesInNetwork))

			nodes[nodesInNetwork[idxInArray]].ForceQuit()
			nodesInNetwork = removeFromArray(nodesInNetwork, idxInArray)

			time.Sleep(forceQuitSleepTime)
		}

		/* Get all data. */
		getInfo := testInfo{
			msg:       fmt.Sprintf("Get (round %d)", t),
			failedCnt: 0,
			totalCnt:  0,
		}
		_, _ = cyan.Printf("Start getting (round %d)\n", t)
		cnt := 0
		for key, value := range kvMap {
			ok, res := nodes[nodesInNetwork[rand.Intn(len(nodesInNetwork))]].Get(key)
			if !ok || res != value {
				if res != value {
					cnt++
				}
				getInfo.fail()
			} else {
				getInfo.success()
			}
		}
		fmt.Println(cnt)
		getInfo.finish(&forceQuitFailedCnt, &forceQuitTotalCnt)
	}

	/* All nodes quit. */
	for i := 0; i <= NodeSize; i++ {
		nodes[i].Quit()
	}

	return panicked, forceQuitFailedCnt, forceQuitTotalCnt
}

func myTest1() (int , int) { // findNode
	_, _ = yellow.Println("Start MyTest1")

	nodes := new([50 + 1]dhtNode)
	nodeAddresses := new([50 + 1]string)
	nodesInNetwork := make([]int, 0, TestNodeSize+1)

	/* InNet all nodes. */
	wg = new(sync.WaitGroup)
	for i := 0; i <= 50; i++ {
		nodes[i] = NewNode(firstPort + i)
		nodeAddresses[i] = portToAddr(localAddress, firstPort+i)

		wg.Add(1)
		go nodes[i].Run()
	}
	time.Sleep(AfterRunSleepTime)

	/***** join *****/
	nodes[0].Create()
	nodesInNetwork = append(nodesInNetwork, 0)
	_, _ = cyan.Printf("Start joining\n")
	for i := 1; i <= 50; i++ {
		addr := nodeAddresses[rand.Intn(i)]
		nodes[i].Join(addr)
		nodesInNetwork = append(nodesInNetwork, i)
		time.Sleep(JoinSleepTime)
	}
	_, _ = green.Printf("join complete\n")
	time.Sleep(AfterJoinSleepTime)

	key := randString(lengthOfKeyValue)
	value := randString(lengthOfKeyValue)
	nodes[0].Put(key , value)

	failcnt := 0
	totalcnt := 0
	/***** nearestNode *****/
	_, _ = cyan.Printf("Start get\n")
	for i := 1 ; i <= 50 ; i++ {
		totalcnt++
		ok , v := nodes[i].Get(key)
		if ok && v == value {
			continue
		} else {
			failcnt++
		}
	}
	_, _ = green.Printf("get complete\n")
	return totalcnt , failcnt
}

func myTest2() bool { // findNode
	_, _ = yellow.Println("Start MyTest2")

	nodes := new([10 + 1]dhtNode)
	nodeAddresses := new([10 + 1]string)
	nodesInNetwork := make([]int, 0, TestNodeSize+1)

	/* InNet all nodes. */
	wg = new(sync.WaitGroup)
	for i := 0; i <= 10; i++ {
		nodes[i] = NewNode(firstPort + i)
		nodeAddresses[i] = portToAddr(localAddress, firstPort+i)

		wg.Add(1)
		go nodes[i].Run()
	}
	time.Sleep(AfterRunSleepTime)

	/***** join *****/
	nodes[0].Create()
	nodesInNetwork = append(nodesInNetwork, 0)
	_, _ = cyan.Printf("Start joining\n")
	for i := 1; i <= 10; i++ {
		addr := nodeAddresses[rand.Intn(i)]
		nodes[i].Join(addr)
		nodesInNetwork = append(nodesInNetwork, i)
		time.Sleep(JoinSleepTime)
	}
	_, _ = green.Printf("join complete\n")
	time.Sleep(AfterJoinSleepTime)

	key := randString(lengthOfKeyValue)
	value := randString(lengthOfKeyValue)
	key2 := randString(lengthOfKeyValue)
	value2 := randString(lengthOfKeyValue)
	nodes[0].Put(key , value)
	nodes[0].Put(key2 , value2)

	/***** nearestNode *****/
	_, _ = cyan.Printf("Start get\n")

	ch := make(chan string)
	go func() {
		ticker := time.Tick(time.Second)
		for i := range ticker {
			_ , _ = cyan.Println(i)
			ok , v := nodes[1].Get(key)
			if ok && v == value {
				continue
			} else {
				ch <- v
				break
			}
		}
	}()

	select {
	case res := <- ch:
		_ , _ = red.Println(res)
		return false
	case <- time.After(15 * time.Second):
		ok2 , v2 := nodes[1].Get(key2)
		if ok2 && v2 == value2 {
			_ , _ = red.Println(v2)
			return false
		}
	}
	_, _ = green.Printf("get complete\n")
	return true
}