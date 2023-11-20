package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func init() {
	addMigration(createUsersTable)
}

func createUsersTable(ctx context.Context, tx pgx.Tx) error {
	query := `
		create table if not exists users (
			user_id text not null,
			email text not null,
			first_name text null,
			last_name text null,
			primary key (user_id)
		);
    `
	if _, err := tx.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to run create users table migration: %w", err)
	}
	return nil
}
