package config

var Default Config

func init() {
	Default.Bind = ":8080"
	Default.MaxLease = 600
	Default.SweepInterval = 1
	Default.Metrics.ReservoirSize = 100
	Default.DefaultBucket.Compression = "none"
	Default.DefaultBucket.Encryption = "aes256-ctr"
}
