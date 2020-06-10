package meter

import (
	"io"
	"sync"
)

type Writer struct {
	l    sync.Mutex
	w    io.Writer
	b, n int64
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) Write(b []byte) (int, error) {
	w.l.Lock()
	defer w.l.Unlock()

	n, err := w.w.Write(b)
	if err == nil {
		w.n += int64(n)
	}
	return n, err
}

func (w *Writer) Total() int64 {
	w.l.Lock()
	defer w.l.Unlock()

	return w.n
}

func (w *Writer) Delta() int64 {
	w.l.Lock()
	defer w.l.Unlock()

	n := w.n - w.b
	w.b = w.n
	return n
}
