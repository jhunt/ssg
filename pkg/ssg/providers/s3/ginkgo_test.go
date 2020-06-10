package s3_test

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

	"github.com/jhunt/ssg/pkg/ssg/providers/s3"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3 Provider Test Suite")
}

var _ = Describe("S3 Provider", func() {
	Context("full stack", func() {
		var provider s3.Provider

		BeforeEach(func() {
			if v := os.Getenv("TEST_S3_LIVE"); v != "yes" {
				Skip("TEST_S3_LIVE not found in environment")
				return
			}

			Ω(os.Getenv("TEST_S3_BUCKET")).ShouldNot(Equal(""))
			Ω(os.Getenv("TEST_S3_REGION")).ShouldNot(Equal(""))
			Ω(os.Getenv("TEST_S3_AKI")).ShouldNot(Equal(""))
			Ω(os.Getenv("TEST_S3_KEY")).ShouldNot(Equal(""))

			p, err := s3.Configure(s3.Endpoint{
				Bucket:          os.Getenv("TEST_S3_BUCKET"),
				Region:          os.Getenv("TEST_S3_REGION"),
				AccessKeyID:     os.Getenv("TEST_S3_AKI"),
				SecretAccessKey: os.Getenv("TEST_S3_KEY"),
				Prefix:          os.Getenv("TEST_S3_PREFIX"),
			})
			Ω(err).ShouldNot(HaveOccurred())
			provider = p
		})

		It("should be able to download an uploaded file", func() {
			uploader, err := provider.Upload(s3.RandomKey)
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
			uploader, err := provider.Upload(s3.RandomKey)
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
