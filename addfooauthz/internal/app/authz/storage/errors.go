package storage

import "github.com/cage1016/gokit-istio-security/internal/pkg/errors"

// Error responses common to all storage adapters, be it v1, v2, memstore, postgres, etc.
var (
	// ErrNotFound is returned when a requested policy wasn't found.
	ErrNotFound = errors.New("not found")

	// ErrConflict indicates that the object being created already exists.
	ErrConflict = errors.New("conflict")

	// ErrDatabase results from unexpected database errors.
	ErrDatabase = errors.New("database internal")
)

// TxCommitError occurs when the database attempts to commit a transaction and
// fails.
type TxCommitError struct {
	underlying error
}

func NewTxCommitError(e error) error {
	return &TxCommitError{underlying: e}
}

func (e *TxCommitError) Error() string {
	return "commit db transaction: " + e.underlying.Error()
}

// MissingFieldError occurs when a required field was not passed.
type MissingFieldError struct {
	field string
}

func NewMissingFieldError(f string) error {
	return &MissingFieldError{field: f}
}

func (e *MissingFieldError) Error() string {
	return "must supply policy " + e.field
}

type ForeignKeyError struct {
	Msg string
}

func (e *ForeignKeyError) Error() string {
	return e.Msg
}
