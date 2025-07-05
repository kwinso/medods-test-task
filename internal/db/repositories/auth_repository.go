package repositories

import "github.com/kwinso/medods-test-task/internal/db"

type AuthRepository interface {
	GetAuthByGUID(guid string) (*db.Auth, error)
}

type pgxAuthRepository struct {
	db db.DBTX
}

func NewPgxAuthRepository(db db.DBTX) AuthRepository {
	return &pgxAuthRepository{
		db: db,
	}
}

func (r *pgxAuthRepository) GetAuthByGUID(guid string) (*db.Auth, error) {
	return nil, nil
}
