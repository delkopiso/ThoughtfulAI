package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func init() {
	addMigration(createFavoritesTable)
}

func createFavoritesTable(ctx context.Context, tx pgx.Tx) error {
	query := `
		create table if not exists favorites (
			user_id text not null references users(user_id) on delete cascade,
			ticker text not null references securities(ticker) on delete cascade,
			primary key (user_id, ticker)
		);
    `
	if _, err := tx.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to run create favorites table migration: %w", err)
	}
	return nil
}
