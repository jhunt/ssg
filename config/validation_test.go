package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/jhunt/shield-storage-gateway/config"
)

var _ = Describe("Configuration", func() {
	Describe("Parsing", func() {
		It("should read a valid, explicit configuration", func() {
			c, err := config.Read([]byte(`---
cluster: test
bind: 127.0.0.1:8081
controlTokens:
  - foo
defaultBucket:
  compression: none
  encryption:  none
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
    encryption:  none
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
        prefix: backups/ci/cd/

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c.Cluster).Should(Equal("test"))
			Ω(c.Bind).Should(Equal("127.0.0.1:8081"))
			Ω(c.ControlTokens).Should(Equal([]string{"foo"}))

			Ω(c.DefaultBucket.Compression).Should(Equal("none"))
			Ω(c.DefaultBucket.Encryption).Should(Equal("none"))

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
			Ω(c.Buckets[0].Encryption).Should(Equal("none"))
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
			Ω(c.Buckets[0].Provider.S3.Prefix).Should(Equal("backups/ci/cd/"))
			Ω(c.Buckets[0].Provider.S3.AccessKeyID).Should(Equal("AKI-EXAMPLE-KEY"))
			Ω(c.Buckets[0].Provider.S3.SecretAccessKey).Should(Equal("SECRET-KEY"))
			Ω(c.Buckets[0].Provider.S3.InstanceMetadata).Should(BeFalse())
		})

		It("should read a valid, implicit configuration", func() {
			c, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
        prefix: backups/ci/cd/

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c.Cluster).Should(Equal("test"))
			Ω(c.Bind).Should(Equal("*:8080")) // default
			Ω(c.ControlTokens).Should(Equal([]string{"foo"}))

			Ω(c.DefaultBucket.Compression).Should(Equal("none")) // default
			Ω(c.DefaultBucket.Encryption).Should(Equal("none"))  // default

			Ω(c.DefaultBucket.Vault).ShouldNot(BeNil())
			Ω(c.DefaultBucket.Vault.Kind).Should(Equal("hashicorp"))
			Ω(c.DefaultBucket.Vault.Hashicorp.URL).Should(Equal("https://127.0.0.1:8200"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Prefix).Should(Equal("secret/shared/ssg/test"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Token).Should(Equal("s.ThIsIsNeXaMpLeToKeN"))
			Ω(c.DefaultBucket.Vault.Hashicorp.Role).Should(Equal(""))
			Ω(c.DefaultBucket.Vault.Hashicorp.Secret).Should(Equal(""))

			Ω(len(c.Buckets)).Should(Equal(1))
			Ω(c.Buckets[0].Key).Should(Equal("store"))
			Ω(c.Buckets[0].Name).Should(Equal("store")) // inferred
			Ω(c.Buckets[0].Description).Should(Equal(""))
			Ω(c.Buckets[0].Compression).Should(Equal("none"))
			Ω(c.Buckets[0].Encryption).Should(Equal("none"))
			Ω(c.Buckets[0].Vault).ShouldNot(BeNil())
			Ω(c.Buckets[0].Vault.Kind).Should(Equal("hashicorp"))
			Ω(c.Buckets[0].Vault.Hashicorp.URL).Should(Equal("https://127.0.0.1:8200"))
			Ω(c.Buckets[0].Vault.Hashicorp.Prefix).Should(Equal("secret/shared/ssg/test"))
			Ω(c.Buckets[0].Vault.Hashicorp.Token).Should(Equal("s.ThIsIsNeXaMpLeToKeN"))
			Ω(c.Buckets[0].Vault.Hashicorp.Role).Should(Equal(""))
			Ω(c.Buckets[0].Vault.Hashicorp.Secret).Should(Equal(""))
			Ω(c.Buckets[0].Provider.Kind).Should(Equal("s3"))

			Ω(c.Buckets[0].Provider.FS).Should(BeNil())
			Ω(c.Buckets[0].Provider.WebDAV).Should(BeNil())

			Ω(c.Buckets[0].Provider.S3).ShouldNot(BeNil())
			Ω(c.Buckets[0].Provider.S3.Region).Should(Equal("us-east-1"))
			Ω(c.Buckets[0].Provider.S3.Bucket).Should(Equal("some-storage-bucket-in-s3"))
			Ω(c.Buckets[0].Provider.S3.Prefix).Should(Equal("backups/ci/cd/"))
			Ω(c.Buckets[0].Provider.S3.AccessKeyID).Should(Equal("AKI-EXAMPLE-KEY"))
			Ω(c.Buckets[0].Provider.S3.SecretAccessKey).Should(Equal("SECRET-KEY"))
			Ω(c.Buckets[0].Provider.S3.InstanceMetadata).Should(BeFalse())
		})

		It("should read all provider configurations", func() {
			c, err := config.Read([]byte(`---
cluster: test
bind: 127.0.0.1:8081
controlTokens:
  - foo
defaultBucket:
  compression: none
  encryption:  none
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: in-s3
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
        prefix: backups/ci/cd/

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY

  - key: in-fs
    provider:
      kind: fs
      fs:
        root: /data

  - key: in-webdav
    provider:
      kind: webdav
      webdav:
        url: https://store1.example.com:9000
        basicAuth:
          username: admin
          password: foo-sekrit
`))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(len(c.Buckets)).Should(Equal(3))

			Ω(c.Buckets[0].Key).Should(Equal("in-s3"))
			Ω(c.Buckets[0].Provider.Kind).Should(Equal("s3"))
			Ω(c.Buckets[0].Provider.FS).Should(BeNil())
			Ω(c.Buckets[0].Provider.WebDAV).Should(BeNil())

			Ω(c.Buckets[0].Provider.S3).ShouldNot(BeNil())
			Ω(c.Buckets[0].Provider.S3.Region).Should(Equal("us-east-1"))
			Ω(c.Buckets[0].Provider.S3.Bucket).Should(Equal("some-storage-bucket-in-s3"))
			Ω(c.Buckets[0].Provider.S3.Prefix).Should(Equal("backups/ci/cd/"))
			Ω(c.Buckets[0].Provider.S3.AccessKeyID).Should(Equal("AKI-EXAMPLE-KEY"))
			Ω(c.Buckets[0].Provider.S3.SecretAccessKey).Should(Equal("SECRET-KEY"))
			Ω(c.Buckets[0].Provider.S3.InstanceMetadata).Should(BeFalse())

			Ω(c.Buckets[1].Key).Should(Equal("in-fs"))
			Ω(c.Buckets[1].Provider.Kind).Should(Equal("fs"))
			Ω(c.Buckets[1].Provider.S3).Should(BeNil())
			Ω(c.Buckets[1].Provider.WebDAV).Should(BeNil())

			Ω(c.Buckets[1].Provider.FS).ShouldNot(BeNil())
			Ω(c.Buckets[1].Provider.FS.Root).Should(Equal("/data"))

			Ω(c.Buckets[2].Key).Should(Equal("in-webdav"))
			Ω(c.Buckets[2].Provider.Kind).Should(Equal("webdav"))
			Ω(c.Buckets[2].Provider.S3).Should(BeNil())
			Ω(c.Buckets[2].Provider.FS).Should(BeNil())

			Ω(c.Buckets[2].Provider.WebDAV).ShouldNot(BeNil())
			Ω(c.Buckets[2].Provider.WebDAV.URL).Should(Equal("https://store1.example.com:9000"))
			Ω(c.Buckets[2].Provider.WebDAV.BasicAuth.Username).Should(Equal("admin"))
			Ω(c.Buckets[2].Provider.WebDAV.BasicAuth.Password).Should(Equal("foo-sekrit"))
		})

		It("should allow vault to be omitted if encryption is none", func() {
			_, err := config.Read([]byte(`---
cluster: test
bind: 127.0.0.1:8081
controlTokens:
  - foo
defaultBucket:
  encryption:  none

buckets:
  - key: in-fs
    provider:
      kind: fs
      fs:
        root: /data
`))
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should fail if we omit a vault and want encryption", func() {
			_, err := config.Read([]byte(`---
cluster: test
bind: 127.0.0.1:8081
controlTokens:
  - foo
defaultBucket:
  encryption:  aes256-ctr

buckets:
  - key: in-fs
    provider:
      kind: fs
      fs:
        root: /data
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we give it something that isn't YAML", func() {
			_, err := config.Read([]byte(`this is not YAML!!`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the cluster directive", func() {
			_, err := config.Read([]byte(`---
#cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
        prefix: backups/ci/cd/

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the control tokens directive", func() {
			_, err := config.Read([]byte(`---
cluster: test
#controlTokens:
#  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
        prefix: backups/ci/cd/

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget provide control tokens", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens: []
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
        prefix: backups/ci/cd/

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the buckets directive", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify no buckets", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN
buckets: []
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the default vault URL directive", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
#      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the default vault prefix directive", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
#      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the default vault token directive", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
#      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget specify an invalid default vault kind", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: magic

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify an invalid default compression mechanism", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  compression: magic
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN
      role:   this-is-re-dund-ant
      secret: 12345

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify an invalid default encryption mechanism", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  encryption: magic
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN
      role:   this-is-re-dund-ant
      secret: 12345

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify mutually exclusive vault auth mechanisms", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN
      role:   this-is-re-dund-ant
      secret: 12345

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the bucket key", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
        prefix: backups/ci/cd/

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the s3 configuration altogether", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the s3 secret access key", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
#        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify mutually exlusive s3 auth mechanisms", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
        instanceMetadata: true
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the s3 auth mechanisms", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the s3 region", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
#        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the s3 bucket", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
#        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the s3 access key id", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

#        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the s3 secret access key", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
#        secretAccessKey: SECRET-KEY
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify mutually exlusive s3 auth mechanisms", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3

        accessKeyID:     AKI-EXAMPLE-KEY
        secretAccessKey: SECRET-KEY
        instanceMetadata: true
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the s3 auth mechanisms", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: s3
      s3:
        region: us-east-1
        bucket: some-storage-bucket-in-s3
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the fs configuration altogether", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: fs
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the fs root path", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: fs
      fs: {}
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we supply a relative root path", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: fs
      fs:
        root: ./files
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the webdav configuration altogether", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: webdav
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we forget the webdav url", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: webdav
      webdav: {}
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify an unparseable webdav url", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: webdav
      webdav:
        url: 🎉
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify a non-HTTP(s) webdav url", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: webdav
      webdav:
        url: mailto:foo@example.com
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify an FTP webdav url", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    provider:
      kind: webdav
      webdav:
        url: ftp://foo.example.com
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify an invalid bucket compression algorithm", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    compression: magic
    provider:
      kind: fs
      fs:
        root: /tmp
`))
			Ω(err).Should(HaveOccurred())
		})

		It("should fail if we specify an invalid bucket encryption algorithm", func() {
			_, err := config.Read([]byte(`---
cluster: test
controlTokens:
  - foo
defaultBucket:
  vault:
    kind: hashicorp
    hashicorp:
      url:    https://127.0.0.1:8200
      prefix: secret/shared/ssg/test
      token:  s.ThIsIsNeXaMpLeToKeN

buckets:
  - key: store
    encryption: magic
    provider:
      kind: fs
      fs:
        root: /tmp
`))
			Ω(err).Should(HaveOccurred())
		})
	})
})
