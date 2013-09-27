package locate

import (
	"os"
	"path/filepath"
	"time"
)

type Record struct {
	Path  string
	Name  string
	IsDir bool
}

func predicate(r Record, pattern string) bool {
	return r.Name == pattern
}

type filepathWalker string

func (r filepathWalker) Walk(f func(string, os.FileInfo, error) error) error {
	return filepath.Walk(string(r), f)
}

type walker interface {
	Walk(func(string, os.FileInfo, error) error) error
}

type Index interface {
	Search(pattern string) []Record
}

func NewIndex(nthreads int, period time.Duration, root string) Index {
	return newIndex(nthreads, period, filepathWalker(root))
}

func newIndex(nthreads int, period time.Duration, w walker) *index {
	idx := &index{
		nthreads:    nthreads,
		indexPeriod: period,
		data:        [][]Record{},
		w:           w,

		searchReq:  make(chan searchReqMsg),
		searchDone: make(chan searchDoneMsg),
		createDone: make(chan createDoneMsg),
	}

	go idx.create()
	go idx.director()

	return idx
}

type searchReqMsg struct {
	pattern string
	reply   chan []Record
}

type searchDoneMsg struct {
	pattern string
	results []Record
}

type createDoneMsg [][]Record

type index struct {
	nthreads    int
	data        [][]Record
	indexPeriod time.Duration
	w           walker

	searchReq  chan searchReqMsg
	searchDone chan searchDoneMsg
	createDone chan createDoneMsg
}

func (p *index) Search(pattern string) []Record {
	reply := make(chan []Record)
	p.searchReq <- searchReqMsg{pattern, reply}
	return <-reply
}

func (p *index) search(pattern string) {
	results := make(chan []Record)
	for i := 0; i < p.nthreads; i++ {
		go func(n int) {
			results <- p.searchi(pattern, n)
		}(i)
	}

	ret := []Record{}
	for i := 0; i < p.nthreads; i++ {
		ret = append(ret, (<-results)...)
	}
	p.searchDone <- searchDoneMsg{pattern, ret}
}

func (p *index) searchi(pattern string, i int) []Record {
	ret := []Record{}
	for _, r := range p.data[i] {
		if predicate(r, pattern) {
			ret = append(ret, r)
		}
	}
	return ret
}

func (p *index) director() {
	pendingSearches := make(map[string][]searchReqMsg)

	doIndex := time.NewTicker(p.indexPeriod)

	p.data = <-p.createDone

	for {
		select {
		case msg := <-p.searchReq:
			go p.search(msg.pattern)
			pendingSearches[msg.pattern] = append(pendingSearches[msg.pattern], msg)
		case msg := <-p.searchDone:
			if reqs, ok := pendingSearches[msg.pattern]; ok {
				for _, req := range reqs {
					req.reply <- msg.results
				}
			}
		case <-doIndex.C:
			go p.create()
		case msg := <-p.createDone:
			p.data = msg
		}
	}
}

func (p *index) create() {
	data := make([][]Record, p.nthreads)
	i := 0

	p.w.Walk(func(fpath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		data[i%p.nthreads] = append(data[i%p.nthreads], Record{fpath, info.Name(), info.IsDir()})
		i++

		return nil
	})

	p.createDone <- data
}
