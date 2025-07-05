package repositories

import (
	"context"

	"github.com/kwinso/medods-test-task/internal/db"
)

type AuthRepository interface {
	GetNextAuthId(ctx context.Context) (int64, error)
	CreateAuth(ctx context.Context, auth db.CreateAuthParams) (db.Auth, error)
	GetAuthById(ctx context.Context, id int32) (db.Auth, error)
	// GetAuthByRefreshToken(ctx context.Context, refreshToken string) (db.Auth, error)
	DeleteAuthById(ctx context.Context, id int32) error
	UpdateAuthRefreshToken(ctx context.Context, id int32, refreshToken string) error
}

type pgxAuthRepository struct {
	queries db.Queries
}

func NewPgxAuthRepository(conn db.DBTX) AuthRepository {
	return &pgxAuthRepository{
		queries: *db.New(conn),
	}
}

func (r *pgxAuthRepository) GetNextAuthId(ctx context.Context) (int64, error) {
	return r.queries.GetNextAuthId(ctx)
}

func (r *pgxAuthRepository) CreateAuth(ctx context.Context, auth db.CreateAuthParams) (db.Auth, error) {
	return r.queries.CreateAuth(ctx, auth)
}

func (r *pgxAuthRepository) GetAuthById(ctx context.Context, id int32) (db.Auth, error) {
	return r.queries.GetAuthById(ctx, id)
}

func (r *pgxAuthRepository) GetAuthByRefreshToken(ctx context.Context, refreshToken string) (db.Auth, error) {
	return r.queries.GetAuthByRefreshToken(ctx, refreshToken)
}

func (r *pgxAuthRepository) DeleteAuthById(ctx context.Context, id int32) error {
	return r.queries.DeleteAuthById(ctx, id)
}

func (r *pgxAuthRepository) UpdateAuthRefreshToken(ctx context.Context, id int32, refreshToken string) error {
	return r.queries.UpdateAuthRefreshToken(ctx, db.UpdateAuthRefreshTokenParams{
		ID:               id,
		RefreshTokenHash: refreshToken,
	})
}
