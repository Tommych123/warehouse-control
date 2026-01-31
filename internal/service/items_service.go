package service

import (
	"context"

	"github.com/jackc/pgx/v5"

	"warehouse/internal/domain"
	"warehouse/internal/repo"
)

type ItemsService struct {
	db   *repo.DB
	repo *repo.ItemsRepo
}

func NewItemsService(db *repo.DB, r *repo.ItemsRepo) *ItemsService {
	return &ItemsService{db: db, repo: r}
}

func (s *ItemsService) List(ctx context.Context, search string) ([]domain.Item, error) {
	return s.repo.List(ctx, search)
}

func (s *ItemsService) Create(ctx context.Context, actor string, role string, in domain.ItemCreate) (domain.Item, error) {
	tx, err := s.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Item{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := repo.SetUserContext(ctx, tx, actor, role); err != nil {
		return domain.Item{}, err
	}

	it, err := s.repo.Create(ctx, tx, in)
	if err != nil {
		return domain.Item{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Item{}, err
	}
	return it, nil
}

func (s *ItemsService) Update(ctx context.Context, actor string, role string, id int64, in domain.ItemUpdate) (domain.Item, error) {
	tx, err := s.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Item{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := repo.SetUserContext(ctx, tx, actor, role); err != nil {
		return domain.Item{}, err
	}

	it, err := s.repo.Update(ctx, tx, id, in)
	if err != nil {
		return domain.Item{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Item{}, err
	}
	return it, nil
}

func (s *ItemsService) Delete(ctx context.Context, actor string, role string, id int64) error {
	tx, err := s.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := repo.SetUserContext(ctx, tx, actor, role); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, tx, id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
