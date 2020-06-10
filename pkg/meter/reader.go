package meter

import (
	"io"
	"sync"
)

type Reader struct {
	l    sync.Mutex
	r    io.ReadCloser
	b, n int64
}

func NewReader(r io.ReadCloser) *Reader {
	return &Reader{r: r}
}

func (r *Reader) Read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	if err == nil || err == io.EOF {
		r.l.Lock()
		defer r.l.Unlock()

		r.n += int64(n)
	}
	return n, err
}

func (r *Reader) Close() error {
	return r.r.Close()
}

func (r *Reader) Total() int64 {
	r.l.Lock()
	defer r.l.Unlock()

	return r.n
}

func (r *Reader) Delta() int64 {
	r.l.Lock()
	defer r.l.Unlock()

	n := r.n - r.b
	r.b = r.n
	return n
}
