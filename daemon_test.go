package locate

import (
    "fmt"
    "testing"
    "time"
)

func TestDaemon(t *testing.T) {
    d, err := NewDaemon("/Users/Adam/")
    if err != nil {
        t.Fatalf("NewDaemon error: %v", err)
    }

    time.Sleep(time.Second * 40)

    for _, info := range d.Locate("main.c") {
        fmt.Println(info.Name())
    }
}
