package mem_test

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/jhunt/ssg/pkg/ssg/providers/mem"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mem Provider Test Suite")
}

var _ = Describe("Mem Provider", func() {
	Context("uploading files", func() {
		var provider mem.Provider

		BeforeEach(func() {
			p, err := mem.Configure()
			Ω(err).ShouldNot(HaveOccurred())
			provider = p
		})

		Context("without a specific path in mind", func() {
			It("should make up a random path", func() {
				uploader, err := provider.Upload("")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(uploader.Path()).Should(MatchRegexp(`.{4}/.{4}/.{16}/.{24}`))
			})
		})

		Context("with a specific path in mind", func() {
			file := "a/test/file"

			It("should honor the path hint if it can", func() {
				uploader, err := provider.Upload(file)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(uploader.Path()).Should(Equal(file))
			})

			It("should error if the path hint already exists", func() {
				uploader, err := provider.Upload(file)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(uploader.Path()).Should(Equal(file))

				_, err = provider.Upload(file)
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("with actual data in them", func() {
			It("should write and close the file", func() {
				uploader, err := provider.Upload(mem.RandomFile)
				Ω(err).ShouldNot(HaveOccurred())

				fmt.Fprintf(uploader, "this is a line\n")
				uploader.Close()

				buf, ok := provider.Files[uploader.Path()]
				Ω(ok).Should(BeTrue())
				Ω(buf.String()).Should(Equal("this is a line\n"))
			})

			It("can handle multiple, subsequent writes", func() {
				uploader, err := provider.Upload(mem.RandomFile)
				Ω(err).ShouldNot(HaveOccurred())

				fmt.Fprintf(uploader, "this is a line\n")
				fmt.Fprintf(uploader, "this is another line\n")
				fmt.Fprintf(uploader, "this is yet another line\n")
				uploader.Close()

				buf, ok := provider.Files[uploader.Path()]
				Ω(ok).Should(BeTrue())
				Ω(buf.String()).Should(Equal("this is a line\n" +
					"this is another line\n" +
					"this is yet another line\n"))
			})
		})

		Context("with a change of heart", func() {
			It("should remove the partially written file", func() {
				uploader, err := provider.Upload(mem.RandomFile)
				Ω(err).ShouldNot(HaveOccurred())

				fmt.Fprintf(uploader, "this is a line\n")
				uploader.Cancel()

				_, ok := provider.Files[uploader.Path()]
				Ω(ok).Should(BeFalse())
			})
		})
	})

	Context("downloading files", func() {
		var provider mem.Provider

		BeforeEach(func() {
			p, err := mem.Configure()
			Ω(err).ShouldNot(HaveOccurred())
			provider = p
		})

		It("should require a real path in the constructor", func() {
			_, err := provider.Download("")
			Ω(err).Should(HaveOccurred())

			_, err = provider.Download(mem.RandomFile)
			Ω(err).Should(HaveOccurred())
		})

		It("should fail to download a non-existent file", func() {
			_, ok := provider.Files["a/test/file"]
			Ω(ok).Should(BeFalse())
			_, err := provider.Download("a/test/file")
			Ω(err).Should(HaveOccurred())
		})

		It("should download files when it can", func() {
			buf := bytes.NewBuffer([]byte("a test file\n"))
			provider.Files["file"] = buf

			downloader, err := provider.Download("file")
			Ω(err).ShouldNot(HaveOccurred())

			b, err := ioutil.ReadAll(downloader)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(b)).Should(Equal("a test file\n"))
		})
	})

	Context("full stack", func() {
		var provider mem.Provider

		BeforeEach(func() {
			p, err := mem.Configure()
			Ω(err).ShouldNot(HaveOccurred())
			provider = p
		})

		It("should be able to download an uploaded file", func() {
			uploader, err := provider.Upload(mem.RandomFile)
			Ω(err).ShouldNot(HaveOccurred())

			fmt.Fprintf(uploader, "The best laid schemes o’ Mice an’ Men\n")
			fmt.Fprintf(uploader, "  Gang aft agley.\n")
			uploader.Close()

			_, ok := provider.Files[uploader.Path()]
			Ω(ok).Should(BeTrue())

			downloader, err := provider.Download(uploader.Path())
			Ω(err).ShouldNot(HaveOccurred())

			b, err := ioutil.ReadAll(downloader)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(string(b)).Should(Equal("The best laid schemes o’ Mice an’ Men\n" +
				"  Gang aft agley.\n"))

			err = provider.Expunge(uploader.Path())
			Ω(err).ShouldNot(HaveOccurred())

			_, ok = provider.Files[uploader.Path()]
			Ω(ok).Should(BeFalse())
		})

		It("should be able to handle really large files", func() {
			uploader, err := provider.Upload(mem.RandomFile)
			Ω(err).ShouldNot(HaveOccurred())

			// generate 10M of data
			// checksum 872e2c6727b8e809cbe5baf15f05997753cd7818
			for i := 0; i < 10240; i++ {
				b := make([]byte, 1024)
				fill := []byte("jrh")
				for j := range b {
					b[j] = fill[j%3]
				}
				uploader.Write(b)
			}
			uploader.Close()

			downloader, err := provider.Download(uploader.Path())
			Ω(err).ShouldNot(HaveOccurred())

			ck := sha1.New()
			io.Copy(ck, downloader)
			Ω(hex.EncodeToString(ck.Sum(nil))).Should(Equal("872e2c6727b8e809cbe5baf15f05997753cd7818"))
		})
	})
})
