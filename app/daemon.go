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

var (
	nthreads = flag.Int("nthreads", 0, "")
	root     = flag.String("root", "/", "")
	hostport = flag.String("addr", ":8000", "")
	period   = flag.Duration("period", 4*time.Hour, "")
)

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", *hostport)
	if err != nil {
		fmt.Println("error listening:", err)
		return
	}

	if *nthreads == 0 {
		*nthreads = runtime.NumCPU()
		runtime.GOMAXPROCS(*nthreads)
	}

	index := locate.NewIndex(*nthreads, *period, *root)

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
