package store

import (
	"context"
	db "diploma/internal/store/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Pool *pgxpool.Pool
	Q    *db.Queries
}

func New(p *pgxpool.Pool) *Store { return &Store{p, db.New(p)} }
func (s *Store) Begin(ctx context.Context) (pgx.Tx, *db.Queries, error) {
	t, e := s.Pool.Begin(ctx)
	if e != nil {
		return nil, nil, e
	}
	return t, db.New(t), nil
}

func (s *Store) FindingTarget(ctx context.Context, id string) (string, error) {
	return s.Q.FindingTarget(ctx, id)
}
