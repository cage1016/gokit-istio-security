package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/openzipkin/zipkin-go"
)

var _ Database = (*database)(nil)

type database struct {
	db *sqlx.DB
}

// Database provides a database interface
type Database interface {
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	MustBeginTx(ctx context.Context, opts *sql.TxOptions) Tx
	Rebind(query string) string
}

// NewDatabase creates a ThingDatabase instance
func NewDatabase(db *sqlx.DB) Database {
	return &database{
		db: db,
	}
}

func (d database) Rebind(query string) string {
	return d.db.Rebind(query)
}

func (d database) MustBeginTx(ctx context.Context, opts *sql.TxOptions) Tx {
	return NewTx(ctx, d.db.MustBeginTx(ctx, opts))
}

func (d database) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	addSpanTags(ctx, query)
	return d.db.NamedExecContext(ctx, query, arg)
}

func (d database) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	addSpanTags(ctx, query)
	return d.db.SelectContext(ctx, dest, query, args...)
}

func (d database) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	addSpanTags(ctx, query)
	return d.db.GetContext(ctx, dest, query, args...)
}

func (d database) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	addSpanTags(ctx, query)
	return d.db.QueryRowxContext(ctx, query, args...)
}

func (d database) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	addSpanTags(ctx, query)
	return d.db.QueryxContext(ctx, query, args)
}

func addSpanTags(ctx context.Context, query string) {
	span := zipkin.SpanFromContext(ctx)
	if span != nil {
		span.Tag("sql.statement", query)
		span.Tag("span.kind", "client")
		span.Tag("peer.service", "postgres")
		span.Tag("db.type", "sql")
	}
}
