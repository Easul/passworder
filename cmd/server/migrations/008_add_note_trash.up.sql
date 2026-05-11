ALTER TABLE personal_files ADD COLUMN deleted_at INTEGER;

CREATE INDEX IF NOT EXISTS idx_personal_files_deleted_at ON personal_files(deleted_at);
