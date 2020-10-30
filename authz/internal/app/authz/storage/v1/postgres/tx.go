package postgres

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/openzipkin/zipkin-go"
)

type Tx interface {
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Commit() error
	Rebind(query string) string
}

var _ Tx = (*transcation)(nil)

type transcation struct {
	query []string
	span  zipkin.Span
	tx    *sqlx.Tx
}

func (t *transcation) Rebind(query string) string {
	return t.tx.Rebind(query)
}

func (t *transcation) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	t.addTxSpanTags(query)
	return t.tx.SelectContext(ctx, dest, query, args...)
}

func (t *transcation) Commit() error {
	if t.span != nil {
		t.span.Tag("sql.statement", strings.Join(t.query, "\n\n"))
	}
	return t.tx.Commit()
}

func (t *transcation) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	t.addTxSpanTags(query)
	return t.tx.GetContext(ctx, dest, query, args...)
}

func (t *transcation) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	t.addTxSpanTags(query)
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *transcation) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	t.addTxSpanTags(query)
	return t.tx.NamedExecContext(ctx, query, arg)
}

// NewDatabase creates a ThingDatabase instance
func NewTx(ctx context.Context, tx *sqlx.Tx) Tx {
	span := zipkin.SpanFromContext(ctx)
	if span != nil {
		span.Tag("span.kind", "client")
		span.Tag("peer.service", "postgres")
		span.Tag("db.type", "sql")
	}

	return &transcation{
		query: []string{},
		span:  span,
		tx:    tx,
	}
}

func (t *transcation) addTxSpanTags(query string) {
	t.query = append(t.query, query)
}
