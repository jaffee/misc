// schedmin illustrates an apparent defect in Go's runtime where on a
// multi-core machine, as little as one goroutine doing CPU intensive
// work can block other goroutines from executing.
//
// It first starts a goroutine which is making periodic HTTP requests,
// then sleeps for a 4 seconds in order to allow a few requests to
// execute. It then generates and XORs random numbers together in a
// tight loop. Periodically invoking runtime.Gosched() in the work
// loop alleviates the issue.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var goSched = false

func main() {
	// start HTTP request goroutine
	go func() {
		for {
			netLoopBody()
			// sleepLoopBody() // does not reproduce issue
		}
	}()

	// allow HTTP goroutine to execute for 4 seconds
	time.Sleep(time.Second * 4)

	// start "work"
	log.Println("starting work")
	val := 0
	rnd := rand.New(rand.NewSource(rand.Int63()))
	for j := uint64(0); j < 1<<40; j++ {
		if goSched && j%10000000 == 0 {
			runtime.Gosched()
		}
		val = val ^ rnd.Int()
	}
	log.Println(val) // "use" val so compiler can't optimize computation away.
}

func netLoopBody() {
	fmt.Printf("1")
	resp, err := http.Get("http://golang.org/")
	if err != nil {
		log.Printf("making http request: %v", err)
		time.Sleep(time.Millisecond * 500)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println(resp.StatusCode)
	}
	fmt.Printf("2")
	bod, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("err reading body: %v", err)
	}
	if len(bod) > 400000 {
		log.Println("huge response body: %d", len(bod))
	}
	fmt.Printf("3")
	time.Sleep(time.Millisecond * 300)
}

// does not reproduce the issue
func sleepLoopBody() {
	fmt.Printf("1")
	time.Sleep(time.Millisecond)
	fmt.Printf("2")
	time.Sleep(time.Millisecond * 2)
	fmt.Printf("3")
	time.Sleep(time.Millisecond * 100)
}
