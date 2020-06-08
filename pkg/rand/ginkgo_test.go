package rand_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/jhunt/shield-storage-gateway/pkg/rand"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Randomness Utilities Test Suite")
}

var _ = Describe("rand.String()", func() {
	Context("randomizing strings", func() {
		It("should honor the string length parameter", func() {
			for n := 1; n < 512; n++ {
				Ω(len(rand.String(n))).Should(Equal(n))
			}
		})
	})

	Context("randomizing paths", func() {
		It("should default to (4)/(4)/(16)/(48) for historic reasons", func() {
			Ω(rand.Path()).Should(MatchRegexp(`.{4}/.{4}/.{16}/.{48}`))
		})
	})
})
