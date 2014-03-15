package main

import (
	"flag"
	"fmt"

	"github.com/shaladdle/locate"
)

var (
	patternArg  = flag.String("pattern", "", "")
	hostportArg = flag.String("addr", ":8000", "")
)

func main() {
	flag.Parse()
	if *patternArg == "" {
		fmt.Println("must provide a pattern to search for")
		return
	}

	cli, err := locate.NewClient(*hostportArg)
	if err != nil {
		fmt.Println("NewClient error:", err)
		return
	}

	for _, info := range cli.Locate(*patternArg) {
		fmt.Println(info.Name())
	}
}
