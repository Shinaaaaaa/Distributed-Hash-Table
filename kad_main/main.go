package main

import (
	"flag"
	"os"
)

func main() {
	_, _ = yellow.Println("Welcome to DHT-2020 Test Program!\n")
	testName := "forcequit"

	switch testName {
	case "forcequit":
		/* ------ Force Quit Test Begins ------ */
		var forceQuitFailRate float64
		forceQuitPanicked, forceQuitFailedCnt, forceQuitTotalCnt := forceQuitTest()
		if forceQuitPanicked {
			_, _ = red.Printf("Force Quit Test Panicked.")
			os.Exit(0)
		}

		forceQuitFailRate = float64(forceQuitFailedCnt) / float64(forceQuitTotalCnt)
		if forceQuitFailRate > forceQuitMaxFailRate {
			_, _ = red.Printf("Force quit test failed with fail rate %.4f\n", forceQuitFailRate)
		} else {
			_, _ = green.Printf("Force quit test passed with fail rate %.4f\n", forceQuitFailRate)
		}
		/* ------ Force Quit Test Ends ------ */

	case "myTest1":
		/* ------ myTest1 Begins ------ */
		totalcnt , failcnt := myTest1()
		FailRate := float64(failcnt) / float64(totalcnt)
		if failcnt > 0 {
			_, _ = red.Printf("myTest1 failed with fail rate %.4f\n", FailRate)
		} else {
			_, _ = green.Printf("myTest1 passed with fail rate %.4f\n", FailRate)
		}
		/* ------ myTest1 Begins ------ */
	}

}

func usage() {
	flag.PrintDefaults()
}
