package repository

import (
	"database/sql"
	"fmt"
	"passworder/internal/model"
	"strings"

	"github.com/jmoiron/sqlx"
)

const (
	accountCreateSQL = `INSERT INTO accounts
		(category_id, title, website, username, password_encrypted, email, reminder_email, remind_at, registration_time, registration_notes, phone, notes, tags, is_favorite, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	accountUpdateSQL = `UPDATE accounts SET
		category_id = ?, title = ?, website = ?, username = ?, password_encrypted = ?,
		email = ?, reminder_email = ?, remind_at = ?, registration_time = ?, registration_notes = ?, phone = ?, notes = ?, tags = ?, is_favorite = ?, status = ?, updated_at = ?
		WHERE id = ?`
	accountDeleteSQL = `UPDATE accounts SET is_deleted = 1, updated_at = ? WHERE id = ?`
	accountGetSQL    = `SELECT
		a.id, a.category_id, a.title, a.website, a.username, a.password_encrypted,
		a.email, a.reminder_email, COALESCE(r.remind_at, a.remind_at) AS remind_at, a.registration_time, a.registration_notes, a.phone, a.notes, a.tags, a.is_favorite, a.status, a.created_at, a.updated_at,
		COALESCE(r.period_type, '') AS reminder_period_type,
		COALESCE(r.period_value, 0) AS reminder_period_value,
		CASE
			WHEN r.id IS NULL THEN 'none'
			WHEN COALESCE(r.period_type, '') = '' AND (r.is_sent = 1 OR r.remind_at <= ?) THEN 'sent'
			WHEN COALESCE(r.period_type, '') = '' THEN 'pending'
			WHEN r.is_sent = 1 AND r.next_remind_at > ? THEN 'sent'
			ELSE 'pending'
		END AS reminder_status
		FROM accounts a
		LEFT JOIN reminders r ON r.account_id = a.id
		WHERE a.id = ? AND a.is_deleted = 0`
	accountListSQL = `SELECT
		a.id, a.category_id, a.title, a.website, a.username, a.password_encrypted,
		a.email, a.reminder_email, a.remind_at, a.registration_time, a.registration_notes, a.phone, a.notes, a.tags, a.is_favorite, a.status, a.created_at, a.updated_at,
		COALESCE(r.period_type, '') AS reminder_period_type,
		COALESCE(r.period_value, 0) AS reminder_period_value,
		CASE
			WHEN r.id IS NULL THEN 'none'
			WHEN COALESCE(r.period_type, '') = '' AND (r.is_sent = 1 OR r.remind_at <= ?) THEN 'sent'
			WHEN COALESCE(r.period_type, '') = '' THEN 'pending'
			WHEN r.is_sent = 1 AND r.next_remind_at > ? THEN 'sent'
			ELSE 'pending'
		END as reminder_status
	FROM accounts a
	LEFT JOIN reminders r ON r.account_id = a.id
	WHERE a.is_deleted = 0
	ORDER BY a.is_favorite DESC, a.title`
	accountSearchSQL = `SELECT
		a.id, a.category_id, a.title, a.website, a.username, a.password_encrypted,
		a.email, a.reminder_email, a.remind_at, a.registration_time, a.registration_notes, a.phone, a.notes, a.tags, a.is_favorite, a.status, a.created_at, a.updated_at,
		COALESCE(r.period_type, '') AS reminder_period_type,
		COALESCE(r.period_value, 0) AS reminder_period_value,
		CASE
			WHEN r.id IS NULL THEN 'none'
			WHEN COALESCE(r.period_type, '') = '' AND (r.is_sent = 1 OR r.remind_at <= ?) THEN 'sent'
			WHEN COALESCE(r.period_type, '') = '' THEN 'pending'
			WHEN r.is_sent = 1 AND r.next_remind_at > ? THEN 'sent'
			ELSE 'pending'
		END as reminder_status
		FROM accounts a
		LEFT JOIN reminders r ON r.account_id = a.id
		WHERE a.is_deleted = 0 AND (
			a.title LIKE ? OR a.website LIKE ? OR a.username LIKE ? OR a.email LIKE ? OR a.notes LIKE ? OR a.registration_notes LIKE ?
		)
		ORDER BY a.is_favorite DESC, a.title`
	accountByCategorySQL = `SELECT
		a.id, a.category_id, a.title, a.website, a.username, a.password_encrypted,
		a.email, a.reminder_email, a.remind_at, a.registration_time, a.registration_notes, a.phone, a.notes, a.tags, a.is_favorite, a.status, a.created_at, a.updated_at,
		COALESCE(r.period_type, '') AS reminder_period_type,
		COALESCE(r.period_value, 0) AS reminder_period_value,
		CASE
			WHEN r.id IS NULL THEN 'none'
			WHEN COALESCE(r.period_type, '') = '' AND (r.is_sent = 1 OR r.remind_at <= ?) THEN 'sent'
			WHEN COALESCE(r.period_type, '') = '' THEN 'pending'
			WHEN r.is_sent = 1 AND r.next_remind_at > ? THEN 'sent'
			ELSE 'pending'
		END as reminder_status
		FROM accounts a
		LEFT JOIN reminders r ON r.account_id = a.id
		WHERE a.is_deleted = 0 AND a.category_id = ?
		ORDER BY a.title`
)

type AccountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(a *model.Account) error {
	if a.CreatedAt == 0 {
		a.CreatedAt = model.Now()
	}
	if a.UpdatedAt == 0 {
		a.UpdatedAt = a.CreatedAt
	}
	result, err := r.db.Exec(accountCreateSQL,
		a.CategoryID, a.Title, a.Website, a.Username, a.PasswordEncrypted,
		a.Email, a.ReminderEmail, a.RemindAt, a.RegistrationTime, a.RegistrationNotes, a.Phone, a.Notes, a.Tags, a.IsFavorite, a.Status, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return err
	}
	a.ID, _ = result.LastInsertId()
	return nil
}

func (r *AccountRepository) Update(a *model.Account) error {
	a.UpdatedAt = model.Now()
	_, err := r.db.Exec(accountUpdateSQL,
		a.CategoryID, a.Title, a.Website, a.Username, a.PasswordEncrypted,
		a.Email, a.ReminderEmail, a.RemindAt, a.RegistrationTime, a.RegistrationNotes, a.Phone, a.Notes, a.Tags, a.IsFavorite, a.Status, a.UpdatedAt, a.ID,
	)
	return err
}

func (r *AccountRepository) Delete(id int64) error {
	_, err := r.db.Exec(accountDeleteSQL, model.Now(), id)
	return err
}

func (r *AccountRepository) Get(id int64, now int64) (*model.Account, error) {
	var a model.Account
	err := r.db.Get(&a, accountGetSQL, now, now, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func (r *AccountRepository) List(now int64) ([]model.Account, error) {
	var accounts []model.Account
	err := r.db.Select(&accounts, accountListSQL, now, now)
	return accounts, err
}

func (r *AccountRepository) Search(query string, now int64) ([]model.Account, error) {
	like := fmt.Sprintf("%%%s%%", strings.TrimSpace(query))
	var accounts []model.Account
	err := r.db.Select(&accounts, accountSearchSQL, now, now, like, like, like, like, like, like)
	return accounts, err
}

func (r *AccountRepository) ListByCategory(categoryID int64, now int64) ([]model.Account, error) {
	var accounts []model.Account
	err := r.db.Select(&accounts, accountByCategorySQL, now, now, categoryID)
	return accounts, err
}
