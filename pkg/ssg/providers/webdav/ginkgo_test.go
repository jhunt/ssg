package webdav_test

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg/providers/webdav"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WEBDAV Provider Test Suite")
}

var _ = Describe("WEBDAV Provider", func() {
	Context("full stack", func() {
		var provider webdav.Provider

		BeforeEach(func() {
			if v := os.Getenv("TEST_WEBDAV_LIVE"); v != "yes" {
				Skip("TEST_WEBDAV_LIVE not found in environment")
				return
			}

			Ω(os.Getenv("TEST_WEBDAV_URL")).ShouldNot(Equal(""))

			p, err := webdav.Configure(webdav.Endpoint{
				URL:      os.Getenv("TEST_WEBDAV_URL"),
				Username: os.Getenv("TEST_WEBDAV_USERNAME"),
				Password: os.Getenv("TEST_WEBDAV_PASSWORD"),
			})
			Ω(err).ShouldNot(HaveOccurred())
			provider = p
		})

		It("should be able to download an uploaded file", func() {
			uploader, err := provider.Upload(webdav.RandomFile)
			Ω(err).ShouldNot(HaveOccurred())

			fmt.Fprintf(uploader, "The best laid schemes o’ Mice an’ Men\n")
			fmt.Fprintf(uploader, "  Gang aft agley.\n")
			uploader.Close()

			downloader, err := provider.Download(uploader.Path())
			Ω(err).ShouldNot(HaveOccurred())

			b, err := ioutil.ReadAll(downloader)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(string(b)).Should(Equal("The best laid schemes o’ Mice an’ Men\n" +
				"  Gang aft agley.\n"))

			err = provider.Expunge(uploader.Path())
			Ω(err).ShouldNot(HaveOccurred())

			_, err = provider.Download(uploader.Path())
			Ω(err).Should(HaveOccurred())
		})

		It("should be able to handle really large files", func() {
			uploader, err := provider.Upload(webdav.RandomFile)
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
