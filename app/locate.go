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
	patternArg = flag.String("pattern", "", "")
	idxFileArg = flag.String("idx", "./index.db", "")
)

func main() {
	flag.Parse()
	if *patternArg == "" {
		fmt.Println("must provide a pattern to search for")
		return
	}

	n := runtime.NumCPU()
    runtime.GOMAXPROCS(n)
	results, err := locate.SearchSplitIndex(*idxFileArg, *patternArg, n)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, r := range results {
		fmt.Println(r.Path)
	}
}
