package main

import (
    "fmt"
    "flag"
    "locate"
)

type Record struct {
    Name string
    Path string
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

    list, err := locate.ReadIndex(*idxFileArg)
    if err != nil {
        fmt.Println("error reading index:", err)
        return
    }

    results := locate.SearchIndex(list, *patternArg)
    for _, r := range results {
        fmt.Println(r.Path)
    }
}
