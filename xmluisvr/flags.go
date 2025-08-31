package xmluisvr

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/xmlui-org/xmluisvr/dbutil"
)

func parseFlagValues() (fv *FlagValues) {

	flags := struct {
		ServerPort *int
		Extension  *string
		ApiDesc    *string
		DbPath     *string
		PgConnStr  *string
		PgPort     *int
		Verbose    *bool
	}{
		ServerPort: new(int),
		Extension:  new(string),
		ApiDesc:    new(string),
		DbPath:     new(string),
		PgConnStr:  new(string),
		PgPort:     new(int),
		Verbose:    new(bool),
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
	flag.IntVar(flags.ServerPort, "port", 8080, "Port to run the server on")
	flag.IntVar(flags.ServerPort, "p", 8080, "Port to run the server on (shorthand)")

	flag.StringVar(flags.Extension, "extension", "", "Path to SQLite extension to load")
	flag.StringVar(flags.ApiDesc, "api", "", "Path to API description file")
	flag.StringVar(flags.DbPath, "db", "data.db", "Path to SQLite database file")
	flag.StringVar(flags.PgConnStr, "pg-conn", "", "PostgreSQL connection string (if provided, use PostgreSQL instead of SQLite)")
	flag.IntVar(flags.PgPort, "pg-port", 0, "PostgreSQL port (optional, overrides port in --pg-conn if provided)")

	flag.BoolVar(flags.Verbose, "verbose", false, "Enable logging of SQL query responses")
	flag.BoolVar(flags.Verbose, "v", false, "Enable logging of SQL query responses (shorthand)")

	flag.Parse()
	return &FlagValues{
		ServerPort: *flags.ServerPort,
		Extension:  *flags.Extension,
		ApiDesc:    *flags.ApiDesc,
		DbPath:     *flags.DbPath,
		PgConnStr:  *flags.PgConnStr,
		PgPort:     *flags.PgPort,
		Verbose:    *flags.Verbose,
	}
}

type FlagValues struct {
	ServerPort int
	Extension  string
	ApiDesc    string
	DbPath     string
	PgConnStr  string
	PgPort     int
	Verbose    bool
}

func (fv FlagValues) DatabaseType() (dt dbutil.DatabaseType) {
	dt = dbutil.SQLiteDatabase
	if fv.PgConnStr != "" {
		dt = dbutil.PostgresDatabase
	}
	return dt
}

func (fv FlagValues) ExtensionPaths() []string {
	if fv.Extension == "" {
		return []string{}
	}
	return []string{fv.Extension}
}

func (fv FlagValues) ConnectionString() (cs string) {
	dt := fv.DatabaseType()
	switch dt {
	case dbutil.SQLiteDatabase:
		cs = fv.DbPath
	case dbutil.PostgresDatabase:
		cs = fv.pgConnectionString()
	default:
		panic(fmt.Sprintf("Invalid database type '%s'", dt))
	}
	return cs
}

var postgresPrefixRE = regexp.MustCompile(`^\s*postgres(ql)?://`)
var dsnFormatRE = regexp.MustCompile(`port=\\d+`)

// pgConnectionString injects or overrides the port in a Postgres connection string (URL or DSN format)
func (fv FlagValues) pgConnectionString() (cs string) {
	if fv.PgPort == 0 {
		cs = fv.PgConnStr
		goto end
	}
	if postgresPrefixRE.MatchString(fv.PgConnStr) {
		u, err := url.Parse(fv.PgConnStr)
		if err != nil {
			goto end
		}
		if u == nil {
			err = fmt.Errorf("invalid PostgreSQL connection string: '%s'", fv.PgConnStr)
			goto end
		}
		if u.Port() == "" {
			goto end
		}
		if u.Port() == strconv.Itoa(fv.PgPort) {
			goto end
		}
		u.Host = fmt.Sprintf("%s:%d", u.Hostname(), fv.PgPort)
		cs = u.String()
		goto end
	}

	// DSN format: add or replace port=...
	if dsnFormatRE.MatchString(fv.PgConnStr) {
		cs = dsnFormatRE.ReplaceAllString(fv.PgConnStr, fmt.Sprintf("port=%d", fv.PgPort))
		goto end
	}
	cs = fmt.Sprintf("%s port=%d", fv.PgConnStr, fv.PgPort)
end:
	return cs
}
