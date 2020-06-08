package main

import (
	"os"

	fmt "github.com/jhunt/go-ansi"

	"github.com/jhunt/go-cli"
	env "github.com/jhunt/go-envirotron"
	"github.com/jhunt/go-log"

	"github.com/jhunt/shield-storage-gateway/pkg/ssg"
)

var Version = ""

func main() {
	var opts struct {
		Help    bool `cli:"-h, --help"`
		Version bool `cli:"-v, --version"`

		Debug bool   `cli:"-D, --debug"     env:"SSG_DEBUG"`
		Log   string `cli:"-L, --log-level" env:"SSG_LOG_LEVEL"`

		Config string `cli:"-c, --config" env:"SSG_CONFIG"`
	}

	opts.Log = "info"
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

	if opts.Debug {
		opts.Log = "debug"
	}
	log.SetupLogging(log.LogConfig{
		Type:  "console",
		Level: opts.Log,
	})

	if opts.Help {
		fmt.Printf("ssg - The SHIELD Storage Gateway\n\n")
		fmt.Printf("Options\n")
		fmt.Printf("  -h, --help          Show this help screen.\n")
		fmt.Printf("  -v, --version       Display the SSG version.\n")
		fmt.Printf("\n")
		fmt.Printf("  -D, --debug         Enable verbose debugging.  Overrides -L\n")
		fmt.Printf("                      Can be set via the @W{$SSG_DEBUG} env var.\n")
		fmt.Printf("\n")
		fmt.Printf("  -L, --log-level     Set the log level; one of: error, warning,\n")
		fmt.Printf("                      info (the default) or debug.\n")
		fmt.Printf("                      Can be set via the @W{$SSG_LOG_LEVEL} env var.\n")
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

	s, err := ssg.NewServerFromFile(opts.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		os.Exit(1)
	}
	if err := s.Run(vers); err != nil {
		fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		os.Exit(2)
	}

	os.Exit(0)
}
