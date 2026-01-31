package repo

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"warehouse/internal/domain"
)

var ErrNotFound = errors.New("not found")

type ItemsRepo struct {
	pool *pgxpool.Pool
}

func NewItemsRepo(db *DB) *ItemsRepo {
	return &ItemsRepo{pool: db.Pool}
}

func (r *ItemsRepo) List(ctx context.Context, search string) ([]domain.Item, error) {
	search = strings.TrimSpace(search)

	q := `
select id, sku, name, qty, location, created_at, updated_at
from items
`
	args := []any{}
	if search != "" {
		q += `where (sku ilike $1 or name ilike $1 or coalesce(location,'') ilike $1)`
		args = append(args, "%"+search+"%")
	}
	q += ` order by id asc`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Item, 0)
	for rows.Next() {
		var it domain.Item
		if err := rows.Scan(&it.ID, &it.SKU, &it.Name, &it.Qty, &it.Location, &it.Created, &it.Updated); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *ItemsRepo) Get(ctx context.Context, id int64) (domain.Item, error) {
	q := `select id, sku, name, qty, location, created_at, updated_at from items where id=$1`
	var it domain.Item
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.SKU, &it.Name, &it.Qty, &it.Location, &it.Created, &it.Updated)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, ErrNotFound
	}
	return it, err
}

func (r *ItemsRepo) Create(ctx context.Context, tx pgx.Tx, in domain.ItemCreate) (domain.Item, error) {
	q := `
insert into items(sku, name, qty, location)
values ($1,$2,$3,$4)
returning id, sku, name, qty, location, created_at, updated_at
`
	var it domain.Item
	err := tx.QueryRow(ctx, q, in.SKU, in.Name, in.Qty, in.Location).
		Scan(&it.ID, &it.SKU, &it.Name, &it.Qty, &it.Location, &it.Created, &it.Updated)
	return it, err
}

func (r *ItemsRepo) Update(ctx context.Context, tx pgx.Tx, id int64, in domain.ItemUpdate) (domain.Item, error) {
	q := `
update items
set sku=$2, name=$3, qty=$4, location=$5
where id=$1
returning id, sku, name, qty, location, created_at, updated_at
`
	var it domain.Item
	err := tx.QueryRow(ctx, q, id, in.SKU, in.Name, in.Qty, in.Location).
		Scan(&it.ID, &it.SKU, &it.Name, &it.Qty, &it.Location, &it.Created, &it.Updated)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, ErrNotFound
	}
	return it, err
}

func (r *ItemsRepo) Delete(ctx context.Context, tx pgx.Tx, id int64) error {
	ct, err := tx.Exec(ctx, `delete from items where id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
