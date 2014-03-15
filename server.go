package locate

import (
	"encoding/gob"
	"fmt"
	"net"
)

type Server interface {
	Wait() error
	Close()
}

type server struct {
	dmn  Locator
	stop chan chan bool
	wait chan chan error
}

func newServer(root, hostport string, locator Locator) (Server, error) {
	l, err := net.Listen("tcp", hostport)
	if err != nil {
		return nil, err
	}

	srv := &server{
		dmn:  locator,
		stop: make(chan chan bool),
		wait: make(chan chan error),
	}

	go srv.director(l)

	return srv, nil
}

func NewServer(root, hostport string) (Server, error) {
    dmn, err := NewDaemon(root)
    if err != nil {
        return nil, err
    }
	return newServer(root, hostport, dmn)
}

func (srv *server) handleClient(conn net.Conn) {
	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	var query string
	if err := dec.Decode(&query); err != nil {
		fmt.Println("decode error:", err)
		return
	}

	results := srv.dmn.Locate(query)

	if err := enc.Encode(results); err != nil {
		fmt.Println("encode error:", err)
		return
	}
}

func (srv *server) director(l net.Listener) {
	type acceptInfo struct {
		conn net.Conn
		err  error
	}
	acceptChan := make(chan acceptInfo)

	go func() {
		for {
			conn, err := l.Accept()
			acceptChan <- acceptInfo{conn, err}
			if err != nil {
				break
			}
		}
	}()

	var (
		waitReply  chan error
		closeReply chan bool
	)

loop:
	for {
		select {
		case info := <-acceptChan:
			if info.err != nil {
				fmt.Println("accept error:", info.err)
				waitReply <- info.err
				break loop
			}

			go srv.handleClient(info.conn)
		case waitReply = <-srv.wait:
		case closeReply = <-srv.stop:
			break loop
		}
	}

	l.Close()

	if waitReply != nil {
		close(waitReply)
	}
	if closeReply != nil {
		close(closeReply)
	}
}

// Wait blocks until the server hsa returned due to an error, at which point the
// offending error is returned.
func (srv *server) Wait() error {
	reply := make(chan error)
	srv.wait <- reply
	return <-reply
}

// Close stops the server and frees any resources that were held by the server.
// This function does not return until all resources have been freed.
func (srv *server) Close() {
	reply := make(chan bool)
	srv.stop <- reply
	<-reply
}
