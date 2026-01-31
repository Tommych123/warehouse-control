package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func SetUserContext(ctx context.Context, tx pgx.Tx, username string, role string) error {
	if _, err := tx.Exec(ctx, "select set_config('app.user', $1, true)", username); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "select set_config('app.role', $1, true)", role); err != nil {
		return err
	}
	return nil
}
