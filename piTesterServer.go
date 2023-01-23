package main

import (
    "net"
    "log"
    "time"
    "flag"
)

func OnConnAccepted(addr net.Addr){
    log.Printf("accepted connection %v", addr)
}

func OnConnClosed(addr net.Addr){
    log.Printf("error accepting connection %v", addr)
}

func OnConnError(addr net.Addr, err error){
    log.Printf("error accepting connection %v reason:  %s", addr, err)
}
//vds1.proxycom.ru:1288
func main () {
    var serverAddr string
    var timeOut uint
    var maxReadBuff int64
    flag.StringVar(&serverAddr, "address", "vds1.proxinet.ru:1288", "Bind listen socket address")
    flag.UintVar(&timeOut, "timeout", 3, "Conection idle timeout (sec)")
    flag.Int64Var(&maxReadBuff, "maxBuffSize", 4096, "Max buffer size")
    flag.Parse()
    server := NewPiTesterServer(serverAddr, time.Duration(timeOut) * time.Second, maxReadBuff)
    server.OnConnAccepted = OnConnAccepted
    server.OnConnClosed = OnConnClosed
    server.OnConnError = OnConnError
    err := server.ListenAndServe()
    if err != nil {
        log.Println(err)
    }
}

