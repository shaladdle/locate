package locate

import (
    "fmt"
    "net"
    "encoding/gob"
)

type Server interface {
    Wait() <-chan error
    Close()
}

type server struct {
    dmn Locator
    dead chan error
    wait chan chan error
}

func newServer(root, hostport string, locator Locator) (Server, error) {
    l, err := net.Listen("tcp", hostport)
    if err != nil {
        return nil, err
    }

    srv := &server{
        dmn: locator,
        dead: make(chan error),
    }

    go srv.director(l)

    return srv, nil
}

func NewServer(root, hostport string) (Server, error) {
    return newServer(root, hostport, NewDaemon(root))
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
    for {
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("accept error:", err)
            return
        }

        go srv.handleClient(conn)
    }
}

func (srv *server) Wait() <-chan error {
    reply := make(chan error, 1)
    srv.wait <- reply
    return reply
}

func (srv *server) Close() {
}
