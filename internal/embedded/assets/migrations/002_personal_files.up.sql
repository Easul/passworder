-- 002_personal_files.up.sql
-- Personal file management (not tied to accounts)

CREATE TABLE IF NOT EXISTS personal_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    original_name TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    mime_type TEXT,
    size_bytes INTEGER NOT NULL DEFAULT 0,
    sha256 TEXT NOT NULL,
    file_type TEXT NOT NULL,
    is_markdown INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now')),
    is_deleted INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_personal_files_deleted ON personal_files(is_deleted);
CREATE INDEX IF NOT EXISTS idx_personal_files_type_deleted ON personal_files(file_type, is_deleted);
CREATE INDEX IF NOT EXISTS idx_personal_files_name ON personal_files(name);
