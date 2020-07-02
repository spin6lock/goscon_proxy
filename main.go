package main

import (
	"flag"
	"log"
	"net"
	"strconv"
    "fmt"
    "os"

	"github.com/ejoy/goscon/scp"
)

//target server address
var optConnect string
func main() {
	port := flag.Int("port", 3333, "Port to accept connections on.")
	host := flag.String("host", "127.0.0.1", "Host or IP to bind to")
	flag.StringVar(&optConnect, "connect", "127.0.0.1:1248", "connect to scon server")
	flag.Parse()

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
    plaintext_buf := make([]byte, 1024)
    scp_buf := make([]byte, 1024)
	for {
		size, err := conn.Read(plaintext_buf)
		if err != nil {
			return
		}
		data := plaintext_buf[:size]
		log.Println("Read new data from connection", data)
        scon.Write(data)
        size, err = scon.Read(scp_buf)
		if err != nil {
            log.Println("scon read error:%s", err.Error())
			return
		}
		data = scp_buf[:size]
        log.Println("Read data from scp:", data)
        conn.Write(data)
	}
}
