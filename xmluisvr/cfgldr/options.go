package cfgldr

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Options struct {
	Timeout               int
	HTTPPort              int
	DBExtensionFiles      []string
	APIFile               string
	ConnectString         string
	DBPort                int
	DBBootstrapFile       string
	Quiet                 bool
	Verbosity             int
	AllowUntrustedQueries bool
}

var options *Options
var ErrVerbosityMustBe1To3 = errors.New("verbosity must be between 1 to 3")

func GetOptions() (opts *Options, err error) {

	if options != nil {
		goto end
	}
	{ // Block scope used here to get around the infernal limitation in Go
		// that you can't declare a variable after a goto, even when that
		// variable's lifetime is limited to the scope of the function.
		// See: https://github.com/golang/go/issues/26058
		flags := struct {
			port                  *int
			apiFile               *string
			connStr               *string
			dbPort                *int
			dbBootstrapFile       *string
			dbExtensions          stringSliceFlag
			quiet                 *bool
			verbosity             *int
			allowUntrustedQueries *bool
		}{
			port:                  new(int),
			apiFile:               new(string),
			connStr:               new(string),
			dbPort:                new(int),
			dbBootstrapFile:       new(string),
			dbExtensions:          stringSliceFlag{},
			quiet:                 new(bool),
			verbosity:             new(int),
			allowUntrustedQueries: new(bool),
		}

		// Set custom flag usage to display double dashes for word options
		flag.Usage = func() {
			fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
			flag.VisitAll(func(f *flag.Flag) {
				prefix := "-"
				// Use double dash for multi-character flags
				if len(f.Name) > 1 {
					prefix = "--"
				}
				fprintf(flag.CommandLine.Output(), "  %s%s: %s\n", prefix, f.Name, f.Usage)
			})
		}

		// Set up command line flags with long and short versions
		flag.IntVar(flags.port, "port", 8080, "dbPort to run the server on")
		flag.IntVar(flags.port, "p", 8080, "dbPort to run the server on (shorthand)")

		flag.StringVar(flags.apiFile, "api", "", "Path to APIConfig description file")
		flag.StringVar(flags.connStr, "db", "data.db", "Path to SQLite connStr file or PostgreSQL connection string or DB description file")
		flag.StringVar(flags.dbBootstrapFile, "db-bootstrap", DefaultDBBootstrapFilepath,
			fmt.Sprintf("Path to database query file containing idempotent queries to run on start of server (default %s)", DefaultDBBootstrapFilepath),
		)
		flag.IntVar(flags.dbPort, "db-port", 0, "PostgreSQL port (optional, overrides port in --db if provided)")
		flag.Var(&flags.dbExtensions, "db-ext", "One or more paths to connStr extensions to load (currently only SQLite3.)")

		flag.BoolVar(flags.quiet, "quiet", false, "Disable display of most command line output")
		flag.BoolVar(flags.quiet, "q", false, "Disable display of most command line output (shorthand)")
		flag.BoolVar(flags.allowUntrustedQueries, "dangerously-allow-untrusted-db-queries", false, "Allow UNTRUSTED Database Queries to be submitted via the API")

		flag.IntVar(flags.verbosity, "verbosity", 1, "Verbosity of most command line output (1 to 3, default 1)")
		flag.IntVar(flags.verbosity, "v", 1, "Verbosity of most command line output (shorthand, 1 to 3, default 1)")

		flag.Parse()

		if !(1 <= *flags.verbosity && *flags.verbosity <= 3) {
			err = errors.Join(ErrVerbosityMustBe1To3, fmt.Errorf("verbosity=%d", *flags.verbosity))
			goto end
		}

		options = &Options{
			HTTPPort:         *flags.port,
			APIFile:          *flags.apiFile,
			ConnectString:    *flags.connStr,
			DBPort:           *flags.dbPort,
			DBBootstrapFile:  *flags.dbBootstrapFile,
			Quiet:            *flags.quiet,
			Verbosity:        *flags.verbosity,
			DBExtensionFiles: flags.dbExtensions.values(),
		}
	}
end:
	return options, err
}

type stringSliceFlag []string

func (f *stringSliceFlag) values() (v []string) {
	v = make([]string, 0, len(*f))
	for _, s := range *f {
		v = append(v, strings.TrimSpace(s))
	}
	return v
}

func (f *stringSliceFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *stringSliceFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func fprintf(w io.Writer, format string, a ...any) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
}
