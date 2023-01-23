package main

import (
    "net"
    "log"
    "time"
    "flag"
    "fmt"
    "strconv"
    "net/http"
    "html/template"
)

var(
    templatesPath = "static/templates/"
    tmpl = template.Must(template.ParseFiles(
    templatesPath+"index.html"))
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

func index(w http.ResponseWriter, r *http.Request){
    tmpl.ExecuteTemplate(w, "index.html", nil)
}

func HttpServer(httpAddr string) {
    //Задание начальных настроек
	var port int

	flag.IntVar(&port, "port", port, "Server listen port")
	flag.Parse()
	fmt.Println(":" + strconv.Itoa(port))

	//Запуск сервера
	http.HandleFunc("/", index)

	err := http.ListenAndServe(httpAddr, nil)
	if err != nil {
		log.Print(err)
	}
}
//vds1.proxycom.ru:1288
func main () {
    var tcpAddr string
    var httpAddr string
    var timeOut uint
    var maxReadBuff int64
    flag.StringVar(&tcpAddr, "tcpAddress", "vds1.proxinet.ru:1288", "Address to tcp server")
    flag.StringVar(&httpAddr, "httpAddress", "127.0.0.1:1289", "Address to http server")
    flag.UintVar(&timeOut, "timeout", 3, "Conection idle timeout (sec)")
    flag.Int64Var(&maxReadBuff, "maxBuffSize", 4096, "Max buffer size")
    flag.Parse()
    server := NewPiTesterServer(tcpAddr, time.Duration(timeOut) * time.Second, maxReadBuff)
    server.OnConnAccepted = OnConnAccepted
    server.OnConnClosed = OnConnClosed
    server.OnConnError = OnConnError
    go HttpServer(httpAddr)
    err := server.ListenAndServe()
    if err != nil {
        log.Println(err)
    }
}

