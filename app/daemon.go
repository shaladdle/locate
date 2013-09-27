package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"locate"
	"net"
	"runtime"
	"time"
)

var hostport = flag.String("addr", ":8000", "")

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", *hostport)
	if err != nil {
		fmt.Println("error listening:", err)
		return
	}

	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	index := locate.NewIndex(n, 4*time.Hour, "/")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
		}

		go clientHandler(conn, index)
	}
}

func clientHandler(conn net.Conn, index locate.Index) {
	defer conn.Close()

	dec, enc := gob.NewDecoder(conn), gob.NewEncoder(conn)

	var pattern string
	if err := dec.Decode(&pattern); err != nil {
		fmt.Println("error decoding pattern:", err)
		return
	}

	if err := enc.Encode(index.Search(pattern)); err != nil {
		fmt.Println("error encoding results:", err)
		return
	}
}
