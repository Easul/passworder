package repository

import (
	"database/sql"
	"passworder/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	personalFileCreateSQL = `INSERT INTO personal_files 
		(title, remarks, body, body_format, original_name, stored_name, mime_type, size_bytes, sha256, file_type, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	personalFileCreateImportedSQL = `INSERT INTO personal_files 
		(id, title, remarks, body, body_format, original_name, stored_name, mime_type, size_bytes, sha256, file_type, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	personalFileUpdateTitleSQL              = `UPDATE personal_files SET title = ?, updated_at = ? WHERE id = ?`
	personalFileUpdateBodySQL               = `UPDATE personal_files SET body = ?, body_format = ?, updated_at = ? WHERE id = ?`
	personalFileUpdateSQL                   = `UPDATE personal_files SET title = ?, remarks = ?, body = ?, body_format = ?, updated_at = ? WHERE id = ?`
	personalFileUpdateContentSQL            = `UPDATE personal_files SET size_bytes = ?, sha256 = ?, updated_at = ? WHERE id = ?`
	personalFileDeleteSQL                   = `UPDATE personal_files SET is_deleted = 1, deleted_at = ?, updated_at = ? WHERE id = ? AND is_deleted = 0`
	personalFileRestoreSQL                  = `UPDATE personal_files SET is_deleted = 0, deleted_at = NULL, updated_at = ? WHERE id = ? AND is_deleted = 1`
	personalFileGetSQL                      = `SELECT pf.id, pf.title, pf.remarks, pf.body, pf.body_format, pf.original_name, pf.stored_name, pf.mime_type, pf.size_bytes, pf.sha256, pf.file_type, pf.created_at, pf.updated_at, pf.deleted_at, pf.is_deleted, COALESCE((SELECT COUNT(1) FROM note_attachments na WHERE na.personal_file_id = pf.id AND na.is_deleted = 0), 0) AS attachment_count FROM personal_files pf WHERE pf.id = ? AND pf.is_deleted = 0`
	personalFileListSQL                     = `SELECT pf.id, pf.title, pf.remarks, pf.body, pf.body_format, pf.original_name, pf.stored_name, pf.mime_type, pf.size_bytes, pf.sha256, pf.file_type, pf.created_at, pf.updated_at, pf.deleted_at, pf.is_deleted, COALESCE((SELECT COUNT(1) FROM note_attachments na WHERE na.personal_file_id = pf.id AND na.is_deleted = 0), 0) AS attachment_count FROM personal_files pf WHERE pf.is_deleted = 0 ORDER BY pf.created_at DESC`
	personalFileListDeletedSQL              = `SELECT pf.id, pf.title, pf.remarks, pf.body, pf.body_format, pf.original_name, pf.stored_name, pf.mime_type, pf.size_bytes, pf.sha256, pf.file_type, pf.created_at, pf.updated_at, pf.deleted_at, pf.is_deleted, COALESCE((SELECT COUNT(1) FROM note_attachments na WHERE na.personal_file_id = pf.id AND na.is_deleted = 0), 0) AS attachment_count FROM personal_files pf WHERE pf.is_deleted = 1 ORDER BY pf.deleted_at DESC, pf.updated_at DESC`
	personalFileListByTypeSQL               = `SELECT pf.id, pf.title, pf.remarks, pf.body, pf.body_format, pf.original_name, pf.stored_name, pf.mime_type, pf.size_bytes, pf.sha256, pf.file_type, pf.created_at, pf.updated_at, pf.deleted_at, pf.is_deleted, COALESCE((SELECT COUNT(1) FROM note_attachments na WHERE na.personal_file_id = pf.id AND na.is_deleted = 0), 0) AS attachment_count FROM personal_files pf WHERE pf.file_type = ? AND pf.is_deleted = 0 ORDER BY pf.created_at DESC`
	personalFileDeleteAttachmentsByTrashSQL = `DELETE FROM note_attachments WHERE personal_file_id IN (SELECT id FROM personal_files WHERE is_deleted = 1)`
	personalFileEmptyTrashSQL               = `DELETE FROM personal_files WHERE is_deleted = 1`
	personalFileHardDeleteSQL               = `DELETE FROM personal_files WHERE id = ?`
	personalFileDeleteAttachmentsByIDSQL    = `DELETE FROM note_attachments WHERE personal_file_id = ?`
)

type PersonalFileRepository struct {
	db *sqlx.DB
}

func NewPersonalFileRepository(db *sqlx.DB) *PersonalFileRepository {
	return &PersonalFileRepository{db: db}
}

func (r *PersonalFileRepository) Create(f *model.PersonalFile) error {
	result, err := r.db.Exec(personalFileCreateSQL, f.Title, f.Remarks, f.Body, f.BodyFormat, f.OriginalName, f.StoredName, f.MimeType, f.SizeBytes, f.Sha256, f.FileType, f.CreatedAt, f.UpdatedAt)
	if err != nil {
		return err
	}
	f.ID, _ = result.LastInsertId()
	return nil
}

func (r *PersonalFileRepository) CreateImported(f *model.PersonalFile) error {
	_, err := r.db.Exec(personalFileCreateImportedSQL, f.ID, f.Title, f.Remarks, f.Body, f.BodyFormat, f.OriginalName, f.StoredName, f.MimeType, f.SizeBytes, f.Sha256, f.FileType, f.CreatedAt, f.UpdatedAt)
	return err
}

func (r *PersonalFileRepository) UpdateTitle(id int64, title string) error {
	_, err := r.db.Exec(personalFileUpdateTitleSQL, title, model.Now(), id)
	return err
}

func (r *PersonalFileRepository) UpdateBody(id int64, body, bodyFormat string) error {
	_, err := r.db.Exec(personalFileUpdateBodySQL, body, bodyFormat, model.Now(), id)
	return err
}

func (r *PersonalFileRepository) Update(id int64, title, remarks, body, bodyFormat string) error {
	_, err := r.db.Exec(personalFileUpdateSQL, title, remarks, body, bodyFormat, model.Now(), id)
	return err
}

func (r *PersonalFileRepository) UpdateContent(id int64, size int64, sha256 string) error {
	_, err := r.db.Exec(personalFileUpdateContentSQL, size, sha256, model.Now(), id)
	return err
}

func (r *PersonalFileRepository) Delete(id int64) error {
	now := model.Now()
	_, err := r.db.Exec(personalFileDeleteSQL, now, now, id)
	return err
}

func (r *PersonalFileRepository) Restore(id int64) error {
	_, err := r.db.Exec(personalFileRestoreSQL, model.Now(), id)
	return err
}

func (r *PersonalFileRepository) Get(id int64) (*model.PersonalFile, error) {
	var f model.PersonalFile
	err := r.db.Get(&f, personalFileGetSQL, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &f, err
}

func (r *PersonalFileRepository) List() ([]model.PersonalFile, error) {
	files := make([]model.PersonalFile, 0)
	err := r.db.Select(&files, personalFileListSQL)
	return files, err
}

func (r *PersonalFileRepository) ListDeleted() ([]model.PersonalFile, error) {
	files := make([]model.PersonalFile, 0)
	err := r.db.Select(&files, personalFileListDeletedSQL)
	return files, err
}

func (r *PersonalFileRepository) ListByType(fileType string) ([]model.PersonalFile, error) {
	files := make([]model.PersonalFile, 0)
	err := r.db.Select(&files, personalFileListByTypeSQL, fileType)
	return files, err
}

func (r *PersonalFileRepository) EmptyTrash() error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(personalFileDeleteAttachmentsByTrashSQL); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.Exec(personalFileEmptyTrashSQL); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *PersonalFileRepository) HardDelete(id int64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(personalFileDeleteAttachmentsByIDSQL, id); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.Exec(personalFileHardDeleteSQL, id); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
