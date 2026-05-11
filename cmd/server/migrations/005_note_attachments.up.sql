CREATE TABLE IF NOT EXISTS note_attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    personal_file_id INTEGER NOT NULL,
    original_name TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    mime_type TEXT,
    size_bytes INTEGER,
    sha256 TEXT,
    file_type TEXT NOT NULL DEFAULT 'other',
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    is_deleted INTEGER DEFAULT 0,
    FOREIGN KEY (personal_file_id) REFERENCES personal_files(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_note_attachments_file ON note_attachments(personal_file_id);
CREATE INDEX IF NOT EXISTS idx_note_attachments_deleted ON note_attachments(is_deleted);

INSERT INTO note_attachments (personal_file_id, original_name, stored_name, mime_type, size_bytes, sha256, file_type, created_at)
SELECT 
    id,
    original_name,
    stored_name,
    mime_type,
    size_bytes,
    sha256,
    file_type,
    created_at
FROM personal_files
WHERE stored_name != '' AND is_deleted = 0;
