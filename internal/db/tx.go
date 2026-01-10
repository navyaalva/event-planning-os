package db

import (
	"context"
	"database/sql"
	"fmt"
)

// RunTx runs fn inside a DB transaction.
// If fn returns an error, everything is rolled back.
func (q *Queries) RunTx(ctx context.Context, dbConn *sql.DB, fn func(qtx *Queries) error) error {
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	qtx := q.WithTx(tx)

	if err := fn(qtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rollback err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
