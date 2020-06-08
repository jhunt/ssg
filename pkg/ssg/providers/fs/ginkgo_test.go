package fs_test

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/providers/fs"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FS Provider Test Suite")
}

var _ = Describe("FS Provider", func() {
	Context("configuration", func() {
		It("should return an error if we try to use a non-existent root directory", func() {
			Ω("/path/to/nowhere").ShouldNot(BeADirectory())
			_, err := fs.Configure("/path/to/nowhere")
			Ω(err).Should(HaveOccurred())
		})

		It("should return an error if we try to use a non-directory root directory", func() {
			Ω("test/file").Should(BeARegularFile())
			_, err := fs.Configure("test/file")
			Ω(err).Should(HaveOccurred())
			Ω(err).Should(MatchError("test/file: not a directory"))
		})

		It("should succeed with a valid root directory", func() {
			provider, err := fs.Configure("test/")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(provider.Root).Should(Equal("test"))
		})
	})

	Context("uploading files", func() {
		var provider fs.Provider

		BeforeEach(func() {
			os.RemoveAll("test/root")
			os.Mkdir("test/root", 0777)

			p, err := fs.Configure("test/root")
			Ω(err).ShouldNot(HaveOccurred())
			provider = p
		})

		Context("without a specific path in mind", func() {
			It("should make up a random path", func() {
				uploader, err := provider.Upload("")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(uploader.Path()).Should(MatchRegexp(`.{4}/.{4}/.{16}/.{24}`))
			})

			It("should create the intervening directories", func() {
				uploader, err := provider.Upload("")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(filepath.Join(provider.Root, path.Dir(uploader.Path()))).Should(BeADirectory())
			})

			It("should create the uploaded file before anything is written to it", func() {
				uploader, err := provider.Upload("")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(filepath.Join(provider.Root, uploader.Path())).Should(BeARegularFile())
			})
		})

		Context("with a specific path in mind", func() {
			file := "a/test/file"

			It("should honor the path hint if it can", func() {
				uploader, err := provider.Upload(file)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(uploader.Path()).Should(Equal(file))
			})

			It("should create the intervening directories", func() {
				_, err := provider.Upload(file)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(filepath.Join(provider.Root, path.Dir(file))).Should(BeADirectory())
			})

			It("should create the uploaded file before anything is written to it", func() {
				_, err := provider.Upload(file)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(filepath.Join(provider.Root, file)).Should(BeARegularFile())
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
				uploader, err := provider.Upload(fs.RandomFile)
				Ω(err).ShouldNot(HaveOccurred())

				fmt.Fprintf(uploader, "this is a line\n")
				uploader.Close()

				b, err := ioutil.ReadFile(filepath.Join(provider.Root, uploader.Path()))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(b)).Should(Equal("this is a line\n"))
			})

			It("can handle multiple, subsequent writes", func() {
				uploader, err := provider.Upload(fs.RandomFile)
				Ω(err).ShouldNot(HaveOccurred())

				fmt.Fprintf(uploader, "this is a line\n")
				fmt.Fprintf(uploader, "this is another line\n")
				fmt.Fprintf(uploader, "this is yet another line\n")
				uploader.Close()

				b, err := ioutil.ReadFile(filepath.Join(provider.Root, uploader.Path()))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(string(b)).Should(Equal("this is a line\n" +
					"this is another line\n" +
					"this is yet another line\n"))
			})
		})

		Context("with a change of heart", func() {
			It("should remove the partially written file", func() {
				uploader, err := provider.Upload(fs.RandomFile)
				Ω(err).ShouldNot(HaveOccurred())

				fmt.Fprintf(uploader, "this is a line\n")
				uploader.Cancel()

				_, err = os.Stat(filepath.Join(provider.Root, uploader.Path()))
				Ω(err).Should(HaveOccurred())
			})
		})
	})

	Context("downloading files", func() {
		var provider fs.Provider

		BeforeEach(func() {
			os.RemoveAll("test/root")
			os.Mkdir("test/root", 0777)

			p, err := fs.Configure("test/root")
			Ω(err).ShouldNot(HaveOccurred())
			provider = p
		})

		It("should require a real path in the constructor", func() {
			_, err := provider.Download("")
			Ω(err).Should(HaveOccurred())

			_, err = provider.Download(fs.RandomFile)
			Ω(err).Should(HaveOccurred())
		})

		It("should fail to download a non-existent file", func() {
			Ω(filepath.Join(provider.Root, "a/test/file")).ShouldNot(BeARegularFile())
			_, err := provider.Download("a/test/file")
			Ω(err).Should(HaveOccurred())
		})

		It("should fail to download a directory", func() {
			err := os.Mkdir(filepath.Join(provider.Root, "dir"), 0777)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(filepath.Join(provider.Root, "dir")).Should(BeADirectory())
			_, err = provider.Download("dir")
			Ω(err).Should(HaveOccurred())
		})

		It("should download files when it can", func() {
			err := ioutil.WriteFile(filepath.Join(provider.Root, "file"), []byte("a test file\n"), 0666)
			Ω(err).ShouldNot(HaveOccurred())

			downloader, err := provider.Download("file")
			Ω(err).ShouldNot(HaveOccurred())

			b, err := ioutil.ReadAll(downloader)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(string(b)).Should(Equal("a test file\n"))
		})
	})

	Context("full stack", func() {
		var provider fs.Provider

		BeforeEach(func() {
			os.RemoveAll("test/root")
			os.Mkdir("test/root", 0777)

			p, err := fs.Configure("test/root")
			Ω(err).ShouldNot(HaveOccurred())
			provider = p
		})

		It("should be able to download an uploaded file", func() {
			uploader, err := provider.Upload(fs.RandomFile)
			Ω(err).ShouldNot(HaveOccurred())

			fmt.Fprintf(uploader, "The best laid schemes o’ Mice an’ Men\n")
			fmt.Fprintf(uploader, "  Gang aft agley.\n")
			uploader.Close()

			Ω(filepath.Join(provider.Root, uploader.Path())).Should(BeARegularFile())
			downloader, err := provider.Download(uploader.Path())
			Ω(err).ShouldNot(HaveOccurred())

			b, err := ioutil.ReadAll(downloader)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(string(b)).Should(Equal("The best laid schemes o’ Mice an’ Men\n" +
				"  Gang aft agley.\n"))

			err = provider.Expunge(uploader.Path())
			Ω(err).ShouldNot(HaveOccurred())

			Ω(filepath.Join(provider.Root, uploader.Path())).ShouldNot(BeARegularFile())
		})

		It("should be able to handle really large files", func() {
			uploader, err := provider.Upload(fs.RandomFile)
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
