package repository

import (
	"database/sql"
	"passworder/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	attachmentCreateSQL = `INSERT INTO attachments 
		(account_id, original_name, stored_name, mime_type, size_bytes, sha256) 
		VALUES (?, ?, ?, ?, ?, ?)`
	attachmentDeleteSQL = `UPDATE attachments SET is_deleted = 1 WHERE id = ?`
	attachmentGetSQL    = `SELECT 
		id, account_id, original_name, stored_name, mime_type, size_bytes, sha256, created_at 
		FROM attachments WHERE id = ? AND is_deleted = 0`
	attachmentListByAccountSQL = `SELECT 
		id, account_id, original_name, stored_name, mime_type, size_bytes, sha256, created_at 
		FROM attachments WHERE account_id = ? AND is_deleted = 0 ORDER BY id`
	attachmentListByAccountNoDataSQL = `SELECT 
		id, account_id, original_name, stored_name, mime_type, size_bytes, sha256, created_at 
		FROM attachments WHERE account_id = ? AND is_deleted = 0 ORDER BY id`
)

type AttachmentRepository struct {
	db *sqlx.DB
}

func NewAttachmentRepository(db *sqlx.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(a *model.Attachment) error {
	result, err := r.db.Exec(attachmentCreateSQL,
		a.AccountID, a.OriginalName, a.StoredName, a.MimeType, a.SizeBytes, a.Sha256,
	)
	if err != nil {
		return err
	}
	a.ID, _ = result.LastInsertId()
	a.CreatedAt = model.Now()
	return nil
}

func (r *AttachmentRepository) Delete(id int64) error {
	_, err := r.db.Exec(attachmentDeleteSQL, id)
	return err
}

func (r *AttachmentRepository) Get(id int64) (*model.Attachment, error) {
	var a model.Attachment
	err := r.db.Get(&a, attachmentGetSQL, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func (r *AttachmentRepository) ListByAccount(accountID int64) ([]model.Attachment, error) {
	var attachments []model.Attachment
	err := r.db.Select(&attachments, attachmentListByAccountSQL, accountID)
	return attachments, err
}
