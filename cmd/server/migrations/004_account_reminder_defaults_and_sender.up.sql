ALTER TABLE accounts ADD COLUMN reminder_email TEXT NOT NULL DEFAULT '';
ALTER TABLE accounts ADD COLUMN remind_at INTEGER NOT NULL DEFAULT 0;

UPDATE accounts
SET reminder_email = COALESCE((
    SELECT r.email
    FROM reminders r
    WHERE r.account_id = accounts.id
    ORDER BY r.remind_at DESC, r.id DESC
    LIMIT 1
), '');

UPDATE accounts
SET remind_at = COALESCE((
    SELECT r.remind_at
    FROM reminders r
    WHERE r.account_id = accounts.id
    ORDER BY r.remind_at DESC, r.id DESC
    LIMIT 1
), 0);

DELETE FROM reminders
WHERE id NOT IN (
    SELECT MAX(id)
    FROM reminders
    GROUP BY account_id
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reminders_account_unique ON reminders(account_id);
