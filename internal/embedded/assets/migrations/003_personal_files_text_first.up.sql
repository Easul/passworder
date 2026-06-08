DROP INDEX IF EXISTS idx_personal_files_deleted;
DROP INDEX IF EXISTS idx_personal_files_type_deleted;
DROP INDEX IF EXISTS idx_personal_files_name;
DROP INDEX IF EXISTS idx_personal_files_title;

DROP TABLE IF EXISTS personal_files_old;

ALTER TABLE personal_files RENAME TO personal_files_old;

CREATE TABLE IF NOT EXISTS personal_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    body_format TEXT NOT NULL DEFAULT 'text' CHECK (body_format IN ('markdown', 'text')),
    original_name TEXT NOT NULL DEFAULT '',
    stored_name TEXT NOT NULL DEFAULT '',
    mime_type TEXT NOT NULL DEFAULT '',
    size_bytes INTEGER NOT NULL DEFAULT 0,
    sha256 TEXT NOT NULL DEFAULT '',
    file_type TEXT NOT NULL DEFAULT 'none',
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now')),
    is_deleted INTEGER DEFAULT 0
);

INSERT INTO personal_files (
    id,
    title,
    body,
    body_format,
    original_name,
    stored_name,
    mime_type,
    size_bytes,
    sha256,
    file_type,
    created_at,
    updated_at,
    is_deleted
)
SELECT
    id,
    name,
    '',
    CASE WHEN is_markdown = 1 THEN 'markdown' ELSE 'text' END,
    original_name,
    stored_name,
    COALESCE(mime_type, ''),
    COALESCE(size_bytes, 0),
    COALESCE(sha256, ''),
    COALESCE(file_type, 'none'),
    created_at,
    updated_at,
    is_deleted
FROM personal_files_old;

DROP TABLE IF EXISTS personal_files_old;

CREATE INDEX IF NOT EXISTS idx_personal_files_deleted ON personal_files(is_deleted);
CREATE INDEX IF NOT EXISTS idx_personal_files_type_deleted ON personal_files(file_type, is_deleted);
CREATE INDEX IF NOT EXISTS idx_personal_files_title ON personal_files(title);
