package main

import (
	"flag"
	"log"
	"net"
	"strconv"
    "fmt"
    "os"
    "net/http"
    _ "net/http/pprof"

	"github.com/ejoy/goscon/scp"
)

//target server address
var optConnect string
func main() {
	port := flag.Int("port", 3333, "Port to accept connections on.")
	host := flag.String("host", "127.0.0.1", "Host or IP to bind to")
	flag.StringVar(&optConnect, "connect", "127.0.0.1:1248", "connect to scon server")
	flag.Parse()

    go func(){
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

	l, err := net.Listen("tcp", *host+":"+strconv.Itoa(*port))
	if err != nil {
		log.Panicln(err)
	}
	log.Println("Listening to connections at '"+*host+"' on port", strconv.Itoa(*port))
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Panicln(err)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	log.Println("Accepted new connection.")
	defer conn.Close()
	defer log.Println("Closed connection.")

	raw, err := net.Dial("tcp", optConnect)
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
	scon, _ := scp.Client(raw, nil)
    defer scon.Close()
    plaintext_buf := make([]byte, 1024)
    scp_buf := make([]byte, 1024)
    ch := make(chan string)
    readCh := make(chan []byte)
    writeCh := make(chan []byte)
    go func() {
        for {
            size, err := conn.Read(plaintext_buf)
            if err != nil {
                log.Printf("conn read error:%s\n", err.Error())
                break
            }
            data := plaintext_buf[:size]
            log.Println("Read new data from connection", data)
            readCh <- data
        }
        ch <- ""
        log.Println("conn.Read exit")
    }()
    go func() {
        for {
            size, err := scon.Read(scp_buf)
            if err != nil {
                log.Printf("scon read error:%s\n", err.Error())
                break
            }
            data := scp_buf[:size]
            log.Println("Read data from scp:", data)
            writeCh <- data
        }
        ch <- ""
        log.Println("scon.Read exit")
    }()
    go readPipe(readCh, ch, scon)
    go writePipe(writeCh, ch, conn)
    <-ch
    ch<-""
    log.Println("handle request close")
}

func readPipe(readCh chan []byte, ctrlCh chan string, scon *scp.Conn) {
	for {
        select {
        case data := <-readCh:
            scon.Write(data)
        case <-ctrlCh:
            log.Println("readPipe got close")
            return
        }
    }
}

func writePipe(writeCh chan []byte, ctrlCh chan string, conn net.Conn) {
    for {
        select {
        case data := <-writeCh:
            conn.Write(data)
        case <-ctrlCh:
            log.Println("writePipe got close")
            return
        }
    }
}
