package meter_test

import (
	"bytes"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/jhunt/ssg/pkg/meter"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metered Reader / Writer Test Suite")
}

var _ = Describe("meter", func() {
	Context("reading", func() {
		var m *meter.Reader

		BeforeEach(func() {
			b := []byte("this is a test string")
			m = meter.NewReader(ioutil.NopCloser(bytes.NewBuffer(b)))
		})

		It("should have read 0 bytes initially", func() {
			Ω(m.Total()).Should(Equal(int64(0)))
			Ω(m.Delta()).Should(Equal(int64(0)))
		})

		It("should record 4 bytes read after Read([4]byte)", func() {
			to := make([]byte, 4)
			n, err := m.Read(to)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(n).Should(Equal(4))

			Ω(m.Total()).Should(Equal(int64(4)))
			Ω(m.Delta()).Should(Equal(int64(4)))
		})

		It("should reset the byte count on subsequent calls to Delta()", func() {
			to := make([]byte, 4)
			n, err := m.Read(to)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(n).Should(Equal(4))

			Ω(m.Total()).Should(Equal(int64(4)))
			Ω(m.Delta()).Should(Equal(int64(4)))
			Ω(m.Total()).Should(Equal(int64(4)))
			Ω(m.Delta()).Should(Equal(int64(0)))

			n, err = m.Read(to)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(n).Should(Equal(4))

			Ω(m.Total()).Should(Equal(int64(8)))
			Ω(m.Delta()).Should(Equal(int64(4)))
			Ω(m.Total()).Should(Equal(int64(8)))
			Ω(m.Delta()).Should(Equal(int64(0)))
		})
	})

	Context("writing", func() {
		var m *meter.Writer

		BeforeEach(func() {
			var b bytes.Buffer
			m = meter.NewWriter(&b)
		})

		It("should have written 0 bytes initially", func() {
			Ω(m.Total()).Should(Equal(int64(0)))
			Ω(m.Delta()).Should(Equal(int64(0)))
		})

		It("should record 4 bytes written after Write([4]byte)", func() {
			n, err := m.Write([]byte("test"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(n).Should(Equal(4))

			Ω(m.Total()).Should(Equal(int64(4)))
			Ω(m.Delta()).Should(Equal(int64(4)))
		})

		It("should reset the byte count on subsequent calls to Delta()", func() {
			n, err := m.Write([]byte("test"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(n).Should(Equal(4))

			Ω(m.Total()).Should(Equal(int64(4)))
			Ω(m.Delta()).Should(Equal(int64(4)))
			Ω(m.Total()).Should(Equal(int64(4)))
			Ω(m.Delta()).Should(Equal(int64(0)))

			n, err = m.Write([]byte("again"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(n).Should(Equal(5))

			Ω(m.Total()).Should(Equal(int64(9)))
			Ω(m.Delta()).Should(Equal(int64(5)))
			Ω(m.Total()).Should(Equal(int64(9)))
			Ω(m.Delta()).Should(Equal(int64(0)))
		})
	})
})
