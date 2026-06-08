-- 001_init.up.sql
-- Initial schema for passworder v2

CREATE TABLE IF NOT EXISTS auth (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    password_hash TEXT NOT NULL,
    kdf_salt BLOB NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    icon TEXT,
    sort_order INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now')),
    is_deleted INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    website TEXT,
    username TEXT NOT NULL,
    password_encrypted BLOB NOT NULL,
    email TEXT,
    phone TEXT,
    notes TEXT,
    tags TEXT DEFAULT '[]',
    is_favorite INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER DEFAULT (strftime('%s', 'now')),
    is_deleted INTEGER DEFAULT 0,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE INDEX IF NOT EXISTS idx_accounts_category ON accounts(category_id);
CREATE INDEX IF NOT EXISTS idx_accounts_title ON accounts(title);
CREATE INDEX IF NOT EXISTS idx_accounts_username ON accounts(username);
CREATE INDEX IF NOT EXISTS idx_accounts_website ON accounts(website);
CREATE INDEX IF NOT EXISTS idx_accounts_favorite ON accounts(is_favorite);
CREATE INDEX IF NOT EXISTS idx_accounts_deleted ON accounts(is_deleted);

CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    original_name TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    mime_type TEXT,
    size_bytes INTEGER,
    sha256 TEXT,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    is_deleted INTEGER DEFAULT 0,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_attachments_account ON attachments(account_id);

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE TABLE IF NOT EXISTS reminders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    remind_at INTEGER NOT NULL,
    email TEXT,
    is_sent INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reminders_time ON reminders(remind_at);
CREATE INDEX IF NOT EXISTS idx_reminders_sent ON reminders(is_sent);

-- Insert default categories
INSERT INTO categories (name, icon, sort_order) VALUES 
    ('网站', 'globe', 1),
    ('工作', 'briefcase', 2),
    ('社交', 'users', 3),
    ('金融', 'credit-card', 4),
    ('游戏', 'gamepad', 5),
    ('其他', 'folder', 99);
