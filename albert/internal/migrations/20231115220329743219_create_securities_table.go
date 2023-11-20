package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func init() {
	addMigration(createSecuritiesTable)
}

func createSecuritiesTable(ctx context.Context, tx pgx.Tx) error {
	query := `
		create table if not exists securities (
			ticker text not null,
			name text not null,
			last_price numeric null,
			last_price_updated timestamptz null,
			primary key (ticker)
		);

		create index if not exists idx_search_securities_name on securities (lower(name));
		create index if not exists idx_search_securities_ticker on securities (lower(ticker));
    `
	if _, err := tx.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to run create securities table migration: %w", err)
	}
	return nil
}
