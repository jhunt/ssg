package config_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jhunt/ssg/pkg/ssg/config"
)

var _ = Describe("Configuration", func() {
	Describe("Parsing", func() {
		BeforeEach(func() {
			os.Setenv("VAULT_TOKEN", "s.ThIsIsAnExAmPlEtOkEn")
			os.Setenv("OTHER_VAULT_TOKEN", "s.SoMeOtHeReXaMpLeToKeN")
			os.Setenv("S3_AKI", "AKI-BUT-JUST-FOR-TESTING")
			os.Setenv("S3_KEY", "A-TEST-SECRET-KEY")
			os.Setenv("FIRST_HALF", "first-half")
			os.Setenv("SECOND_HALF", "second-half")
		})

		It("should read a valid, explicit configuration", func() {
			c, err := config.Read([]byte(`---
cluster: test
bind: 127.0.0.1:8081
controlTokens:
  - foo
defaultBucket:
  compression: zlib
  encryption:  aes256-ctr
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key:  standard-1
    name: Standard Cloud Storage
    description: |-
      Our standard Cloud Storage, hosted in S3.

    compression: none
    encryption:  aes192-cfb
    vault:
      kind: hashicorp
      hashicorp:
        url:    https://127.0.0.1:8200
        prefix: secret/shared/ssg/test2
        token:  s.AnOtHeReXaMpLeToKeN

    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
        prefix: backups/ci/cd/${FIRST_HALF}/${SECOND_HALF}

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c.Cluster).Should(Equal("test"))
			Ω(c.Bind).Should(Equal("127.0.0.1:8081"))
			Ω(c.ControlTokens).Should(Equal([]string{"foo"}))

			Ω(c.DefaultBucket.Compression).Should(Equal("zlib"))
			Ω(c.DefaultBucket.Encryption).Should(Equal("aes256-ctr"))

			Ω(c.DefaultBucket.Vault).ShouldNot(BeNil())
			Ω(c.DefaultBucket.Vault.Kind).Should(Equal("hashicorp"))
			Ω(c.DefaultBucket.Vault.Hashicorp.URL).Should(Equal("https://127.0.0.1:8200"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Prefix).Should(Equal("secret/shared/ssg/test"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Token).Should(Equal("s.ThIsIsNeXaMpLeToKeN"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Role).Should(Equal(""))
			Ω(c.DefaultBucket.Vault.Hashicorp.Secret).Should(Equal(""))

			Ω(len(c.Buckets)).Should(Equal(1))
			Ω(c.Buckets[0].Key).Should(Equal("standard-1"))
			Ω(c.Buckets[0].Name).Should(Equal("Standard Cloud Storage"))
			Ω(c.Buckets[0].Description).Should(Equal("Our standard Cloud Storage, hosted in S3."))
			Ω(c.Buckets[0].Compression).Should(Equal("none"))
			Ω(c.Buckets[0].Encryption).Should(Equal("aes192-cfb"))
			Ω(c.Buckets[0].Vault).ShouldNot(BeNil())
			Ω(c.Buckets[0].Vault.Kind).Should(Equal("hashicorp"))
			Ω(c.Buckets[0].Vault.Hashicorp.URL).Should(Equal("https://127.0.0.1:8200"))
			Ω(c.Buckets[0].Vault.Hashicorp.Prefix).Should(Equal("secret/shared/ssg/test2"))
			Ω(c.Buckets[0].Vault.Hashicorp.Token).Should(Equal("s.AnOtHeReXaMpLeToKeN"))
			Ω(c.Buckets[0].Vault.Hashicorp.Role).Should(Equal(""))
			Ω(c.Buckets[0].Vault.Hashicorp.Secret).Should(Equal(""))
			Ω(c.Buckets[0].Provider.Kind).Should(Equal("s3"))

			Ω(c.Buckets[0].Provider.FS).Should(BeNil())
			Ω(c.Buckets[0].Provider.WebDAV).Should(BeNil())

			Ω(c.Buckets[0].Provider.S3).ShouldNot(BeNil())
			Ω(c.Buckets[0].Provider.S3.Region).Should(Equal("us-east-1"))
			Ω(c.Buckets[0].Provider.S3.Bucket).Should(Equal("some-storage-bucket-in-s3"))
			Ω(c.Buckets[0].Provider.S3.Prefix).Should(Equal("backups/ci/cd/first-half/second-half"))
			Ω(c.Buckets[0].Provider.S3.AccessKeyID).Should(Equal("AKI-EXAMPLE-KEY"))
			Ω(c.Buckets[0].Provider.S3.SecretAccessKey).Should(Equal("SECRET-KEY"))
			Ω(c.Buckets[0].Provider.S3.InstanceMetadata).Should(BeFalse())
		})

		It("should resolve environment variables when asked", func() {
			c, err := config.Read([]byte(`---
cluster: test
bind: 127.0.0.1:8081
controlTokens:
  - foo
defaultBucket:
  compression: zlib
  encryption:  aes256-ctr
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  ${VAULT_TOKEN}

buckets:
  - key:  standard-1
    name: Standard Cloud Storage
    description: |-
      Our standard Cloud Storage, hosted in S3.

    compression: none
    encryption:  aes192-cfb
    vault:
      kind: hashicorp
      hashicorp:
        url:    https://127.0.0.1:8200
        prefix: secret/shared/ssg/test2
        token:  ${OTHER_VAULT_TOKEN}

    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
        prefix: backups/ci/cd/

        accessKeyID:     ${S3_AKI}
        secretAccessKey: ${S3_KEY}
`))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c.Cluster).Should(Equal("test"))
			Ω(c.Bind).Should(Equal("127.0.0.1:8081"))
			Ω(c.ControlTokens).Should(Equal([]string{"foo"}))

			Ω(c.DefaultBucket.Compression).Should(Equal("zlib"))
			Ω(c.DefaultBucket.Encryption).Should(Equal("aes256-ctr"))

			Ω(c.DefaultBucket.Vault).ShouldNot(BeNil())
			Ω(c.DefaultBucket.Vault.Kind).Should(Equal("hashicorp"))
			Ω(c.DefaultBucket.Vault.Hashicorp.URL).Should(Equal("https://127.0.0.1:8200"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Prefix).Should(Equal("secret/shared/ssg/test"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Token).Should(Equal("s.ThIsIsAnExAmPlEtOkEn"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Role).Should(Equal(""))
			Ω(c.DefaultBucket.Vault.Hashicorp.Secret).Should(Equal(""))

			Ω(len(c.Buckets)).Should(Equal(1))
			Ω(c.Buckets[0].Key).Should(Equal("standard-1"))
			Ω(c.Buckets[0].Name).Should(Equal("Standard Cloud Storage"))
			Ω(c.Buckets[0].Description).Should(Equal("Our standard Cloud Storage, hosted in S3."))
			Ω(c.Buckets[0].Compression).Should(Equal("none"))
			Ω(c.Buckets[0].Encryption).Should(Equal("aes192-cfb"))
			Ω(c.Buckets[0].Vault).ShouldNot(BeNil())
			Ω(c.Buckets[0].Vault.Kind).Should(Equal("hashicorp"))
			Ω(c.Buckets[0].Vault.Hashicorp.URL).Should(Equal("https://127.0.0.1:8200"))
			Ω(c.Buckets[0].Vault.Hashicorp.Prefix).Should(Equal("secret/shared/ssg/test2"))
			Ω(c.Buckets[0].Vault.Hashicorp.Token).Should(Equal("s.SoMeOtHeReXaMpLeToKeN"))
			Ω(c.Buckets[0].Vault.Hashicorp.Role).Should(Equal(""))
			Ω(c.Buckets[0].Vault.Hashicorp.Secret).Should(Equal(""))
			Ω(c.Buckets[0].Provider.Kind).Should(Equal("s3"))

			Ω(c.Buckets[0].Provider.FS).Should(BeNil())
			Ω(c.Buckets[0].Provider.WebDAV).Should(BeNil())

			Ω(c.Buckets[0].Provider.S3).ShouldNot(BeNil())
			Ω(c.Buckets[0].Provider.S3.Region).Should(Equal("us-east-1"))
			Ω(c.Buckets[0].Provider.S3.Bucket).Should(Equal("some-storage-bucket-in-s3"))
			Ω(c.Buckets[0].Provider.S3.Prefix).Should(Equal("backups/ci/cd/"))
			Ω(c.Buckets[0].Provider.S3.AccessKeyID).Should(Equal("AKI-BUT-JUST-FOR-TESTING"))
			Ω(c.Buckets[0].Provider.S3.SecretAccessKey).Should(Equal("A-TEST-SECRET-KEY"))
			Ω(c.Buckets[0].Provider.S3.InstanceMetadata).Should(BeFalse())
		})
	})
})
