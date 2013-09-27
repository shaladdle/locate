package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"locate"
	"net"
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

	if conn, err := net.Dial("tcp", *hostportArg); err != nil {
		fmt.Println("error dialing:", err)
	} else {
		dec, enc := gob.NewDecoder(conn), gob.NewEncoder(conn)

		if err := enc.Encode(*patternArg); err != nil {
			fmt.Println("error encoding pattern:", err)
			return
		}

		var results []locate.Record
		if err := dec.Decode(&results); err != nil {
			fmt.Println("error decoding results:", err)
			return
		}

		for _, r := range results {
			fmt.Println(r.Path)
		}
	}
}
