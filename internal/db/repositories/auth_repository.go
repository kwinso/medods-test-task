package repositories

import (
	"context"

	"github.com/kwinso/medods-test-task/internal/db"
)

type AuthRepository interface {
	CreateAuth(ctx context.Context, auth db.CreateAuthParams) (db.Auth, error)
}

type pgxAuthRepository struct {
	queries db.Queries
}

func NewPgxAuthRepository(conn db.DBTX) AuthRepository {
	return &pgxAuthRepository{
		queries: *db.New(conn),
	}
}

func (r *pgxAuthRepository) CreateAuth(ctx context.Context, auth db.CreateAuthParams) (db.Auth, error) {
	return r.queries.CreateAuth(ctx, auth)
}
