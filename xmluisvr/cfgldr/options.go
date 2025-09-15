package cfgldr

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Options struct {
	Timeout          int
	HTTPPort         int
	DBExtensionFiles []string
	APIFile          string
	ConnectString    string
	DBPort           int
	DBSchemaFile     string
	Verbose          bool
}

var options *Options

func GetOptions() (opts *Options) {

	if options != nil {
		goto end
	}
	{ // Block scope used here to get around the infernal limitation in Go
		// that you can't declare a variable after a goto, even when that
		// variable's lifetime is limited to the scope of the function.
		// See: https://github.com/golang/go/issues/26058
		flags := struct {
			port         *int
			apiFile      *string
			connStr      *string
			dbPort       *int
			dbSchemaFile *string
			dbExtensions stringSliceFlag
			verbose      *bool
		}{
			port:         new(int),
			apiFile:      new(string),
			connStr:      new(string),
			dbPort:       new(int),
			dbSchemaFile: new(string),
			dbExtensions: stringSliceFlag{},
			verbose:      new(bool),
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
		flag.StringVar(flags.dbSchemaFile, "db-schema", "schema.sql", "Path to idempotent SQL file containing SQL to create your desired connStr schema")
		flag.IntVar(flags.dbPort, "db-port", 0, "PostgreSQL port (optional, overrides port in --db if provided)")
		flag.Var(&flags.dbExtensions, "db-ext", "One or more paths to connStr extensions to load (currently only SQLite3.)")

		flag.BoolVar(flags.verbose, "verbose", false, "Enable logging of SQL query responses")
		flag.BoolVar(flags.verbose, "v", false, "Enable logging of SQL query responses (shorthand)")

		flag.Parse()
		options = &Options{
			HTTPPort:         *flags.port,
			APIFile:          *flags.apiFile,
			ConnectString:    *flags.connStr,
			DBPort:           *flags.dbPort,
			DBSchemaFile:     *flags.dbSchemaFile,
			Verbose:          *flags.verbose,
			DBExtensionFiles: flags.dbExtensions.values(),
		}
	}
end:
	return options
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
