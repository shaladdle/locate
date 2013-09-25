package main

import (
	"flag"
	"fmt"
	"locate"
	"runtime"
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
    n := runtime.NumCPU()
    runtime.GOMAXPROCS(n)
	if err := locate.WriteSplitIndex(
		*idxFileArg,
		locate.CreateIndex(*rootArg),
		n,
	); err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("done")
	}
}
