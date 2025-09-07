package postgrespkg

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/xmlui-org/xmluisvr/cliutil"
	"github.com/xmlui-org/xmluisvr/dbpkg"
)

const PostgresDatabase dbpkg.DatabaseType = "postgres"

func init() {
	dbpkg.RegisterDatabase(&Postgres{})
}

var _ dbpkg.Database = (*Postgres)(nil)

type database = dbpkg.BaseDatabase
type Postgres struct {
	*database
	port   int
	writer cliutil.Writer
	logger *slog.Logger
}

func (p *Postgres) CreateNew(args dbpkg.DatabaseArgs) dbpkg.Database {
	return NewPostgres(args)
}

func (p *Postgres) TypeName() string {
	return "Postgres"
}

func (p *Postgres) String() string {
	//TODO parse out database name from URL or DSN
	return p.TypeName()
}

func (*Postgres) Type() dbpkg.DatabaseType {
	return PostgresDatabase
}

func NewPostgres(args dbpkg.DatabaseArgs) *Postgres {
	pdb := &Postgres{}
	if args.Port != 0 {
		pdb.port = args.Port
	}
	pdb.database = dbpkg.NewBaseDatabase(pdb, args)
	return pdb
}

func (p *Postgres) isConnectionString(cs string) bool {
	_, err := p.formatConnectionString(cs)
	return err == nil
}

func (p *Postgres) Open() (err error) {
	var cs string
	p.writer.Printf("Using PostgreSQL database\n")
	p.logger.Info("Opening PostgreSQL database")
	cs, err = p.formatConnectionString(p.ConnectString())
	if err != nil {
		err = errors.Join(dbpkg.ErrInvalidConnString, err)
		goto end
	}
	p.DB, err = sql.Open("postgres", cs)
	if err != nil {
		err = errors.Join(dbpkg.ErrConnFailed, err)
		goto end
	}
end:
	return err
}

func (p *Postgres) Query(ctx dbpkg.Context, q string, params ...any) (*sql.Rows, error) {
	return p.database.Query(ctx, FormatQueryForPostgres(q), params...)
}

func (p *Postgres) CheckConnection(cs string) (err error) {
	if !p.isConnectionString(cs) {
		// TODO Make this check more robust
		goto end
	}
	err = p.database.CheckConnection(cs)
end:
	return err
}

// formatConnectionString injects or overrides the port in a Postgres connection string (URL or DSN format)
func (p *Postgres) formatConnectionString(cs string) (_ string, err error) {
	return FormatPGConnectString(cs, p.port)
}

var postgresPrefixRE = regexp.MustCompile(`^\s*postgres(ql)?://`)
var dsnFormatRE = regexp.MustCompile(`port=\\d+`)

// FormatPGConnectString injects or overrides the port in a Postgres connection string (URL or DSN format)
func FormatPGConnectString(cs string, port int) (_ string, err error) {
	if port == 0 {
		goto end
	}
	if postgresPrefixRE.MatchString(cs) {
		u, err := url.Parse(cs)
		if err != nil {
			goto end
		}
		if u == nil {
			err = fmt.Errorf("invalid PostgreSQL connection string: '%s'", cs)
			goto end
		}
		if u.Port() == "" {
			goto end
		}
		if u.Port() == strconv.Itoa(port) {
			goto end
		}
		u.Host = fmt.Sprintf("%s:%d", u.Hostname(), port)
		cs = u.String()
		goto end
	}

	// DSN format: add or replace port=...
	// TODO Can we do more to validate here?
	if dsnFormatRE.MatchString(cs) {
		cs = dsnFormatRE.ReplaceAllString(cs, fmt.Sprintf("port=%d", port))
		goto end
	}
	cs = fmt.Sprintf("%s port=%d", cs, port)
end:
	return cs, err
}

// FormatQueryForPostgres replaces ? in query w/numbered params in $n format
func FormatQueryForPostgres(q string) string {
	newQ := make([]byte, len(q)*2)
	pNum := 1
	i, j := 0, 0
	for i < len(q) {
		newQ[j] = q[i]
		if q[i] != '?' {
			i++
			j++
			continue
		}
		i++
		newQ[j] = '$'
		j++
		for n, c := range strconv.Itoa(pNum) {
			newQ[j+n] = byte(c)
			j++
		}
		pNum++
	}
	return string(newQ[:j])
}
