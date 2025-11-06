package cfgldr

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/xmlui-org/localdev/xmluisvr/common"
	"github.com/xmlui-org/localdev/xmluisvr/dbqvars"
)

const (
	DefaultTimeout               = 3
	DefaultHTTPPort              = 8080
	DefaultAPIFile               = ""
	DefaultDBPort                = 0
	DefaultDBBootstrapFile       = "bootstrap.sql"
	DefaultQuiet                 = false
	DefaultAllowUntrustedQueries = false
	DefaultVerbosity             = cliutil.DefaultVerbosity
	DefaultDBAccessMode          = int(dbqvars.DBReadWriteMode)
)

const (
	DefaultErrorStyle = string(common.DefaultErrorStyle)
)

const (
	AllowUntrustedQueriesFlag = "dangerously-allow-untrusted-db-queries"
)

var (
	DefaultConnectString = DefaultSQLite3Database
)

type Options struct {
	Timeout               int
	HTTPPort              int
	APIFile               string
	ConnectString         string
	DBPort                int
	DBBootstrapFile       string
	Quiet                 bool
	Verbosity             int
	ErrorStyle            string
	AllowUntrustedQueries bool
	DBAccessMode          int
	DBExtensionFiles      []string
}

func (*Options) Options() {}

type OptionsArgs struct {
	Timeout               *int
	HTTPPort              *int
	APIFile               *string
	ConnectString         *string
	DBPort                *int
	DBBootstrapFile       *string
	Quiet                 *bool
	Verbosity             *int
	ErrorStyle            *string
	DBAccessMode          *int
	AllowUntrustedQueries *bool
	DBExtensionFiles      []string
}

func NewOptions(args OptionsArgs) *Options {
	opts := &Options{}

	if args.Timeout != nil {
		opts.Timeout = *args.Timeout
	}
	if args.HTTPPort != nil {
		opts.HTTPPort = *args.HTTPPort
	}
	if args.APIFile != nil {
		opts.APIFile = *args.APIFile
	}
	if args.ConnectString != nil {
		opts.ConnectString = *args.ConnectString
	}
	if args.DBPort != nil {
		opts.DBPort = *args.DBPort
	}
	if args.DBBootstrapFile != nil {
		opts.DBBootstrapFile = *args.DBBootstrapFile
	}
	if args.Quiet != nil {
		opts.Quiet = *args.Quiet
	}
	if args.Verbosity != nil {
		opts.Verbosity = *args.Verbosity
	}
	if args.ErrorStyle != nil {
		opts.ErrorStyle = *args.ErrorStyle
	}
	if args.DBAccessMode != nil {
		opts.DBAccessMode = *args.DBAccessMode
	}
	if args.DBExtensionFiles != nil {
		opts.DBExtensionFiles = args.DBExtensionFiles
	}
	return opts
}

var options *Options

func GetOptions() (opts *Options, err error) {
	var verbosity cliutil.Verbosity

	if options != nil {
		opts = options
		goto end
	}
	{ // Block scope used here to get around the infernal limitation in Go
		// that you can't declare a variable after a goto, even when that
		// variable's lifetime is limited to the scope of the function.
		// See: https://github.com/golang/go/issues/26058
		flags := struct {
			timeout               *int
			port                  *int
			apiFile               *string
			connStr               *string
			dbPort                *int
			dbBootstrapFile       *string
			dbExtensions          stringSliceFlag
			quiet                 *bool
			verbosity             *int
			errorStyle            *string
			dbAccessMode          *int
			allowUntrustedQueries *bool
		}{
			timeout:               new(int),
			port:                  new(int),
			apiFile:               new(string),
			connStr:               new(string),
			dbPort:                new(int),
			dbBootstrapFile:       new(string),
			dbExtensions:          stringSliceFlag{},
			quiet:                 new(bool),
			verbosity:             new(int),
			errorStyle:            new(string),
			dbAccessMode:          new(int),
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
		flag.IntVar(flags.port, "port", DefaultHTTPPort, "dbPort to run the server on")
		flag.IntVar(flags.port, "p", DefaultHTTPPort, "dbPort to run the server on (shorthand)")

		flag.IntVar(flags.timeout, "timeout", DefaultTimeout, "Timeout(in seconds) (TODO explain what this controls)")

		flag.StringVar(flags.apiFile, "api", DefaultAPIFile, "Path to APIConfig description file")
		flag.StringVar(flags.connStr, "db", DefaultConnectString, "Path to SQLite connStr file or PostgreSQL connection string or DB description file")
		flag.StringVar(flags.dbBootstrapFile, "db-bootstrap", DefaultDBBootstrapFilepath,
			fmt.Sprintf("Path to database query file containing idempotent queries to run on start of server (default %s)", DefaultDBBootstrapFilepath),
		)
		flag.IntVar(flags.dbPort, "db-port", 0, "PostgreSQL port (optional, overrides port in --db if provided)")
		flag.Var(&flags.dbExtensions, "db-ext", "One or more paths to connStr extensions to load (currently only SQLite3.)")
		flag.IntVar(flags.dbAccessMode, "db-access", DefaultDBAccessMode, "Mode for API access the database (1=Read-only,2=Read-Write,3=Admin,4=SuperAdmin, default 2)")

		flag.BoolVar(flags.quiet, "quiet", DefaultQuiet, "Disable display of most command line output")
		flag.BoolVar(flags.quiet, "q", DefaultQuiet, "Disable display of most command line output (shorthand)")
		flag.BoolVar(flags.allowUntrustedQueries, AllowUntrustedQueriesFlag, false, "Allow UNTRUSTED Database Queries to be submitted via the API")

		flag.IntVar(flags.verbosity, "verbosity", DefaultVerbosity, "Verbosity of most command line output (1 to 3, default 1)")
		flag.IntVar(flags.verbosity, "v", DefaultVerbosity, "Verbosity of most command line output (shorthand, 1 to 3, default 1)")

		flag.StringVar(flags.errorStyle, "err-style", DefaultErrorStyle, "Errors style can be 'dev' for Developer style, or 'pres' for Presentation style")

		flag.Parse()

		// Parse Verbosity because it is the only one that gets used immediately that
		// needs to be parsed.
		verbosity, err = cliutil.ParseVerbosity(*flags.verbosity)

		opts = NewOptions(OptionsArgs{
			HTTPPort:              flags.port,
			APIFile:               flags.apiFile,
			ConnectString:         flags.connStr,
			DBPort:                flags.dbPort,
			DBBootstrapFile:       flags.dbBootstrapFile,
			Quiet:                 flags.quiet,
			Verbosity:             intPtr(int(verbosity)),
			DBExtensionFiles:      flags.dbExtensions.values(),
			Timeout:               flags.timeout,
			ErrorStyle:            flags.errorStyle,
			AllowUntrustedQueries: flags.allowUntrustedQueries,
		})
	}
	options = opts
end:
	return opts, err
}
func intPtr(n int) *int {
	return &n
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
