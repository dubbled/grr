package grr

import (
	"fmt"
	"testing"
	"time"
)

func listenAndPrint(p *Thread) {
	for out := range p.Status {
		fmt.Println(out)
		if out.Msg == ThreadFinished {
			break
		}
	}
}

func TestSimple(t *testing.T) {
	th := NewSimple(func() {
		fmt.Println("Testing Simple")
	})

	go th.Start()
	listenAndPrint(th)
}

func TestIterator(t *testing.T) {
	th := NewIterator(func() {
		fmt.Println("Testing Iterator")
		time.Sleep(time.Second * 3)
	}, 3)

	go th.Start()
	listenAndPrint(th)
}

func TestTicker(t *testing.T) {
	th := NewTicker(func() {
		fmt.Println("Testing Ticker")
	}, time.Second*3, 3)

	go th.Start()
	listenAndPrint(th)
}

func TestCancel(t *testing.T) {
	th := NewTicker(func() {
		fmt.Println("Testing Cancellation")
	}, time.Second*3, 5)

	go th.Start()
	for out := range th.Status {
		fmt.Println(out)
		if out.Msg == ThreadIterate {
			th.Stop()
		} else if out.Msg == ThreadFinished {
			break
		}
	}
}
