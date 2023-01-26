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
    packetPrefix = "name_pref"
    //tmpl = template.Must(template.ParseFiles(
    //templatesPath+"index.html"))
    hosts []PiHost
)

type PiHost struct {
    Name string
    Ip string
    InnerIPs []string
    Actived bool
}

func OnMessageReaded(p *tcpPacket, addr net.Addr){
    //Урезаю значение порта
    outerAddr := strings.Split(addr.String(), ":")[0]
    contains := false
    for i, host := range hosts {
        if host.Name == p.deviceName {
            hosts[i].Ip = outerAddr
            hosts[i].Actived = true
            contains = true
            break
        } else {
            if host.Ip == outerAddr {
                hosts[i].Name = p.deviceName
                contains = true
                break
            }
        }
    }
    if !contains {
        fmt.Println(addr.String())
        hosts = append(hosts, PiHost {p.deviceName, outerAddr, p.innerAddrs, true})
    }
}

func OnConnAccepted(addr net.Addr){
    log.Printf("accepted connection %v", addr)
}

func OnConnClosed(addr net.Addr){
    addrStr := addr.String()
    for i, host := range hosts {
        if host.Ip == addrStr {
            fmt.Println(host.Ip + " disactived");
            hosts[i].Actived = false
        }
    }
    log.Printf("closed connection %v", addr)
}

func OnConnError(addr net.Addr, err error){
    log.Printf("error accepting connection %v reason:  %s", addr, err)
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
    server.OnMessageReaded = OnMessageReaded

    go HttpServer(httpAddr)
    err := server.ListenAndServe()
    if err != nil {
        log.Println(err)
    }
}

