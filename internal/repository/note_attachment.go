package repository

import (
	"database/sql"
	"passworder/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	noteAttachmentCreateSQL = `INSERT INTO note_attachments 
		(personal_file_id, original_name, stored_name, mime_type, size_bytes, sha256, file_type, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	noteAttachmentDeleteSQL       = `UPDATE note_attachments SET is_deleted = 1 WHERE id = ?`
	noteAttachmentGetSQL          = `SELECT id, personal_file_id, original_name, stored_name, mime_type, size_bytes, sha256, file_type, created_at, is_deleted FROM note_attachments WHERE id = ? AND is_deleted = 0`
	noteAttachmentListSQL         = `SELECT id, personal_file_id, original_name, stored_name, mime_type, size_bytes, sha256, file_type, created_at, is_deleted FROM note_attachments WHERE personal_file_id = ? AND is_deleted = 0 ORDER BY created_at DESC`
	noteAttachmentDeleteByFileSQL = `UPDATE note_attachments SET is_deleted = 1 WHERE personal_file_id = ?`
)

type NoteAttachmentRepository struct {
	db *sqlx.DB
}

func NewNoteAttachmentRepository(db *sqlx.DB) *NoteAttachmentRepository {
	return &NoteAttachmentRepository{db: db}
}

func (r *NoteAttachmentRepository) Create(a *model.NoteAttachment) error {
	result, err := r.db.Exec(noteAttachmentCreateSQL, a.PersonalFileID, a.OriginalName, a.StoredName, a.MimeType, a.SizeBytes, a.Sha256, a.FileType, a.CreatedAt)
	if err != nil {
		return err
	}
	a.ID, _ = result.LastInsertId()
	return nil
}

func (r *NoteAttachmentRepository) Delete(id int64) error {
	_, err := r.db.Exec(noteAttachmentDeleteSQL, id)
	return err
}

func (r *NoteAttachmentRepository) Get(id int64) (*model.NoteAttachment, error) {
	var a model.NoteAttachment
	err := r.db.Get(&a, noteAttachmentGetSQL, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

func (r *NoteAttachmentRepository) ListByFile(personalFileID int64) ([]model.NoteAttachment, error) {
	var attachments []model.NoteAttachment
	err := r.db.Select(&attachments, noteAttachmentListSQL, personalFileID)
	return attachments, err
}

func (r *NoteAttachmentRepository) DeleteByFile(personalFileID int64) error {
	_, err := r.db.Exec(noteAttachmentDeleteByFileSQL, personalFileID)
	return err
}
