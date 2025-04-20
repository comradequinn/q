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

func Spin() (stopFunc func()) {
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

	return func() {
		taskDone <- struct{}{}
		<-spinDone
	}
}
