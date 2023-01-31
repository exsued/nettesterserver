package main

import (
    "net"
    "sync"
    "log"
    "time"
    "io"
    "os"
    "fmt"
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
func LogFile(out string, dirpath string) {
    nowtime := time.Now()
    finalString := nowtime.Format("15:04:05\t") + out + "\n"
    fileName := dirpath + nowtime.Format("2006-01-02") + ".txt"

    f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
    if _, err = f.WriteString(finalString); err != nil {
        log.Fatal(err)
    }
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
    BothLog("starting Pi tester server on (TCP): " + addr)
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
            msg := fmt.Sprintf("Decode err: %v(%v)\n", err.Error(), conn.RemoteAddr())
            BothLog(msg)
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
