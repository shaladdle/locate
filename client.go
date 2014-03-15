package locate

import (
    "encoding/gob"
    "fmt"
    "net"
)

type queryChMsg struct {
    query string
    reply chan []Result
}

type client struct {
    queries chan queryChMsg
}

func NewClient(hostport string) (Locator, error) {
    cli := &client{
        queries: make(chan queryChMsg),
    }

    conn, err := net.Dial("tcp", hostport)
    if err != nil {
        return nil, err
    }

    go cli.director(conn)

    return cli, nil
}

func (cli *client) director(conn net.Conn) {
    enc := gob.NewEncoder(conn)
    dec := gob.NewDecoder(conn)
    for {
        msg := <-cli.queries

        if err := enc.Encode(msg.query); err != nil {
            fmt.Println("encode error:", err)
            return
        }

        var results []Result
        if err := dec.Decode(&results); err != nil {
            fmt.Println("decode error:", err)
            return
        }

        msg.reply <- results
    }
}

func (cli *client) Locate(query string) []Result {
    msg := queryChMsg{
        query: query,
        reply: make(chan []Result),
    }
    cli.queries <- msg
    return <-msg.reply
}
