package api_test

import (
	"bytes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/shieldproject/shield-storage-gateway/api"
)

var _ = Describe("API Streams", func() {
	Describe("Authorization", func() {
		It("should authorize a matched token", func() {
			s, err := api.NewUploadStream("path/to/file", 10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(s.Authorize(s.Token.Secret)).Should(BeTrue())
		})

		It("should not authorize a mismatched token", func() {
			s, err := api.NewUploadStream("path/to/file", 10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(s.Authorize("bad:" + s.Token.Secret)).Should(BeFalse())
		})

		It("should not authorize an expired token", func() {
			s, err := api.NewUploadStream("path/to/file", 10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())
			s.Token.Expire()

			Ω(s.Authorize(s.Token.Secret)).Should(BeFalse())
		})
	})

	Describe("Uploads", func() {
		var dir string
		var s api.Stream

		BeforeEach(func() {
			var err error
			dir, err = ioutil.TempDir("", "apitest")
			Ω(err).ShouldNot(HaveOccurred())

			s, err = api.NewUploadStream(dir+"/test/path/to/file", 10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should be able to upload files piecemeal", func() {
			defer os.RemoveAll(dir)

			_, err := s.UploadChunk(s.Token.Secret, []byte("this is the first line\n"))
			Ω(err).ShouldNot(HaveOccurred())

			_, err = s.UploadChunk(s.Token.Secret, []byte("this is the second line\n"))
			Ω(err).ShouldNot(HaveOccurred())

			b, err := ioutil.ReadFile(s.Path)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(string(b)).Should(Equal("this is the first line\nthis is the second line\n"))
		})

		It("should not upload file chunks with a bad token", func() {
			defer os.RemoveAll(dir)

			_, err := s.UploadChunk("INVALID TOKEN", []byte("this should never get written\n"))
			Ω(err).Should(HaveOccurred())

			_, err = ioutil.ReadFile(s.Path)
			Ω(err).Should(HaveOccurred())
			Ω(os.IsNotExist(err)).Should(BeTrue())
		})

		It("should not upload file chunks for an expired token", func() {
			defer os.RemoveAll(dir)

			s.Token.Expire()
			_, err := s.UploadChunk(s.Token.Secret, []byte("this should never get written\n"))
			Ω(err).Should(HaveOccurred())

			_, err = ioutil.ReadFile(s.Path)
			Ω(err).Should(HaveOccurred())
			Ω(os.IsNotExist(err)).Should(BeTrue())
		})
	})

	Describe("Downloads", func() {
		var dir string
		var uploader api.Stream

		BeforeEach(func() {
			var err error
			dir, err = ioutil.TempDir("", "apitest")
			Ω(err).ShouldNot(HaveOccurred())

			uploader, err = api.NewUploadStream(dir+"/file", 10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())
			_, err = uploader.UploadChunk(uploader.Token.Secret, []byte("this is the first line\n"))
			Ω(err).ShouldNot(HaveOccurred())
			_, err = uploader.UploadChunk(uploader.Token.Secret, []byte("this is the second line\n"))
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should be able to download uploaded files", func() {
			defer os.RemoveAll(dir)

			s, err := api.NewDownloadStream(uploader.Path, 10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			out, err := s.Reader(s.Token.Secret)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(out).ShouldNot(BeNil())

			var b bytes.Buffer
			io.Copy(&b, out)
			Ω(b.String()).Should(Equal("this is the first line\nthis is the second line\n"))
		})

		It("should not download file chunks with a bad token", func() {
			defer os.RemoveAll(dir)

			s, err := api.NewDownloadStream(uploader.Path, 10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			out, err := s.Reader("INVALID TOKEN")
			Ω(err).Should(HaveOccurred())
			Ω(out).Should(BeNil())
		})

		It("should not download file chunks for an expired token", func() {
			defer os.RemoveAll(dir)

			s, err := api.NewDownloadStream(uploader.Path, 10*time.Minute)
			Ω(err).ShouldNot(HaveOccurred())

			s.Token.Expire()
			out, err := s.Reader(s.Token.Secret)
			Ω(err).Should(HaveOccurred())
			Ω(out).Should(BeNil())
		})
	})
})
