package locate

import (
	"encoding/gob"
	"os"
	"path/filepath"
)

func init() {
    gob.Register(Record{})
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

	return gob.NewEncoder(f).Encode(index)
}

func ReadIndex(fpath string) ([]Record, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	list := []Record{}

    err = gob.NewDecoder(f).Decode(&list)
	return list, err
}

func SearchIndex(index []Record, pattern string) []Record {
	predicate := func(r Record) bool {
		return r.Name == pattern
	}

	matches := []Record{}

	for _, r := range index {
		if predicate(r) {
			matches = append(matches, r)
		}
	}

	return matches
}
