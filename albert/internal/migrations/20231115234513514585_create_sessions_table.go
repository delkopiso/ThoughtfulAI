package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func init() {
	addMigration(createSessionsTable)
}

func createSessionsTable(ctx context.Context, tx pgx.Tx) error {
	query := `
		create table if not exists sessions (
			token text not null,
			data bytea not null,
			expiry timestamptz not null,
			primary key (token)
		);
    `
	if _, err := tx.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to run create sessions table migration: %w", err)
	}
	return nil
}
