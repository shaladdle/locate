package locate

import (
	"encoding/gob"
	"io"
)

type resultsMsg []Result

type stringMsg string

type Reader interface {
	io.Reader
	io.ByteReader
}

func (sr stringMsg) Name() string {
	return string(sr)
}

func (sr stringMsg) Marshal(w io.Writer) error {
	return gob.NewEncoder(w).Encode(string(sr))
}

func (sr stringMsg) Unmarshal(r Reader) error {
	return gob.NewDecoder(r).Decode(&sr)
}

type marshaler interface {
	Marshal(w io.Writer) error
	Unmarshal(r Reader) error
}
