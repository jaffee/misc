// package main tests out zipf. Try misc -v 50 -s 1.6 -num 100000 -max 1000 | hist  -b 100
package main

import (
	"fmt"
	"math/rand"

	"github.com/jaffee/commandeer"
)

type Main struct {
	S   float64
	V   float64
	Max uint64
	Num int
}

func NewMain() *Main {
	return &Main{
		S:   1.1,
		V:   1024,
		Max: 100,
		Num: 100,
	}
}

func (m *Main) Run() error {
	s := rand.NewSource(1)
	r := rand.New(s)
	z := rand.NewZipf(r, m.S, m.V, m.Max)
	for i := 0; i < m.Num; i++ {
		fmt.Println(z.Uint64())
	}
	return nil
}

func main() {
	err := commandeer.Run(NewMain())
	if err != nil {
		fmt.Println("running", err)
	}
}
