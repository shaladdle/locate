package main

import (
	"flag"
	"fmt"
	"locate"
)

type Record struct {
	Name  string
	Path  string
	IsDir bool
}

var (
	rootArg    = flag.String("root", "/home/adam", "")
	idxFileArg = flag.String("idx", "./index.db", "")
)

func main() {
	flag.Parse()
	fmt.Printf("creating index for \"%s\"..", *rootArg)
	if err := locate.WriteIndex(
		*idxFileArg,
		locate.CreateIndex(*rootArg),
	); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("done")
	}
}
