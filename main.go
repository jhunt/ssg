package main

import (
	"encoding/json"
	"io"
	"os"

	fmt "github.com/jhunt/go-ansi"

	"github.com/jhunt/go-cli"
	env "github.com/jhunt/go-envirotron"
	"github.com/jhunt/go-log"

	"github.com/jhunt/shield-storage-gateway/pkg/client"
	"github.com/jhunt/shield-storage-gateway/pkg/ssg"
)

var Version = ""

func main() {
	var opts struct {
		Help    bool `cli:"-h, --help"`
		Version bool `cli:"-v, --version"`

		Server struct {
			Debug bool   `cli:"-D, --debug"     env:"SSG_DEBUG"`
			Log   string `cli:"-L, --log-level" env:"SSG_LOG_LEVEL"`

			Config string `cli:"-c, --config" env:"SSG_CONFIG"`
		} `cli:"server"`

		Control struct {
			URL   string `cli:"-u, --url"   env:"SSG_URL"`
			Token string `cli:"-t, --token" env:"SSG_CONTROL_TOKEN"`

			Upload   struct{} `cli:"upload"`
			Download struct{} `cli:"download"`
			Expunge  struct{} `cli:"expunge, delete, rm"`
		} `cli:"control, c"`

		Stream struct {
			URL   string `cli:"-u, --url"   env:"SSG_URL"`
			Token string `cli:"-t, --token" env:"SSG_STREAM_TOKEN"`

			Get struct{} `cli:"get"`
			Put struct{} `cli:"put"`
		} `cli:"stream, s"`

		Upload struct {
			URL   string `cli:"-u, --url"   env:"SSG_URL"`
			Token string `cli:"-t, --token" env:"SSG_CONTROL_TOKEN"`
		} `cli:"upload, up"`

		Download struct {
			URL   string `cli:"-u, --url"   env:"SSG_URL"`
			Token string `cli:"-t, --token" env:"SSG_CONTROL_TOKEN"`
		} `cli:"download, down"`
	}

	opts.Server.Log = "info"
	opts.Server.Config = "/etc/ssg/ssg.yml"
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
		vers = fmt.Sprintf("ssg v%s")
	}
	if opts.Version {
		fmt.Printf("%s\n", vers)
		os.Exit(0)
	}

	if command == "server" {
		if len(args) != 0 {
			fmt.Fprintf(os.Stderr, "!!! extra arguments found\n")
			os.Exit(1)
		}
		if opts.Server.Debug {
			opts.Server.Log = "debug"
		}
		log.SetupLogging(log.LogConfig{
			Type:  "console",
			Level: opts.Server.Log,
		})

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

	if command == "control upload" {
		if opts.Control.Token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token}\n")
			os.Exit(2)
		}
		if opts.Control.URL == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
			os.Exit(1)
		}
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "!! missing required REMOTE-PATH argument\n")
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
			os.Exit(1)
		}

		c := client.Controller{
			URL:   opts.Control.URL,
			Token: opts.Control.Token,
		}

		stream, err := c.NewUpload(args[0])
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
		if opts.Control.Token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token}\n")
			os.Exit(1)
		}
		if opts.Control.URL == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
			os.Exit(1)
		}
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "!! missing required REMOTE-PATH argument\n")
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
			os.Exit(1)
		}

		c := client.Controller{
			URL:   opts.Control.URL,
			Token: opts.Control.Token,
		}

		stream, err := c.NewDownload(args[0])
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
		if opts.Control.Token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token}\n")
			os.Exit(1)
		}
		if opts.Control.URL == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
			os.Exit(1)
		}
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "!! missing required REMOTE-PATH argument\n")
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
			os.Exit(1)
		}

		c := client.Controller{
			URL:   opts.Control.URL,
			Token: opts.Control.Token,
		}

		err := c.Expunge(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/control} failed: @R{%s}\n", err)
			os.Exit(2)
		}
		os.Exit(0)
	}

	if command == "stream get" {
		if opts.Stream.Token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token}\n")
			os.Exit(1)
		}
		if opts.Stream.URL == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
			os.Exit(1)
		}
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "!! missing required STREAM-ID argument\n")
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
			os.Exit(1)
		}

		c := client.Customer{
			URL: opts.Stream.URL,
		}
		rd, err := c.Download(args[0], opts.Stream.Token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		}
		io.Copy(os.Stdout, rd)
		rd.Close()
		os.Exit(0)
	}

	if command == "stream put" {
		if opts.Stream.Token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token}\n")
			os.Exit(1)
		}
		if opts.Stream.URL == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
			os.Exit(1)
		}
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "!! missing required STREAM-PATH argument\n")
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
			os.Exit(1)
		}

		c := client.Customer{
			URL: opts.Stream.URL,
		}
		n, err := c.Upload(args[0], opts.Stream.Token, os.Stdin, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		}
		fmt.Fprintf(os.Stderr, "uploaded %d bytes\n", n)
		os.Exit(0)
	}

	if command == "upload" {
		if opts.Upload.Token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token}\n")
			os.Exit(2)
		}
		if opts.Upload.URL == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
			os.Exit(1)
		}
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "!! missing required REMOTE-PATH argument\n")
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
			os.Exit(1)
		}

		cc := client.Controller{
			URL:   opts.Upload.URL,
			Token: opts.Upload.Token,
		}

		stream, err := cc.NewUpload(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/control} failed: @R{%s}\n", err)
			os.Exit(2)
		}

		cu := client.Customer{
			URL: opts.Stream.URL,
		}
		fmt.Fprintf(os.Stderr, "uploading to @C{%s}\n", stream.Canon)
		n, err := cu.Upload(stream.ID, stream.Token, os.Stdin, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		}
		fmt.Fprintf(os.Stderr, "uploaded %d bytes\n", n)
		os.Exit(0)
	}

	if command == "download" {
		if opts.Download.Token == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--token}\n")
			os.Exit(1)
		}
		if opts.Download.URL == "" {
			fmt.Fprintf(os.Stderr, "!! missing required @Y{--url}\n")
			os.Exit(1)
		}
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "!! missing required REMOTE-PATH argument\n")
			os.Exit(1)
		}
		if len(args) > 1 {
			fmt.Fprintf(os.Stderr, "!! extra arguments found\n")
			os.Exit(1)
		}

		cc := client.Controller{
			URL:   opts.Download.URL,
			Token: opts.Download.Token,
		}

		stream, err := cc.NewDownload(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @W{/control} failed: @R{%s}\n", err)
			os.Exit(2)
		}

		cu := client.Customer{
			URL: cc.URL,
		}
		rd, err := cu.Download(stream.ID, stream.Token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "!! @R{%s}\n", err)
		}
		io.Copy(os.Stdout, rd)
		rd.Close()
		os.Exit(0)
	}

}
