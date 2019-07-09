// schedissue illustrates an apparent defect in Go's runtime where on
// a multi-core machine, as little as one goroutine doing CPU
// intensive work can block other goroutines from executing.
//
// It first starts a configurable number of goroutines which are
// making periodic HTTP requests, then sleeps for a 4 seconds in order
// to allow a few requests to execute. It then launches a configurable
// number of 'worker' goroutines which are generating and XORing
// random numbers together in a tight loop.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/jaffee/commandeer"
)

type Main struct {
	Concurrency int           `help:"Number of goroutines to spawn."`
	Iterations  uint64        `help:"Number of times each routine should loop."`
	GoSched     bool          `help:"insert gosched calls into cpu loop"`
	HTTPCount   int           `help:"number of goroutines making periodic HTTP requests"`
	HTTPSleep   time.Duration `help:"sleep duration between HTTP requests"`
}

func NewMain() *Main {
	return &Main{
		Concurrency: runtime.NumCPU(),
		Iterations:  1 << 40,
		HTTPCount:   1,
		HTTPSleep:   time.Millisecond * 100,
	}
}

func (m *Main) Run() error {
	for i := 0; i < m.HTTPCount; i++ {
		go func() {
			for {
				fmt.Printf("1")
				resp, err := http.Get("http://google.com")
				if err != nil {
					log.Printf("making http request: %v", err)
					continue
				}
				if resp.StatusCode != 200 {
					log.Println(resp.StatusCode)
				}
				fmt.Printf("2")
				bod, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Printf("err reading body: %v", err)
				}
				if len(bod) > 100000 {
					log.Println("huge response body: %d", len(bod))
				}
				fmt.Printf("3")
				time.Sleep(m.HTTPSleep)
			}
		}()
	}

	time.Sleep(time.Second * 4)
	log.Println("starting workers")
	wg := sync.WaitGroup{}
	vals := make([]int, m.Concurrency)
	for i := 0; i < m.Concurrency; i++ {
		vals[i] = rand.Int()
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rnd := rand.New(rand.NewSource(rand.Int63()))
			for j := uint64(0); j < m.Iterations; j++ {
				if m.GoSched && j%10000000 == 0 {
					runtime.Gosched()
				}
				vals[idx] = vals[idx] ^ rnd.Int()
			}
		}(i)
	}
	wg.Wait()
	log.Println(vals) // "use" vals so compiler can't optimize computation away.
	return nil
}

func main() {
	if err := commandeer.Run(NewMain()); err != nil {
		log.Fatal(err)
	}
}
