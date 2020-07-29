package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"

	"github.com/cage1016/gokit-istio-security/internal/app/authz/storage"
)

// nolint:lll
//go:generate go-bindata -pkg $GOPACKAGE -o sqls.bindata.go -ignore .*_test.rego -ignore Makefile -ignore README\.md sqls/...

// Config defines the options that are used when connecting to a PostgreSQL instance
type Config struct {
	Host        string
	Port        string
	User        string
	Pass        string
	Name        string
	SSLMode     string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
}

func (cfg Config) ToURL() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass, cfg.SSLMode, cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert)
}

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate
// failure.
func Connect(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.ToURL())
	if err != nil {
		return nil, err
	}

	if err := MigrateDB(db, cfg.User); err != nil {
		return nil, err
	}
	return db, nil
}

func MigrateDB(db *sqlx.DB, tableOwner string) error {
	te := NewEngine()
	sqls, err := te.AssetDir("sqls")
	if err != nil {
		return err
	}

	// add need override model
	mp := map[string]map[string]string{}
	mp["1.sql.up"] = map[string]string{"owner": tableOwner}

	f := func(key string) interface{} {
		if m, ok := mp[key]; ok {
			return m
		}
		return nil
	}

	mgs := []*migrate.Migration{}
	for i := 1; i <= len(sqls)/2; i++ {
		ukey := fmt.Sprintf("%d.sql.up", i)
		dkey := fmt.Sprintf("%d.sql.down", i)

		var up, down string
		var err error
		if up, err = te.Execute(ukey, f(ukey)); err != nil {
			return err
		}
		if down, err = te.Execute(dkey, f(dkey)); err != nil {
			return err
		}
		mgs = append(mgs, &migrate.Migration{
			Id:   fmt.Sprintf("authz_%d", i),
			Up:   []string{up},
			Down: []string{down},
		})
	}

	migrations := &migrate.MemoryMigrationSource{
		Migrations: mgs,
	}

	_, err = migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}

// ProcessError is used to translate DB-related errors into the error types
// defined for our storage implementations.
func ProcessError(err error) error {
	if err == sql.ErrNoRows {
		return storage.ErrNotFound
	}

	// The not found on json unmarshall case
	if strings.HasPrefix(err.Error(), "sql: Scan error on column index 0") &&
		strings.HasSuffix(err.Error(), "not found") {
		return storage.ErrNotFound
	}

	if err, ok := err.(*pq.Error); ok {
		return parsePQError(err)
	}

	return err
}

// parsePQError is able to parse pq specific errors into storage interface errors.
func parsePQError(e *pq.Error) error {
	switch e.Code {
	case "23505": // Unique violation
		return storage.ErrConflict
	case "P0002": // Not found in plpgsql ("no_data_found")
		return storage.ErrNotFound
	case "20000": // Not found
		return storage.ErrNotFound
	}

	return storage.ErrDatabase
}
