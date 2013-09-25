package locate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func predicate(r Record, pattern string) bool {
	return r.Name == pattern
}

func init() {
	//json.Register(Record{})
}

type Record struct {
	Name  string
	Path  string
	IsDir bool
}

func CreateIndex(root string) []Record {
	list := []Record{}
	filepath.Walk(root, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		list = append(list, Record{info.Name(), fpath, info.IsDir()})

		return nil
	})

	return list
}

func WriteIndex(fpath string, index []Record) error {
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(index)
}

func ReadIndex(fpath string) ([]Record, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	list := []Record{}

	err = json.NewDecoder(f).Decode(&list)
	return list, err
}

func SearchIndex(index []Record, pattern string) []Record {
	matches := []Record{}

	for _, r := range index {
		if predicate(r, pattern) {
			matches = append(matches, r)
		}
	}

	return matches
}

// WriteSplitIndex writes a list of records to 8 separate files
func WriteSplitIndex(fpath string, index []Record, n int) error {
	done := make(chan error, n)

	encodeIndex := func(w io.Writer, start int) {
		enc := json.NewEncoder(w)
		for j := 0; j < len(index); j += n {
			if idx := j + start; idx < len(index) {
				if err := enc.Encode(index[idx]); err != nil {
					done <- err
					return
				}
			}
		}

		done <- nil
	}

	for i := 0; i < n; i++ {
		f, err := os.Create(fmt.Sprintf("%s%d", fpath, i))
		if err != nil {
			return err
		}
		defer f.Close()

		go encodeIndex(f, i)
	}

	var ret error
	for i := 0; i < n; i++ {
		if err := <-done; err != nil {
			ret = err
		}
	}

	return ret
}

func SearchSplitIndex(fpath, pattern string, n int) ([]Record, error) {
	resCh := make(chan Record, n)
	doneCh := make(chan error, n)

	decodeAndSearch := func(r io.Reader) {
		dec := json.NewDecoder(r)

		for {
			var r Record
			if err := dec.Decode(&r); err == io.EOF {
				doneCh <- nil
				return
			} else if err != nil {
				doneCh <- err
			} else if predicate(r, pattern) {
				resCh <- r
			}
		}

		doneCh <- nil
	}

	for i := 0; i < n; i++ {
		f, err := os.Open(fmt.Sprintf("%s%d", fpath, i))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		go decodeAndSearch(f)
	}

	list := []Record{}
	done := 0
	var err error
	for i := 0; i < n; i++ {
		select {
		case err1 := <-doneCh:
			if err1 != nil {
				err = err1
			}

			if done++; done == n {
				return list, err
			}
		case r := <-resCh:
			list = append(list, r)
		}
	}

	return list, nil
}
