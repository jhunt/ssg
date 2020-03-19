package main

import (
	fmt "github.com/jhunt/go-ansi"
	"net/http"
	"os"
	"time"

	"github.com/jhunt/go-cli"
	env "github.com/jhunt/go-envirotron"

	"github.com/shieldproject/shield-storage-gateway/api"
)

var Version = ""

func main() {
	var opts struct {
		Help    bool `cli:"-h, --help"`
		Debug   bool `cli:"-D, --debug"`
		Version bool `cli:"-v, --version"`

		Listen string `cli:"-l, --listen" env:"SSG_LISTEN"`

		Mode     string `cli:"-m, --mode" env:"SSG_MODE"`
		FileRoot string `cli:"--file-root" env:"SSG_FILE_ROOT"`

		Cleanup int `cli:"-c, --cleanup" env:"SSG_CLEANUP"`
		Lease   int `cli:"-L, --lease" env:"SSG_LEASE"`

		// eventually: S3 creds
	}

	opts.Listen = ":3100"
	opts.Mode = "fs"
	opts.Cleanup = 5
	opts.Lease = 600

	env.Override(&opts)

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
		fmt.Printf("  -h, --help       Show this help screen.\n")
		fmt.Printf("  -v, --version    Display the SSG version.\n")
		fmt.Printf("\n")
		fmt.Printf("  -l, --listen     The IP:port on which to bind and listen.\n")
		fmt.Printf("                   Can be set via the @W{$SSG_LISTEN} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  -L, --lease      How long should tokens for accessing streams\n")
		fmt.Printf("                   be valid (in seconds).  Every use of a token\n")
		fmt.Printf("                   renews it by this lease amount.\n")
		fmt.Printf("                   Defaults to 600 seconds (10 min).\n")
		fmt.Printf("                   Can be set via the @W{$SSG_LEASE} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  -c, --cleanup    How often (in seconds) to clean up expired.\n")
		fmt.Printf("                   streams and delete half-uploaded files.\n")
		fmt.Printf("                   Defaults to 5 seconds.\n")
		fmt.Printf("                   Can be set via the @W{$SSG_CLEANUP} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  -m, --mode       What mode to operate in for backend storage.\n")
		fmt.Printf("                   Must be one of 'fs' or 's3'\n")
		fmt.Printf("                   Can be set via the @W{$SSG_MODE} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  --file-root      Where to store / find files in --mode 'fs'.\n")
		fmt.Printf("                   Can be set via the @W{$SSG_FILE_ROOT} env var.\n")
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

	if opts.Cleanup < 1 {
		fmt.Fprintf(os.Stderr, "@R{!! invalid (non-positive) value for --cleanup: %d}\n", opts.Cleanup)
		os.Exit(1)
	}

	if opts.Lease < 1 {
		fmt.Fprintf(os.Stderr, "@R{!! invalid (non-positive) value for --lease: %d}\n", opts.Lease)
		os.Exit(1)
	}

	var ssg api.API
	if opts.Mode == "fs" {
		if opts.FileRoot == "" {
			fmt.Fprintf(os.Stderr, "@R{!! no --file-root specified for --mode fs}\n")
			os.Exit(1)
		}
		ssg = api.New(opts.FileRoot)
	} else if opts.Mode == "s3" {
		fmt.Fprintf(os.Stderr, "@Y{not yet finished...}\n")
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stderr, "@R{!! unrecognized --mode '%s'}\n", opts.Mode)
		os.Exit(1)
	}

	ssg.Debug = opts.Debug
	ssg.Lease = time.Duration(opts.Lease) * time.Second
	http.Handle("/", ssg.Router())

	fmt.Fprintf(os.Stderr, "ssg starting up...\n")
	fmt.Fprintf(os.Stderr, " - running cleanup routine @C{every %d seconds}\n", opts.Cleanup)
	go ssg.Sweeper(time.Duration(opts.Cleanup) * time.Second)

	fmt.Fprintf(os.Stderr, " - listening on @C{%s} (TCP)\n", opts.Listen)
	fmt.Fprintf(os.Stderr, " - leasing tokens on streams for up to @C{%d seconds}\n", opts.Lease)
	if opts.Mode == "fs" {
		fmt.Fprintf(os.Stderr, " - backed by @W{filesystem} at @C{%s}\n", opts.FileRoot)
	} else {
		fmt.Fprintf(os.Stderr, " - backed by S3 at ...\n") // FIXME
	}
	if err := http.ListenAndServe(opts.Listen, nil); err != nil {
		fmt.Fprintf(os.Stderr, "@R{!! bind %s failed: %s}\n", opts.Listen, err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "ssg shutting down...\n")
	os.Exit(0)
}
