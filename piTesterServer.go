package main

import (
    "net"
    "sync"
    "log"
    "bufio"
    "time"
    "flag"
    "strings"
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
	OnAccept func()
    OnConnClosed func()
    OnAcceptError func()
    OnMessageReaded func(string)
    ctx     *context
	mu      sync.Mutex

}

func NewPiTesterServer(addr string, idleTimeout time.Duration, maxReadBuffer int64) *PiTesterServer {
	return &PiTesterServer {
        Addr : addr,
        IdleTimeout: idleTimeout,
        MaxReadBuffer: maxReadBuffer,
		OnAccept:  nil,
		OnConnClosed:  nil,
        OnAcceptError:  nil,
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
            log.Printf("error accepting connection %v", err)
            continue
        }
        log.Printf("accepted connection from %v", newConn.RemoteAddr())
        conn := &Conn{
            Conn:        newConn,
            IdleTimeout: srv.IdleTimeout,
            MaxReadBuffer: srv.MaxReadBuffer,
        }
        conn.SetDeadline(time.Now().Add(conn.IdleTimeout))
        go handle(conn)
    }
}

func handle(conn net.Conn) error {
    defer func() {
        log.Printf("closing connection from %v", conn.RemoteAddr())
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
        w.WriteString(strings.ToUpper(scanr.Text()) + "\n")
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

//vds1.proxycom.ru:1288
func main () {
    var serverAddr string
    var timeOut uint
    var maxReadBuff int64
    flag.StringVar(&serverAddr, "address", "127.0.0.1:23", "Bind listen socket address")
    flag.UintVar(&timeOut, "timeout", 3, "Conection idle timeout (sec)")
    flag.Int64Var(&maxReadBuff, "maxBuffSize", 4096, "Max buffer size")
    flag.Parse()
    server := NewPiTesterServer(serverAddr, time.Duration(timeOut) * time.Second, maxReadBuff)
    err := server.ListenAndServe()
    if err != nil {
        log.Println(err)
    }
}

