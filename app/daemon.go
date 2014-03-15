package main

import (
	"flag"
	"fmt"
	"github.com/shaladdle/locate"
)

var (
	root     = flag.String("root", "/Users/Adam", "")
	hostport = flag.String("addr", ":8000", "")
)

func main() {
	flag.Parse()

	dmn, err := locate.NewServer(*root, *hostport)
	if err != nil {
		fmt.Println("NewServer error:", err)
		return
	}

	fmt.Println("Error:", dmn.Wait())
}
