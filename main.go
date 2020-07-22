package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	fmt "github.com/jhunt/go-ansi"

	"github.com/jhunt/go-cli"
	env "github.com/jhunt/go-envirotron"
	"github.com/jhunt/go-log"

	"github.com/jhunt/ssg/pkg/client"
	"github.com/jhunt/ssg/pkg/ssg"
)

var Version = ""

func main() {
	var opts struct {
		Help    bool   `cli:"-h, --help"`
		Version bool   `cli:"-v, --version"`
		Debug   bool   `cli:"-D, --debug"     env:"SSG_DEBUG"`
		Log     string `cli:"-L, --log-level" env:"SSG_LOG_LEVEL"`

		Server struct {
			Config string `cli:"-c, --config" env:"SSG_CONFIG"`
		} `cli:"server"`

		URL   string `cli:"-u, --url"   env:"SSG_URL"`
		Token string `cli:"-t, --token" env:"SSG_CONTROL_TOKEN"`

		Ping struct {
			Quiet bool `cli:"-q, --quiet"`
		} `cli:"ping"`

		Control struct {
			Buckets  struct{} `cli:"buckets, ls"`
			Upload   struct{} `cli:"upload"`
			Download struct{} `cli:"download"`
			Expunge  struct{} `cli:"expunge, delete, rm"`
		} `cli:"control, c"`

		Stream struct {
			Get struct{} `cli:"get"`
			Put struct{} `cli:"put"`
		} `cli:"stream, s"`

		Upload   struct{
			SegmentSize int `cli:"-s, --segment-size"`
		} `cli:"upload, up"`
		Download struct{} `cli:"download, down"`
	}

	opts.Log = "info"
	opts.Server.Config = "/etc/ssg/ssg.yml"
	opts.Upload.SegmentSize = 1024 * 1024
	env.Override(&opts)
	command, args, err := cli.Parse(&opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "!!! %s\n", err)
		os.Exit(1)
	}

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
		vers = fmt.Sprintf("ssg v%s", Version)
	}
	if opts.Version {
		fmt.Printf("%s\n", vers)
		os.Exit(0)
	}

	if opts.Debug {
		opts.Log = "debug"
	}
	log.SetupLogging(log.LogConfig{
		Type:  "console",
		Level: opts.Log,
	})

	if command == "server" {
		if opts.URL != "" {
			fmt.Fprintf(os.Stderr, "!! @W{warning:} the --url flag is ignored in server mode.\n")
		}
		if opts.Token != "" {
			fmt.Fprintf(os.Stderr, "!! @W{warning:} the --token flag is ignored in server mode.\n")
		}

		if len(args) != 0 {
			fmt.Fprintf(os.Stderr, "!!! extra arguments found\n")
			os.Exit(1)
		}

		s, err := ssg.NewServerFromFile(opts.Server.Config)
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

	if command == "ping" {
		if opts.URL == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
			os.Exit(1)
		}
		c := client.Client{URL: opts.URL}
		helo, err := c.Ping()
		if err != nil {
			if !opts.Ping.Quiet {
				fmt.Fprintf(os.Stderr, "!! @W{/ (ping)} failed: @R{%s}\n", err)
			}
			os.Exit(1)
		}
		if !opts.Ping.Quiet {
			fmt.Printf("%s\n", helo)
		}
		os.Exit(0)
	}

	if command == "control buckets" {
		c := controller(opts.URL, opts.Token, "SSG_CONTROL_TOKEN")
		if len(args) > 0 {
			fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
			os.Exit(1)
		}

		buckets, err := c.Buckets()
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/buckets} failed: @R{%s}\n", err)
			os.Exit(2)
		}

		b, err := json.MarshalIndent(buckets, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! failed to json: @R{%s}\n", err)
			os.Exit(3)
		}
		fmt.Printf("%s\n", string(b))
		os.Exit(0)
	}

	if command == "control upload" {
		c := controller(opts.URL, opts.Token, "SSG_CONTROL_TOKEN")
		target := needTarget(args, "REMOTE-PATH")

		stream, err := c.NewUpload(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/control} failed: @R{%s}\n", err)
			os.Exit(2)
		}

		b, err := json.MarshalIndent(stream, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! failed to json: @R{%s}\n", err)
			os.Exit(3)
		}
		fmt.Printf("%s\n", string(b))
		os.Exit(0)
	}

	if command == "control download" {
		c := controller(opts.URL, opts.Token, "SSG_CONTROL_TOKEN")
		target := needTarget(args, "REMOTE-PATH")

		stream, err := c.NewDownload(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/control} failed: @R{%s}\n", err)
			os.Exit(2)
		}

		b, err := json.MarshalIndent(stream, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! failed to json: @R{%s}\n", err)
			os.Exit(3)
		}
		fmt.Printf("%s\n", string(b))
		os.Exit(0)
	}

	if command == "control expunge" {
		c := controller(opts.URL, opts.Token, "SSG_CONTROL_TOKEN")
		target := needTarget(args, "REMOTE-PATH")

		err := c.Expunge(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/control} failed: @R{%s}\n", err)
			os.Exit(2)
		}
		os.Exit(0)
	}

	if command == "stream get" {
		c, token := streamer(opts.URL, opts.Token, "SSG_STREAM_TOKEN")
		target := needTarget(args, "REMOTE-ID")

		rd, err := c.Get(target, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		}
		io.Copy(os.Stdout, rd)
		rd.Close()
		os.Exit(0)
	}

	if command == "stream put" {
		c, token := streamer(opts.URL, opts.Token, "SSG_STREAM_TOKEN")
		target := needTarget(args, "REMOTE-ID")

		n, err := c.Put(target, token, os.Stdin, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "uploaded %d bytes\n", n)
		}
		os.Exit(0)
	}

	if command == "upload" {
		c := controller(opts.URL, opts.Token, "SSG_CONTROL_TOKEN")
		c.SegmentSize = opts.Upload.SegmentSize
		target := needTarget(args, "REMOTE-PATH")

		stream, err := c.NewUpload(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/control} failed: @R{%s}\n", err)
			os.Exit(2)
		}

		fmt.Fprintf(os.Stderr, "uploading to @C{%s}\n", stream.Canon)
		n, err := c.Put(stream.ID, stream.Token, os.Stdin, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "uploaded %d bytes\n", n)
		}
		os.Exit(0)
	}

	if command == "download" {
		c := controller(opts.URL, opts.Token, "SSG_CONTROL_TOKEN")
		target := needTarget(args, "REMOTE-PATH")

		stream, err := c.NewDownload(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/control} failed: @R{%s}\n", err)
			os.Exit(2)
		}

		rd, err := c.Get(stream.ID, stream.Token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		}
		io.Copy(os.Stdout, rd)
		rd.Close()
		os.Exit(0)
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "!! no command supplied.\n")
	} else {
		if command == "" {
			fmt.Fprintf(os.Stderr, "!! unrecognized command @Y{'%s'}\n", strings.Join(args, " "))
		} else {
			fmt.Fprintf(os.Stderr, "!! unrecognized @C{%s} sub-command @Y{'%s'}\n", command, strings.Join(args, " "))
		}
	}
	os.Exit(3)
}

func controller(url, token, env string) client.Client {
	if url == "" {
		fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
		os.Exit(1)
	}
	if token == "" {
		token = os.Getenv(env)
		if token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token} / @Y{$%s}\n", env)
			os.Exit(2)
		}
	}

	return client.Client{
		URL:          url,
		ControlToken: token,
	}
}

func streamer(url, token, env string) (client.Client, string) {
	if url == "" {
		fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
		os.Exit(1)
	}
	if token == "" {
		token = os.Getenv(env)
		if token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token} / @Y{$%s}\n", env)
			os.Exit(2)
		}
	}

	return client.Client{
		URL: url,
	}, token
}

func needTarget(args []string, name string) string {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "!! missing required @Y{%s} argument\n", name)
		os.Exit(1)
	}
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
		os.Exit(1)
	}
	return args[0]
}
