package ratelimiter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var done = make(chan struct{})

type RateLimiter struct {
	// Function that returns if a rate limite shall be applied to this object
	Key     func(interface{}) bool
	channel chan interface{}
	r       io.Reader
	w       io.Writer
	sync.WaitGroup
}

func (rl *RateLimiter) writeLoop() {
	for {
		select {
		case <-time.Tick(time.Second * 2):
			select {
			case object := <- rl.channel:
				msg := fmt.Sprintf("\nRate limited: '%s'", object.(string))
				n, err := fmt.Fprintln(rl.w, msg)
				if err != nil {
					fmt.Println("N is %d and error is %s", n, err)
				}
			default:
				continue
			}
		case <-done:
			defer fmt.Println("Finish: write loop exited.")
			return
		}
	}
}

func (rl *RateLimiter) Limit(object interface{}) {
	if ! rl.Key(object) {
		rl.w.Write(object.([]byte))
	}
	rl.Add(1)
	go func() {
		defer rl.Done()
		select {
		case rl.channel <- object:
		default:
			errMsg := fmt.Sprintf("\nObject '%s' was discarded.\n", object)
			rl.w.Write([]byte(errMsg))
		}
	}()
}

func NewRateLimiter(cap int, w io.Writer) *RateLimiter {
	rl := &RateLimiter{
		Key: func(i interface{}) bool {return true},
		channel: make(chan interface{}, cap),
		w: w,
	}
	go rl.writeLoop()
	return rl
}

func main() {
	rl := NewRateLimiter(5, os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)
	defer fmt.Println("Finish: Wait group awaiting done.")
	defer rl.Wait()
	for scanner.Scan() {
		input := scanner.Text()
		if input == "exit" {
			close(done)
			fmt.Println("Finish: Channel closed")
			return
		}
		rl.Limit(input)
	}
}
