package config

var Default Config

func init() {
	Default.Bind = "*:8080"
	Default.DefaultBucket.Compression = "none"
	Default.DefaultBucket.Encryption = "none"
}