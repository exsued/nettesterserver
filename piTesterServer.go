package main

import (
    "net"
    "log"
    "time"
    "flag"
    "fmt"
    "strings"
    "strconv"
    "net/http"
    "html/template"
)

var(
    templatesPath = "static/templates/"
    logPath = ""
    //tmpl = template.Must(template.ParseFiles(
    //templatesPath+"index.html"))
    hosts []PiHost
)

func BothLog(msg string){
    log.Println(msg)
    LogFile(msg, logPath)
}
type PiHost struct {
    Name string
    Ip string
    InnerIPs []string
    Actived bool
}
func OnMessageReaded(p tcpPacket, addr net.Addr){
    /*Урезаю значение порта*/ outerAddr := strings.Split(addr.String(), ":")[0]
    contains := false
    for i, host := range hosts {
        if host.Name == p.DeviceName {
            hosts[i].Ip = outerAddr
            hosts[i].Actived = true
            contains = true
            break
        } else {
            if host.Ip == outerAddr {
                hosts[i].Name = p.DeviceName
                contains = true
                break
            }
        }
    }
    if !contains {
        hosts = append(hosts, PiHost {p.DeviceName, outerAddr, p.InnerAddrs, true})
        msg := fmt.Sprintf("added new node %v\n", (hosts[len(hosts) - 1]))
        BothLog(msg)
    }
}
func OnConnAccepted(addr net.Addr){
    msg := fmt.Sprintf("accepted connection %v", addr)
    BothLog(msg)
}
func OnConnClosed(addr net.Addr){
    addrStr := strings.Split(addr.String(), ":")[0]
    for i, host := range hosts {
        if host.Ip == addrStr {
            hosts[i].Actived = false
        }
    }
    msg := fmt.Sprintf("closed connection %v", addr)
    BothLog(msg)
}
func OnConnError(addr net.Addr, err error){
    msg := fmt.Sprintf("error accepting connection %v reason:  %s", addr, err)
    BothLog(msg)
}
func index(w http.ResponseWriter, r *http.Request){
    tmpl := template.Must(template.ParseFiles(templatesPath+"index.html"))
    tmpl.ExecuteTemplate(w, "index.html", hosts)
}
func HttpServer(httpAddr string) {
    //Задание начальных настроек
	var port int

	flag.IntVar(&port, "port", port, "Server listen port")
	flag.Parse()
	fmt.Println(":" + strconv.Itoa(port))

    fs := http.FileServer(http.Dir("static"))

	//Запуск сервера
    http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", index)

	err := http.ListenAndServe(httpAddr, nil)
	if err != nil {
        BothLog(err.Error())
	}
}

func main () {
    var tcpAddr string
    var httpAddr string
    var timeOut uint
    var maxReadBuff int64
    flag.StringVar(&tcpAddr, "tcpAddress", "vds1.proxinet.ru:1288", "Address to tcp server")
    flag.StringVar(&httpAddr, "httpAddress", "127.0.0.1:1289", "Address to http server")
    flag.StringVar(&logPath, "logPath", "./logs/", "Logs dir")
    flag.UintVar(&timeOut, "timeout", 5, "Conection idle timeout (sec)")
    flag.Int64Var(&maxReadBuff, "maxBuffSize", 4096, "Max buffer size")
    flag.Parse()
    server := NewPiTesterServer(tcpAddr, time.Duration(timeOut) * time.Second, maxReadBuff)
    server.OnConnAccepted = OnConnAccepted
    server.OnConnClosed = OnConnClosed
    server.OnConnError = OnConnError
    server.OnMessageReaded = OnMessageReaded

    go HttpServer(httpAddr)
    err := server.ListenAndServe()
    if err != nil {
        BothLog(err.Error())
    }
}
