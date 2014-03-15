package locate

import (
    "reflect"
    "testing"
    "encoding/gob"
)

type simpleTestResult string

func (r simpleTestResult) Name() string {
    return string(r)
}

func (r simpleTestResult) Size() int {
    return 0
}

type simpleLocator []Result

func (sl simpleLocator) Locate(query string) []Result {
    return sl
}

type mirrorLocator struct {}

func (mirrorLocator) Locate(query string) []Result {
    return []Result{simpleTestResult(query)}
}

func TestClientServerComm(t *testing.T) {
    gob.Register(simpleTestResult(""))

    results := []Result{simpleTestResult("hi there")}
    l := simpleLocator(results)

    srv, err := newServer("/", ":9000", l)
    if err != nil {
        t.Fatalf("newServer error: %v", err)
    }

    cli, err := NewClient(":9000")
    if err != nil {
        t.Fatalf("NewClient error: %v", err)
    }

    got := cli.Locate("banana")

    if !reflect.DeepEqual(results, got) {
        t.Fatalf("Results don't match :(")
    }

    srv.Close()
}
