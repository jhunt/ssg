package ssg

import (
	"math"
	"sync"

	"github.com/jhunt/go-sample"
)

type delta struct {
	base, n int64
}

func (d *delta) set(to int64) {
	d.n = to
}

func (d *delta) delta() int64 {
	n := d.n - d.base
	d.base = d.n
	return n
}

func (d *delta) total() int64 {
	return d.n
}

type metrics struct {
	lock sync.Mutex

	Operations struct {
		Upload   int `json:"upload"`
		Download int `json:"download"`
		Expunge  int `json:"expunge"`
	} `json:"operations"`

	Canceled struct {
		Upload   int `json:"upload"`
		Download int `json:"download"`
	} `json:"canceled"`

	segments sample.Reservoir
	Segments struct {
		Total int `json:"total"`
		Bytes struct {
			Minimum int     `json:"minimum"`
			Maximum int     `json:"maximum"`
			Median  float64 `json:"median"`
			Sigma   float64 `json:"sigma"`
		} `json:"bytes"`
	} `json:"segments"`

	Transfer struct {
		Front struct {
			In  int64 `json:"in"`
			Out int64 `json:"out"`
		} `json:"front"`

		Back struct {
			In  int64 `json:"in"`
			Out int64 `json:"out"`
		} `json:"back"`
	} `json:"transfer"`
}

func newMetric(max int) *metrics {
	m := &metrics{}
	m.segments = sample.NewReservoir(max)
	return m
}

func (m *metrics) Recalculate() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.Segments.Total = m.segments.Seen()

	if v := m.segments.Minimum(); !math.IsNaN(v) {
		m.Segments.Bytes.Minimum = int(v)
	} else {
		m.Segments.Bytes.Minimum = 0
	}

	if v := m.segments.Maximum(); !math.IsNaN(v) {
		m.Segments.Bytes.Maximum = int(v)
	} else {
		m.Segments.Bytes.Maximum = 0
	}

	if v := m.segments.Median(); !math.IsNaN(v) {
		m.Segments.Bytes.Median = v
	} else {
		m.Segments.Bytes.Median = 0.0
	}

	if v := m.segments.Stdev(); !math.IsNaN(v) {
		m.Segments.Bytes.Sigma = v
	} else {
		m.Segments.Bytes.Sigma = 0.0
	}
}

func (m *metrics) Reset() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.Operations.Upload = 0
	m.Operations.Download = 0
	m.Operations.Expunge = 0

	m.Canceled.Upload = 0
	m.Canceled.Download = 0

	m.segments.Reset()

	m.Transfer.Front.In = 0
	m.Transfer.Front.Out = 0
	m.Transfer.Back.In = 0
	m.Transfer.Back.Out = 0
}

func (m *metrics) StartUpload() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Operations.Upload++
}

func (m *metrics) CancelUpload() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Canceled.Upload++
}

func (m *metrics) StartDownload() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Operations.Download++
}

func (m *metrics) CancelDownload() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Canceled.Download++
}

func (m *metrics) Expunge() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Operations.Expunge++
}

func (m *metrics) Segment(size int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.segments.Sample(float64(size))
}

func (m *metrics) InFront(bytes int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Transfer.Front.In += bytes
}

func (m *metrics) OutFront(bytes int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Transfer.Front.Out += bytes
}

func (m *metrics) InBack(bytes int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Transfer.Back.In += bytes
}

func (m *metrics) OutBack(bytes int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Transfer.Back.Out += bytes
}
