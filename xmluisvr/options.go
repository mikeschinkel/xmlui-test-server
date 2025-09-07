package xmluisvr

import (
	"flag"
	"os"
	"strings"
)

type Options struct {
	Timeout      int
	ServerPort   int
	DBExtensions []string
	API          string
	Database     string
	DatabasePort int
	Verbose      bool
}

func parseOptions() (opts *Options) {

	flags := struct {
		port         *int
		api          *string
		database     *string
		dbPort       *int
		dbExtensions stringSliceFlag
		verbose      *bool
	}{
		port:         new(int),
		api:          new(string),
		database:     new(string),
		dbPort:       new(int),
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

	flag.StringVar(flags.api, "api", "", "Path to API description file")
	flag.StringVar(flags.database, "db", "data.db", "Path to SQLite database file or PostgreSQL connection string or DB description file")
	flag.IntVar(flags.dbPort, "db-port", 0, "PostgreSQL port (optional, overrides port in --db if provided)")
	flag.Var(&flags.dbExtensions, "db-ext", "One or more paths to database extensions to load (currently only SQLite3.)")

	flag.BoolVar(flags.verbose, "verbose", false, "Enable logging of SQL query responses")
	flag.BoolVar(flags.verbose, "v", false, "Enable logging of SQL query responses (shorthand)")

	flag.Parse()
	return &Options{
		ServerPort:   *flags.port,
		API:          *flags.api,
		Database:     *flags.database,
		DatabasePort: *flags.dbPort,
		Verbose:      *flags.verbose,
		DBExtensions: flags.dbExtensions.values(),
	}
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
