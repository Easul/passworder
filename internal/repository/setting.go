package repository

import (
	"passworder/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	settingGetSQL    = `SELECT key, value, updated_at FROM settings WHERE key = ?`
	settingSetSQL    = `INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, ?)`
	settingDeleteSQL = `DELETE FROM settings WHERE key = ?`
	settingListSQL   = `SELECT key, value, updated_at FROM settings ORDER BY key`
)

type SettingRepository struct {
	db *sqlx.DB
}

func NewSettingRepository(db *sqlx.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) Get(key string) (*model.Setting, error) {
	var s model.Setting
	err := r.db.Get(&s, settingGetSQL, key)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SettingRepository) Set(key, value string) error {
	_, err := r.db.Exec(settingSetSQL, key, value, model.Now())
	return err
}

func (r *SettingRepository) Delete(key string) error {
	_, err := r.db.Exec(settingDeleteSQL, key)
	return err
}

func (r *SettingRepository) List() ([]model.Setting, error) {
	var settings []model.Setting
	err := r.db.Select(&settings, settingListSQL)
	return settings, err
}
