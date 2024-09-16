package http

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	t.Run("pool test", func(t *testing.T) {

		const (
			max      = 5
			attempts = max * 2
			msg      = "bye\n"
		)

		l, err := net.Listen("tcp", "127.0.0.1:8888")
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		pool := NewPool(l, 5)

		wg.Add(1)
		saturated := make(chan struct{})
		go func() {
			defer wg.Done()

			accepted := 0
			for {
				c, err := pool.Accept()
				if err != nil {
					t.Log(err)
					break
				}

				defer c.Close()

				accepted++
				if accepted == max {
					close(saturated)
				}
			}

			if accepted != max {
				t.Errorf("want exactly %d", max)
			}
		}()

		dialer := &net.Dialer{}

		dialCtx, cancelDial := context.WithCancel(context.Background())
		defer cancelDial()

		var dialed, served int32
		var pendingDials sync.WaitGroup

		for n := attempts; n > 0; n-- {
			wg.Add(1)
			pendingDials.Add(1)

			go func() {
				defer wg.Done()
				c, err := dialer.DialContext(dialCtx, l.Addr().Network(), l.Addr().String())
				pendingDials.Done()
				if err != nil {
					t.Log(err)
					return
				}

				atomic.AddInt32(&dialed, 1)

				defer c.Close()

				if b, err := io.ReadAll(c); len(b) < len(msg) {
					t.Log(err)
					return
				}
				atomic.AddInt32(&served, 1)
			}()
		}

		//<-saturated
		time.Sleep(10 * time.Millisecond)
		cancelDial()
		pendingDials.Wait()
		pool.Close()
		if served > max {
			t.Errorf("expected at most %d served", max)
		}
	})
}
