package main

import (
    "net"
    "sync"
    "log"
    "time"
    "io"
    "encoding/gob"
)

type tcpPacket struct {
        DeviceName string
        InnerAddrs []string
}

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
    OnMessageReaded func(tcpPacket, net.Addr)
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
    dec := gob.NewDecoder(conn)
    p := tcpPacket{}
    for {
        err := dec.Decode(&p)
        if err != nil {
            log.Println(err)
        }
        if err != nil {
            log.Printf("%v(%v)", err, conn.RemoteAddr())
            return err
        }
        if OnMessageReaded != nil {
            OnMessageReaded(p, conn.RemoteAddr())
        }
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
