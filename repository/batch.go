package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func executeBatchResults(results pgx.BatchResults, count int, action string) error {
	for i := 0; i < count; i++ {
		if _, err := results.Exec(); err != nil {
			if closeErr := results.Close(); closeErr != nil {
				return fmt.Errorf("%s: %w", action, errors.Join(err, fmt.Errorf("close batch: %w", closeErr)))
			}

			return fmt.Errorf("%s: %w", action, err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("close %s batch: %w", action, err)
	}

	return nil
}

func requireRowsAffected(tag pgconn.CommandTag, want int64, action string) error {
	if got := tag.RowsAffected(); got != want {
		return fmt.Errorf("%s: affected %d rows, want %d", action, got, want)
	}

	return nil
}
