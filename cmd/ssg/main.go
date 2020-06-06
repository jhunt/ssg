package main

import (
	"net/http"
	"os"
	"time"

	fmt "github.com/jhunt/go-ansi"

	"github.com/jhunt/go-cli"
	env "github.com/jhunt/go-envirotron"
	"github.com/jhunt/go-log"
	"github.com/jhunt/go-s3"

	"github.com/jhunt/shield-storage-gateway/api"
	"github.com/jhunt/shield-storage-gateway/vault"
)

var Version = ""

func sanitize(s string) string {
	runes := []rune(s)
	for i := 3; i < len(runes)-4; i++ {
		runes[i] = '-'
	}
	return string(runes)
}

func main() {
	var opts struct {
		Help    bool `cli:"-h, --help"`
		Version bool `cli:"-v, --version"`

		Debug bool   `cli:"-D, --debug"`
		Log   string `cli:"-l, --log-level" env:"SSG_LOG_LEVEL"`

		Listen string `cli:"--listen" env:"SSG_LISTEN"`

		Compression string `cli:"--compression" env:"SSG_COMPRESSION"`

		Encryption string `cli:"--encryption" env:"SSG_ENCRYPTION"`
		VaultURL   string `cli:"--vault-url" env:"SSG_VAULT_URL"`
		VaultToken string `cli:"--vault-token" env:"SSG_VAULT_TOKEN"`

		Mode     string `cli:"-m, --mode" env:"SSG_MODE"`
		FileRoot string `cli:"--file-root" env:"SSG_FILE_ROOT"`
		S3Bucket string `cli:"--s3-bucket" env:"SSG_S3_BUCKET"`
		S3Region string `cli:"--s3-region" env:"SSG_S3_REGION"`
		S3AKI    string `cli:"--s3-aki" env:"SSG_S3_AKI"`
		S3Key    string `cli:"--s3-key" env:"SSG_S3_KEY"`

		Cleanup int `cli:"-c, --cleanup" env:"SSG_CLEANUP"`
		Lease   int `cli:"-L, --lease" env:"SSG_LEASE"`

		AdminUsername string `cli:"--admin-username" env:"SSG_ADMIN_USERNAME"`
		AdminPassword string `cli:"--admin-password" env:"SSG_ADMIN_PASSWORD"`

		ControlUsername string `cli:"--control-username" env:"SSG_CONTROL_USERNAME"`
		ControlPassword string `cli:"--control-password" env:"SSG_CONTROL_PASSWORD"`
	}

	opts.Log = "info"
	opts.Listen = ":3100"
	opts.Mode = "fs"
	opts.S3Region = "us-east-1"
	opts.Cleanup = 5
	opts.Lease = 600
	opts.AdminUsername = "admin"
	opts.AdminPassword = "password"
	opts.ControlUsername = "control"
	opts.ControlPassword = "shield"
	opts.Compression = "zlib"
	opts.Encryption = "aes256-ctr"
	opts.VaultURL = "http://127.0.0.1:8200"

	env.Override(&opts)
	if opts.Encryption == "none" {
		opts.Encryption = ""
	}
	if opts.Compression == "none" {
		opts.Compression = ""
	}

	_, args, err := cli.Parse(&opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!!! %s\n", err)
		os.Exit(1)
	}
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "!!! extra arguments found\n")
		os.Exit(1)
	}

	if opts.Help {
		fmt.Printf("ssg - The SHIELD Storage Gateway\n\n")
		fmt.Printf("Options\n")
		fmt.Printf("  -h, --help          Show this help screen.\n")
		fmt.Printf("  -v, --version       Display the SSG version.\n")
		fmt.Printf("\n")
		fmt.Printf("  -l, --listen        The IP:port on which to bind and listen.\n")
		fmt.Printf("                      Can be set via the @W{$SSG_LISTEN} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  -L, --lease         How long should tokens for accessing streams\n")
		fmt.Printf("                      be valid (in seconds).  Every use of a token\n")
		fmt.Printf("                      renews it by this lease amount.\n")
		fmt.Printf("                      Defaults to 600 seconds (10 min).\n")
		fmt.Printf("                      Can be set via the @W{$SSG_LEASE} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  --admin-username    The credentials for the ADMIN account, which\n")
		fmt.Printf("  --admin-password    is able to query for SSG internals.  Can be set via\n")
		fmt.Printf("                      the @W{SSG_ADMIN_USERNAME} and @W{SSG_ADMIN_PASSWORD}\n")
		fmt.Printf("                      environment variables.\n")
		fmt.Printf("                      Defaults to admin/password.\n")
		fmt.Printf("\n")
		fmt.Printf("  --control-username  The credentials for the CONTROL account, which\n")
		fmt.Printf("  --control-password  is able to start new streams.  Can be set via\n")
		fmt.Printf("                      the @W{SSG_CONTROL_USERNAME} and @W{SSG_CONTROL_PASSWORD}\n")
		fmt.Printf("                      environment variables.\n")
		fmt.Printf("                      Defaults to control/shield.\n")
		fmt.Printf("\n")
		fmt.Printf("  -c, --cleanup       How often (in seconds) to clean up expired.\n")
		fmt.Printf("                      streams and delete half-uploaded files.\n")
		fmt.Printf("                      Defaults to 5 seconds.\n")
		fmt.Printf("                      Can be set via the @W{$SSG_CLEANUP} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  -m, --mode          What mode to operate in for backend storage.\n")
		fmt.Printf("                      Must be one of 'fs' or 's3'\n")
		fmt.Printf("                      Can be set via the @W{$SSG_MODE} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  --file-root         Where to store / find files in --mode 'fs'.\n")
		fmt.Printf("                      Can be set via the @W{$SSG_FILE_ROOT} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  --encryption        Which encryption algorithm to use.  One of:\n")
		fmt.Printf("                      none, aes256-ctr, aes256-cfb, or aes256-ofb.\n")
		fmt.Printf("                      Can be set via the @W{$SSG_ENCRYPTION} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  --vault-url         A Vault endpoint, where encryption parameters will\n")
		fmt.Printf("                      be securely stored.\n")
		fmt.Printf("                      Only honored when --encryption is not 'none'\n")
		fmt.Printf("                      Can be set via the @W{$SSG_VAULT_URL} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  --vault-token       A static token for accessing the Vault.\n")
		fmt.Printf("                      Only honored when --encryption is not 'none'\n")
		fmt.Printf("                      Can be set via the @W{$SSG_VAULT_TOKEN} env var.\n")
		fmt.Printf("\n")
		os.Exit(0)
	}

	if opts.Version {
		if Version == "" || Version == "dev" {
			fmt.Printf("ssg (development)\n")
		} else {
			fmt.Printf("ssg v%s\n", Version)
		}
		os.Exit(0)
	}

	if opts.Debug {
		opts.Log = "debug"
	}
	log.SetupLogging(log.LogConfig{
		Type:  "console",
		Level: opts.Log,
	})

	if opts.Cleanup < 1 {
		fmt.Fprintf(os.Stderr, "@R{!! invalid (non-positive) value for --cleanup: %d}\n", opts.Cleanup)
		os.Exit(1)
	}

	if opts.Lease < 1 {
		fmt.Fprintf(os.Stderr, "@R{!! invalid (non-positive) value for --lease: %d}\n", opts.Lease)
		os.Exit(1)
	}

	ssg := api.New()
	if opts.Mode == "fs" {
		if opts.FileRoot == "" {
			fmt.Fprintf(os.Stderr, "@R{!! no --file-root specified for --mode fs}\n")
			os.Exit(1)
		}
		ssg.UseFiles(opts.FileRoot)

	} else if opts.Mode == "s3" {
		if opts.S3AKI == "" {
			fmt.Fprintf(os.Stderr, "@R{!! no --s3-aki specified for --mode s3}\n")
			os.Exit(1)
		}
		if opts.S3Key == "" {
			fmt.Fprintf(os.Stderr, "@R{!! no --s3-key specified for --mode s3}\n")
			os.Exit(1)
		}
		if opts.S3Bucket == "" {
			fmt.Fprintf(os.Stderr, "@R{!! no --s3-bucket specified for --mode s3}\n")
			os.Exit(1)
		}
		ssg.UseS3(s3.Client{
			Bucket:          opts.S3Bucket,
			AccessKeyID:     opts.S3AKI,
			SecretAccessKey: opts.S3Key,
			Region:          opts.S3Region,
		})

	} else {
		fmt.Fprintf(os.Stderr, "@R{!! unrecognized --mode '%s'}\n", opts.Mode)
		os.Exit(1)
	}

	if opts.Encryption != "" {
		c, err := vault.Connect(opts.VaultURL, opts.VaultToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect to vault at %s\n", opts.VaultURL)
			os.Exit(1)
		}

		ssg.SetStreamConfig(api.StreamConfig{
			Compression: opts.Compression,
			Encryption:  opts.Encryption,
			VaultClient: c,
		})

	} else {
		ssg.SetStreamConfig(api.StreamConfig{
			Compression: opts.Compression,
			Encryption:  opts.Encryption,
		})
	}

	ssg.Lease = time.Duration(opts.Lease) * time.Second
	ssg.Admin.Username = opts.AdminUsername
	ssg.Admin.Password = opts.AdminPassword
	ssg.Control.Username = opts.ControlUsername
	ssg.Control.Password = opts.ControlPassword
	http.Handle("/", ssg.Router())

	fmt.Fprintf(os.Stderr, "ssg starting up...\n")
	if opts.Compression != "" {
		fmt.Fprintf(os.Stderr, "ssg using [%s] compression\n", opts.Compression)
	}
	if opts.Encryption != "" {
		fmt.Fprintf(os.Stderr, "ssg using [%s] encryption\n", opts.Encryption)
		fmt.Fprintf(os.Stderr, "ssg using vault at %s\n", opts.VaultURL)
	}
	fmt.Fprintf(os.Stderr, " - running cleanup routine @C{every %d seconds}\n", opts.Cleanup)
	go ssg.Sweeper(time.Duration(opts.Cleanup) * time.Second)

	fmt.Fprintf(os.Stderr, " - listening on @C{%s} (TCP)\n", opts.Listen)
	fmt.Fprintf(os.Stderr, " - leasing tokens on streams for up to @C{%d seconds}\n", opts.Lease)
	if opts.Mode == "fs" {
		fmt.Fprintf(os.Stderr, " - backed by @W{filesystem} at @C{%s}\n", opts.FileRoot)
	} else {
		fmt.Fprintf(os.Stderr, " - backed by S3 bucket @C{%s}\n", opts.S3Bucket)
		fmt.Fprintf(os.Stderr, "   in region @C{%s}\n", opts.S3Region)
		fmt.Fprintf(os.Stderr, "   accessed via @C{%s}\n", sanitize(opts.S3AKI))
	}
	if err := http.ListenAndServe(opts.Listen, nil); err != nil {
		fmt.Fprintf(os.Stderr, "@R{!! bind %s failed: %s}\n", opts.Listen, err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "ssg shutting down...\n")
	os.Exit(0)
}
