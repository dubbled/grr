package grr

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ThreadType determines the different types of a thread
type ThreadType int

// All ThreadTypes are listed here
const (
	Simple ThreadType = iota
	Iterator
	Ticker
)

// ikey is an internal type for holding within Thread object contexts
type ikey int

const (
	iterations ikey = iota
	timer
)

// Thread represents a goroutine
type Thread struct {
	ctx    context.Context
	ID     uuid.UUID
	Status chan message
	Typ    ThreadType
	F      func()
}

type message struct {
	Time       time.Time
	Msg        StatCode
	Additional interface{}
}

// StatCode represents a message type on a thread's status channel
type StatCode int

// All types of StatCodes
const (
	ThreadPending StatCode = iota
	ThreadStarted
	ThreadFinished
	ThreadIterate
)

func makeThread(ctx context.Context, t ThreadType, f func()) *Thread {
	var statusBuf int
	switch t {
	case Simple:
		statusBuf = 3
	default:
		statusBuf = ctx.Value(iterations).(int) + 3
	}

	statusChan := make(chan message, statusBuf)
	statusChan <- message{
		time.Now(),
		ThreadPending,
		nil,
	}

	return &Thread{
		ctx:    ctx,
		ID:     uuid.New(),
		Typ:    t,
		F:      f,
		Status: statusChan,
	}
}

// NewSimple creates a new simple goroutine
func NewSimple(f func()) *Thread {
	return makeThread(
		context.Background(),
		Simple,
		f,
	)
}

// NewIterator creates a new iterator goroutine
// Parameter n dictates how many times the function will run
// if n < 1, f will run indefinitely
func NewIterator(f func(), n int) *Thread {
	ctx := context.Background()
	ctx = context.WithValue(ctx, iterations, n)

	return makeThread(
		ctx,
		Iterator,
		f,
	)
}

// NewTicker creates a new ticker goroutine
// Parameter t dictates how long the thread sleeps once the function is finished
// Parameter n dictates how many times the function will run.
// if n < 1, f will run indefinitely
func NewTicker(f func(), t time.Duration, n int) *Thread {
	ctx := context.Background()
	ctx = context.WithValue(ctx, timer, t)
	ctx = context.WithValue(ctx, iterations, n)

	return makeThread(
		ctx,
		Ticker,
		f,
	)
}

func (t *Thread) iterate(i int) {
	t.Status <- message{
		time.Now(),
		ThreadIterate,
		i,
	}
}

// Start starts a thread's task
func (t *Thread) Start() {
	t.Status <- message{
		time.Now(),
		ThreadStarted,
		nil,
	}

	switch t.Typ {
	case Simple:
		t.F()
	case Iterator:
		loops := t.ctx.Value(iterations).(int)

	Iterator:
		for loops > 0 {
			loops--
			t.iterate(loops)

			select {
			case <-t.ctx.Done():
				break Iterator
			default:
				t.F()
			}

		}

	case Ticker:
		td := t.ctx.Value(timer).(time.Duration)
		loops := t.ctx.Value(iterations).(int)

	Ticker:
		for loops > 0 {
			loops--
			t.iterate(loops)

			select {
			case <-t.ctx.Done():
				break Ticker
			default:
				t.F()
				if loops != 0 { // last tick
					time.Sleep(td)
				}
			}
		}
	}

	t.Stop()
}

// Stop terminates a thread's process assuming function execution has not already begun
func (t *Thread) Stop() {
	ctx, cancel := context.WithCancel(t.ctx)
	t.ctx = ctx
	cancel()

	t.Status <- message{
		time.Now(),
		ThreadFinished,
		nil,
	}

	close(t.Status)
}
