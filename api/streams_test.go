package api_test

import (
	"bytes"
	"sync"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/jhunt/shield-storage-gateway/api"
	"github.com/jhunt/shield-storage-gateway/backend"
)

type SharedMemory struct {
	lock sync.Mutex
	data map[string] []byte
}

func (sm *SharedMemory) Get(key string) []byte {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	return sm.data[key]
}

type MemoryBackend struct {
	key string
	off int
	em *SharedMemory
}

func (m *MemoryBackend) Write(b []byte) (int, error) {
	m.em.lock.Lock()
	defer m.em.lock.Unlock()

	m.em.data[m.key] = append(m.em.data[m.key], b...)
	return len(b), nil
}

func (m *MemoryBackend) Retrieve() (io.ReadCloser, error) {
	return m, nil
}

func (m *MemoryBackend) Cancel() error {
	m.em.lock.Lock()
	defer m.em.lock.Unlock()

	delete(m.em.data, m.key)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *MemoryBackend) Read(b []byte) (int, error) {
	m.em.lock.Lock()
	defer m.em.lock.Unlock()

	if n := min(len(m.em.data[m.key]) - m.off, len(b)); n > 0 {
		copy(b, m.em.data[m.key][m.off:m.off+n])
		m.off += n
		return n, nil
	}

	return 0, io.EOF
}

func (m *MemoryBackend) Close() error {
	return nil
}

var _ = Describe("API Streams", func() {
	mem := SharedMemory{
		data: make(map[string] []byte),
	}
	builder := func (k string) backend.Backend {
		return &MemoryBackend{
			em: &mem,
			key: k,
		}
	}

	Describe("Authorization", func() {
		It("should authorize a matched token", func() {
			s, err := api.NewStream("path/to/file", builder)
			Ω(err).ShouldNot(HaveOccurred())
			err = s.Lease(10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(s.Authorize(s.Token())).Should(BeTrue())
		})

		It("should not authorize a mismatched token", func() {
			s, err := api.NewStream("path/to/file", builder)
			Ω(err).ShouldNot(HaveOccurred())
			err = s.Lease(10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(s.Authorize("bad:" + s.Token())).Should(BeFalse())
		})

		It("should not authorize an expired token", func() {
			s, err := api.NewStream("path/to/file", builder)
			Ω(err).ShouldNot(HaveOccurred())
			err = s.Lease(-8*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(s.Authorize(s.Token())).Should(BeFalse())
		})
	})

	Describe("Uploads", func() {
		var s api.Stream

		BeforeEach(func() {
			var err error
			s, err = api.NewStream("test/path/to/file", builder)
			Ω(err).ShouldNot(HaveOccurred())
			err = s.Lease(10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should be able to upload files piecemeal", func() {
			_, err := s.AuthorizedWrite(s.Token(), []byte("this is the first line\n"))
			Ω(err).ShouldNot(HaveOccurred())

			_, err = s.AuthorizedWrite(s.Token(), []byte("this is the second line\n"))
			Ω(err).ShouldNot(HaveOccurred())

			b := mem.Get(s.Path)
			Ω(string(b)).Should(Equal("this is the first line\nthis is the second line\n"))
		})

		It("should not upload file chunks with a bad token", func() {
			_, err := s.AuthorizedWrite("INVALID TOKEN", []byte("this should never get written\n"))
			Ω(err).Should(HaveOccurred())

			_, err = ioutil.ReadFile(s.Path)
			Ω(err).Should(HaveOccurred())
			Ω(os.IsNotExist(err)).Should(BeTrue())
		})

		It("should not upload file chunks for an expired token", func() {
			s.Lease(-1 * time.Second)
			_, err := s.AuthorizedWrite(s.Token(), []byte("this should never get written\n"))
			Ω(err).Should(HaveOccurred())

			_, err = ioutil.ReadFile(s.Path)
			Ω(err).Should(HaveOccurred())
			Ω(os.IsNotExist(err)).Should(BeTrue())
		})
	})

	Describe("Downloads", func() {
		var uploader api.Stream

		BeforeEach(func() {
			var err error
			uploader, err = api.NewStream("download/file", builder)
			Ω(err).ShouldNot(HaveOccurred())
			err = uploader.Lease(10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())
			_, err = uploader.AuthorizedWrite(uploader.Token(), []byte("this is the first line\n"))
			Ω(err).ShouldNot(HaveOccurred())
			_, err = uploader.AuthorizedWrite(uploader.Token(), []byte("this is the second line\n"))
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should be able to download uploaded files", func() {
			s, err := api.NewStream(uploader.Path, builder)
			Ω(err).ShouldNot(HaveOccurred())
			err = s.Lease(10 * time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			out, err := s.AuthorizedRetrieve(s.Token())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(out).ShouldNot(BeNil())

			var b bytes.Buffer
			io.Copy(&b, out)
			Ω(b.String()).Should(Equal("this is the first line\nthis is the second line\n"))
		})

		It("should not download file chunks with a bad token", func() {
			s, err := api.NewStream(uploader.Path, builder)
			Ω(err).ShouldNot(HaveOccurred())
			err = s.Lease(10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			out, err := s.AuthorizedRetrieve("INVALID TOKEN")
			Ω(err).Should(HaveOccurred())
			Ω(out).Should(BeNil())
		})

		It("should not download file chunks for an expired token", func() {
			s, err := api.NewStream(uploader.Path, builder)
			Ω(err).ShouldNot(HaveOccurred())
			err = s.Lease(10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			s.Lease(-1 * time.Second)
			out, err := s.AuthorizedRetrieve(s.Token())
			Ω(err).Should(HaveOccurred())
			Ω(out).Should(BeNil())
		})
	})
})
