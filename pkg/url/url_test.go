package url_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jhunt/shield-storage-gateway/pkg/url"
)

var _ = Describe("SSG URLs", func() {
	Context("a valid SSG URL with a simple path", func() {
		It("should parse", func() {
			u, err := url.Parse("ssg://cluster/bucket/simple-path")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(u.Cluster).Should(Equal("cluster"))
			Ω(u.Bucket).Should(Equal("bucket"))
			Ω(u.Path).Should(Equal("/simple-path"))
		})
	})

	Context("a valid SSG URL without a path", func() {
		It("should parse with the trailing slash", func() {
			u, err := url.Parse("ssg://cluster/bucket/")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(u.Cluster).Should(Equal("cluster"))
			Ω(u.Bucket).Should(Equal("bucket"))
			Ω(u.Path).Should(Equal(""))
		})
		It("should parse without the trailing slash", func() {
			u, err := url.Parse("ssg://cluster/bucket")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(u.Cluster).Should(Equal("cluster"))
			Ω(u.Bucket).Should(Equal("bucket"))
			Ω(u.Path).Should(Equal(""))
		})
	})

	Context("a valid SSG URL without a bucket or path", func() {
		It("should parse with the trailing slash", func() {
			u, err := url.Parse("ssg://cluster/")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(u.Cluster).Should(Equal("cluster"))
			Ω(u.Bucket).Should(Equal(""))
			Ω(u.Path).Should(Equal(""))
		})
		It("should parse without the trailing slash", func() {
			u, err := url.Parse("ssg://cluster")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(u.Cluster).Should(Equal("cluster"))
			Ω(u.Bucket).Should(Equal(""))
			Ω(u.Path).Should(Equal(""))
		})
	})

	Context("an invalid SSG URL with a bucket but no cluster", func() {
		It("should not parse with the trailing slash", func() {
			_, err := url.Parse("ssg:///bucket/")
			Ω(err).Should(HaveOccurred())
		})
		It("should not parse without the trailing slash", func() {
			_, err := url.Parse("ssg:///bucket")
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("an HTTP URL", func() {
		It("should not parse as a valid SSG URL", func() {
			_, err := url.Parse("http://example.com")
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("a crafted URL object", func() {
		var u url.URL

		BeforeEach(func() {
			u = url.URL{
				Cluster: "test1",
				Bucket:  "backups",
				Path:    "/prod/snapshots/postgres/DECA-FBAD",
			}
		})

		It("should stringify", func() {
			Ω(u.String()).Should(Equal("ssg://test1/backups/prod/snapshots/postgres/DECA-FBAD"))
		})

		Context("with a path that doesn't start with a forward slash", func() {
			It("inserts the missing leading forward slash", func() {
				u.Path = "a/relative/path"
				Ω(u.String()).Should(Equal("ssg://test1/backups/a/relative/path"))
				Ω(u.Path).Should(Equal("a/relative/path"))
			})
		})

		Context("with a path that ends with a forward slash", func() {
			It("omits the trailing forward slash", func() {
				u.Path = "a/trailing/slash/"
				Ω(u.String()).Should(Equal("ssg://test1/backups/a/trailing/slash"))
				Ω(u.Path).Should(Equal("a/trailing/slash/"))
			})
		})
	})
})
