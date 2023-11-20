package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbMigrations []dbMigration

type dbMigration func(ctx context.Context, tx pgx.Tx) error

func addMigration(migration dbMigration) {
	dbMigrations = append(dbMigrations, migration)
}

func Run(ctx context.Context, dbConn *pgxpool.Pool) (err error) {
	var tx pgx.Tx
	if tx, err = dbConn.BeginTx(ctx, pgx.TxOptions{}); err != nil {
		return fmt.Errorf("failed to start migration transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback transaction after err %w due to %w", err, rollbackErr)
			}
		}
	}()

	for _, migration := range dbMigrations {
		if err = migration(ctx, tx); err != nil {
			return
		}
	}
	return tx.Commit(ctx)
}
