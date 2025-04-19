package cli

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

var (
	scanner = bufio.NewScanner(os.Stdin)
	writer  = fmt.Printf
)

func Spin() (chan struct{}, chan struct{}) {
	taskDone, spinDone := make(chan struct{}), make(chan struct{})

	go func() {
		spinner, i := []rune{'|', '/', '-', '\\'}, 0

		for {
			select {
			case <-taskDone:
				writer("\r \n")
				spinDone <- struct{}{}
				return
			default:
				i++
				writer("\r%c", spinner[i%len(spinner)])
				<-time.After(150 * time.Millisecond)
			}
		}
	}()

	return taskDone, spinDone
}
