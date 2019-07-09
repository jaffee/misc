package main

import (
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/jaffee/commandeer"
	"github.com/pkg/errors"
)

type Main struct {
	Name        string   `help:"memberlist name"`
	Port        int      `help:"port to bind"`
	Seeds       []string `help:"seeds to join"`
	Concurrency int      `help:"Number of goroutines to spawn."`
	Iterations  uint64   `help:"Number of times each routine should loop."`
	GoSched     bool     `help:"insert gosched calls into cpu loop"`
}

func NewMain() *Main {
	return &Main{
		Concurrency: runtime.NumCPU(),
		Iterations:  1 << 40,
	}
}

func (m *Main) Run() error {
	log.Printf("%#v\n", m)
	config := memberlist.DefaultWANConfig()
	config.Name = m.Name
	config.BindPort = m.Port
	config.Delegate = &delegate{name: m.Name}
	config.Events = &eventReceiver{name: m.Name}
	mlist, err := memberlist.Create(config)
	if err != nil {
		return errors.Wrap(err, "creating memberlist")
	}
	_, err = mlist.Join(m.Seeds)
	if err != nil {
		return errors.Wrap(err, "joining memberlist")
	}

	time.Sleep(time.Second * 10)
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

type Command interface {
	Run()
}

type eventReceiver struct {
	name string

	notifyJoin   chan Command
	notifyLeave  chan Command
	notifyUpdate chan Command
}

func (e *eventReceiver) NotifyJoin(n *memberlist.Node) {
	log.Printf("%4s %6s %4s %12s %d", e.name, "Join", n.Name, n.Addr, n.Port)
}

func (e *eventReceiver) NotifyLeave(n *memberlist.Node) {
	log.Printf("%4s %6s %4s %12s %d", e.name, "Leave", n.Name, n.Addr, n.Port)
}

func (e *eventReceiver) NotifyUpdate(n *memberlist.Node) {
	log.Printf("%4s %6s %4s %12s %d", e.name, "Update", n.Name, n.Addr, n.Port)
}

func main() {
	if err := commandeer.Run(NewMain()); err != nil {
		log.Fatal(err)
	}
}

type delegate struct {
	name string
}

func (d *delegate) NodeMeta(limit int) []byte {
	log.Printf("%4s %12s %d", d.name, "NodeMeta", limit)
	return []byte(d.name + "Meta")
}

func (d *delegate) NotifyMsg(msg []byte) {
	log.Printf("%4s %12s %s", d.name, "NotifyMsg", msg)
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	// log.Printf("%4s %12s %d %d", d.name, "GetBroadcasts", overhead, limit)
	return nil
}

func (d *delegate) LocalState(join bool) []byte {
	log.Printf("%4s %12s %v", d.name, "LocalState", join)
	return []byte(d.name)
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	log.Printf("%4s %12s %v %s", d.name, "MergeRemoteState", join, buf)
}
