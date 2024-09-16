package http

import (
	"errors"
	"net"
	"sync"
)

var (
	ErrMaxActiveConnReached = errors.New("maximum active connections reached")
)

type PoolConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func(pConn *PoolConn)
}

func (p *PoolConn) Close() error {
	err := p.Conn.Close()
	p.releaseOnce.Do(func() { p.release(p) })
	return err
}

type ConnPool struct {
	net.Listener
	newConnsCount int
	maxActive     int
	listener      net.Listener
	activeConns   chan *PoolConn
	closeOnce     sync.Once
	close         chan struct{}
}

func NewPool(listener net.Listener, maxActive int) *ConnPool {
	return &ConnPool{
		listener:    listener,
		activeConns: make(chan *PoolConn, maxActive),
		maxActive:   maxActive,
		close:       make(chan struct{}),
	}
}

func (p *ConnPool) ReleaseConn(pConn *PoolConn) {
	select {
	case p.activeConns <- pConn:

	default:
	}
}

func (p *ConnPool) Accept() (net.Conn, error) {
	for {
		select {
		case conn := <-p.activeConns:
			return conn, nil
		default:
			if len(p.activeConns) > p.maxActive {
				return nil, ErrMaxActiveConnReached
			}

			conn, err := p.listener.Accept()
			if err != nil {
				p.Release()
				return nil, err
			}

			p.newConnsCount++

			return &PoolConn{Conn: conn, release: p.ReleaseConn}, nil
		}
	}
}

func (p *ConnPool) Release() {
	close(p.activeConns)
	for c := range p.activeConns {
		c.Close()
	}
}

func (p *ConnPool) Close() error {
	p.closeOnce.Do(func() {
		close(p.close)
		p.Release()
	})

	return nil
}
