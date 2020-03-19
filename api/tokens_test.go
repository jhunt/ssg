package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"

	"github.com/shieldproject/shield-storage-gateway/api"
)

var _ = Describe("API Tokens", func() {
	Describe("Generating token secrets", func() {
		It("should be able to generate a new token, without error", func() {
			t, err := api.NewRandomString(16)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(t).ShouldNot(BeEmpty())
		})

		It("should create unique tokens", func() {
			t1, err := api.NewRandomString(16)
			Ω(err).ShouldNot(HaveOccurred())

			t2, err := api.NewRandomString(16)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(t1).ShouldNot(Equal(t2))
		})
	})

	Describe("Generating tokens", func() {
		It("should honor parametric lifetimes", func() {
			t, err := api.NewToken(15 * time.Second)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(t.Expired()).Should(BeFalse())
			t.Expires = time.Now().Add(-1 * time.Second)
			Ω(t.Expired()).Should(BeTrue())
		})

		It("can create already-expired tokens", func() {
			t, err := api.NewToken(-5 * time.Second)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(t.Expired()).Should(BeTrue())
		})

		It("can renew tokens", func() {
			t, err := api.NewToken(1 * time.Second)
			Ω(err).ShouldNot(HaveOccurred())

			// forcibly expire
			t.Expires = time.Now().Add(-1 * time.Second)
			Ω(t.Expired()).Should(BeTrue())

			t.Renew()
			Ω(t.Expired()).Should(BeFalse())
		})
	})
})
