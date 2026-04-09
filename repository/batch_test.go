package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeBatchResults struct {
	execCalls  int
	failOnExec int
	execErr    error
	closeErr   error
	closed     bool
}

func (f *fakeBatchResults) Exec() (pgconn.CommandTag, error) {
	f.execCalls++
	if f.failOnExec > 0 && f.execCalls == f.failOnExec {
		return pgconn.NewCommandTag(""), f.execErr
	}

	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

func (f *fakeBatchResults) Query() (pgx.Rows, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeBatchResults) QueryRow() pgx.Row {
	return nil
}

func (f *fakeBatchResults) Close() error {
	f.closed = true
	return f.closeErr
}

func TestExecuteBatchResults_ClosesBatchOnExecError(t *testing.T) {
	t.Parallel()

	execErr := errors.New("exec failed")
	results := &fakeBatchResults{
		failOnExec: 2,
		execErr:    execErr,
	}

	err := executeBatchResults(results, 3, "save snowfall")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, execErr) {
		t.Fatalf("expected wrapped exec error, got %v", err)
	}
	if !results.closed {
		t.Fatal("expected batch to be closed after exec error")
	}
}

func TestExecuteBatchResults_ReturnsCloseError(t *testing.T) {
	t.Parallel()

	closeErr := errors.New("close failed")
	results := &fakeBatchResults{closeErr: closeErr}

	err := executeBatchResults(results, 1, "save reading")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("expected wrapped close error, got %v", err)
	}
}

func TestRequireRowsAffected(t *testing.T) {
	t.Parallel()

	if err := requireRowsAffected(pgconn.NewCommandTag("UPDATE 1"), 1, "mark failed attempt retried"); err != nil {
		t.Fatalf("requireRowsAffected() unexpected error: %v", err)
	}

	err := requireRowsAffected(pgconn.NewCommandTag("UPDATE 0"), 1, "mark failed attempt retried")
	if err == nil {
		t.Fatal("expected error for zero rows affected")
	}
}
