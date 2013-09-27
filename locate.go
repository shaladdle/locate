package locate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Index interface {
	Search(pattern string) []Record
	Create(root string)
}

func NewIndex() Index {
	return make(hashIndex)
}

type hashIndex map[string][]Record

func (h hashIndex) Create(root string) {
	filepath.Walk(root, func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		h[info.Name()] = append(h[info.Name()], Record{info.Name(), fpath, info.IsDir()})

		return nil
	})
}

func (h hashIndex) Search(pattern string) []Record {
	if r, ok := h[pattern]; ok {
		return r
	}

	return nil
}

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

	enc := json.NewEncoder(f)

	for _, r := range index {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}

	return nil
}

func ReadIndex(fpath string) ([]Record, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	list := []Record{}

	for {
		var r Record
		if err := dec.Decode(&r); err == io.EOF {
			return list, nil
		} else if err != nil {
			return nil, err
		}
		list = append(list, r)
	}

	return list, nil
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

func SplitSearchSingleReader(fpath, pattern string, n int) ([]Record, error) {
	connector := make(chan Record)
	filtered := make(chan Record)
	readerErr := make(chan error)
	done := make(chan bool)

	reader := func(r io.Reader, next chan<- Record) {
		dec := json.NewDecoder(r)

		for {
			var r Record
			if err := dec.Decode(&r); err == io.EOF {
				close(next)
				break
			} else if err != nil {
				readerErr <- err
				close(next)
			} else {
				next <- r
			}
		}
	}

	searcher := func(input <-chan Record, output chan<- Record) {
		for r := range input {
			if predicate(r, pattern) {
				output <- r
			}
		}
		done <- true
	}

	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	go reader(f, connector)

	for i := 0; i < n; i++ {
		go searcher(connector, filtered)
	}

	list := []Record{}
	doneN := 0
	for {
		select {
		case r := <-filtered:
			list = append(list, r)
		case <-done:
			doneN++
			if doneN >= n {
				return list, nil
			}
		case err := <-readerErr:
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, fmt.Errorf("Unreachable")
}
