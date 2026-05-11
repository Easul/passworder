package repository

import (
	"database/sql"
	"passworder/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	categoryCreateSQL = `INSERT INTO categories (name, icon, sort_order) VALUES (?, ?, ?)`
	categoryUpdateSQL = `UPDATE categories SET name = ?, icon = ?, sort_order = ?, updated_at = ? WHERE id = ?`
	categoryDeleteSQL = `UPDATE categories SET is_deleted = 1, updated_at = ? WHERE id = ?`
	categoryGetSQL    = `SELECT id, name, icon, sort_order, created_at, updated_at FROM categories WHERE id = ? AND is_deleted = 0`
	categoryListSQL   = `SELECT id, name, icon, sort_order, created_at, updated_at FROM categories WHERE is_deleted = 0 ORDER BY sort_order, id`
)

type CategoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(c *model.Category) error {
	result, err := r.db.Exec(categoryCreateSQL, c.Name, c.Icon, c.SortOrder)
	if err != nil {
		return err
	}
	c.ID, _ = result.LastInsertId()
	c.CreatedAt = model.Now()
	c.UpdatedAt = c.CreatedAt
	return nil
}

func (r *CategoryRepository) Update(c *model.Category) error {
	c.UpdatedAt = model.Now()
	_, err := r.db.Exec(categoryUpdateSQL, c.Name, c.Icon, c.SortOrder, c.UpdatedAt, c.ID)
	return err
}

func (r *CategoryRepository) Delete(id int64) error {
	_, err := r.db.Exec(categoryDeleteSQL, model.Now(), id)
	return err
}

func (r *CategoryRepository) Get(id int64) (*model.Category, error) {
	var c model.Category
	err := r.db.Get(&c, categoryGetSQL, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *CategoryRepository) List() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Select(&categories, categoryListSQL)
	return categories, err
}
