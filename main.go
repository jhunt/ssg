package main

import (
	"net/http"
	"os"

	fmt "github.com/jhunt/go-ansi"

	"github.com/jhunt/go-cli"
	env "github.com/jhunt/go-envirotron"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg/config"
)

var Version = ""

func main() {
	var opts struct {
		Help    bool `cli:"-h, --help"`
		Version bool `cli:"-v, --version"`

		Config string `cli:"-c, --config" env:"SSG_CONFIG"`
	}

	opts.Config = "/etc/ssg/ssg.yml"
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
		fmt.Printf("  -h, --help          Show this help screen.\n")
		fmt.Printf("  -v, --version       Display the SSG version.\n")
		fmt.Printf("\n")
		fmt.Printf("  -c, --config        Path to the SSG configuration file (YAML!)\n")
		fmt.Printf("                      Can be set via the @W{$SSG_CONFIG} env var.\n")
		fmt.Printf("\n")
		os.Exit(0)
	}

	var vers string
	if Version == "" || Version == "dev" {
		vers = "ssg (development)"
	} else {
		vers = fmt.Sprintf("ssg v%s")
	}
	if opts.Version {
		fmt.Printf("%s\n", vers)
		os.Exit(0)
	}

	cfg, err := config.ReadFile(opts.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		os.Exit(1)
	}

	s, err := ssg.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		os.Exit(1)
	}

	http.Handle("/", s.Router(vers))
	go s.Sweep()

	fmt.Fprintf(os.Stderr, "ssg starting up on %s...\n", s.Bind)
	if err := http.ListenAndServe(s.Bind, nil); err != nil {
		fmt.Fprintf(os.Stderr, "@R{!! bind %s failed: %s}\n", s.Bind, err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "ssg shutting down...\n")
	os.Exit(0)
}
