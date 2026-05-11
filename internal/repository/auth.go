package repository

import (
	"passworder/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	authCreateSQL = `INSERT OR REPLACE INTO auth 
		(id, password_hash, kdf_salt, created_at, updated_at) 
		VALUES (1, ?, ?, ?, ?)`
	authGetSQL = `SELECT password_hash, kdf_salt, created_at, updated_at FROM auth WHERE id = 1`
)

type AuthRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) Save(auth *model.Auth) error {
	_, err := r.db.Exec(authCreateSQL, auth.PasswordHash, auth.KDFSalt, auth.CreatedAt, auth.UpdatedAt)
	return err
}

func (r *AuthRepository) Get() (*model.Auth, error) {
	var auth model.Auth
	auth.ID = 1
	err := r.db.QueryRowx(authGetSQL).Scan(&auth.PasswordHash, &auth.KDFSalt, &auth.CreatedAt, &auth.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &auth, nil
}
