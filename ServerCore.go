package main

import (
    "net"
    "sync"
    "log"
    "bufio"
    "time"
    "io"
)

type context struct {
	stop chan bool
	done chan bool
	err  error
}

func newContext() *context {
	return &context{
		stop: make(chan bool),
		done: make(chan bool),
	}
}

type PiTesterServer struct {
    Addr string
    IdleTimeout time.Duration
    MaxReadBuffer int64
	OnConnAccepted func(net.Addr)
    OnConnClosed func(net.Addr)
    OnConnError func(net.Addr, error)
    OnMessageReaded func(string)
    ctx     *context
	mu      sync.Mutex

}

func NewPiTesterServer(addr string, idleTimeout time.Duration, maxReadBuffer int64) *PiTesterServer {
	return &PiTesterServer {
        Addr : addr,
        IdleTimeout: idleTimeout,
        MaxReadBuffer: maxReadBuffer,
		OnConnAccepted:  nil,
		OnConnClosed:  nil,
        OnConnError: nil,
        OnMessageReaded: nil,
	}
}

func (srv PiTesterServer) ListenAndServe() error {
    addr := srv.Addr
    if addr == "" {
        addr = ":8080"
    }
    log.Printf("starting Pi tester server on %v\n", addr)
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    defer listener.Close()
    for {
        newConn, err := listener.Accept()
        if err != nil {
            if srv.OnConnError != nil {
                srv.OnConnError(newConn.RemoteAddr(), err)
            }
            continue
        }
        conn := &Conn{
            Conn:        newConn,
            IdleTimeout: srv.IdleTimeout,
            MaxReadBuffer: srv.MaxReadBuffer,
        }
        if srv.OnConnAccepted != nil {
            srv.OnConnAccepted(conn.RemoteAddr())
        }
        conn.SetDeadline(time.Now().Add(conn.IdleTimeout))
        go srv.handle(conn)
    }
}

func (srv PiTesterServer) handle(conn net.Conn) error {
    defer func() {
            if srv.OnConnClosed != nil {
                srv.OnConnClosed(conn.RemoteAddr())
            }
        conn.Close()
    }()
    r := bufio.NewReader(conn)
    w := bufio.NewWriter(conn)
    scanr := bufio.NewScanner(r)
    for {
        scanned := scanr.Scan()
        if !scanned {
            if err := scanr.Err(); err != nil {
                log.Printf("%v(%v)", err, conn.RemoteAddr())
                return err
            }
            break
        }
        w.WriteString("!accepted!\n")
        w.Flush()
    }
    return nil
}

type Conn struct {
    net.Conn
    IdleTimeout time.Duration
    MaxReadBuffer int64

}

func (c *Conn) Write(p []byte) (int, error) {
    c.updateDeadline()
    return c.Conn.Write(p)
}

func (c *Conn) Read(b []byte) (int, error) {
    c.updateDeadline()
    r := io.LimitReader(c.Conn, c.MaxReadBuffer)
    return r.Read(b)
}

func (c *Conn) updateDeadline() {
    idleDeadline := time.Now().Add(c.IdleTimeout)
    c.Conn.SetDeadline(idleDeadline)
}
