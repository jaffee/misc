package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()
		dur, err := time.ParseDuration(text)
		if err != nil {
			fmt.Printf("%s\n", text)
			continue
		}
		fmt.Printf("%d\n", int64(dur))
	}
}
