package repository

import (
	"passworder/internal/model"

	"github.com/jmoiron/sqlx"
)

const (
	reminderCreateSQL = `INSERT INTO reminders
	(account_id, title, remind_at, next_remind_at, email, period_type, period_value, period_desc)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	reminderUpsertByAccountSQL = `INSERT INTO reminders (account_id, title, remind_at, next_remind_at, email, is_sent, period_type, period_value, period_desc)
	VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?)
	ON CONFLICT(account_id) DO UPDATE SET
	title = excluded.title,
	remind_at = excluded.remind_at,
	next_remind_at = excluded.next_remind_at,
	email = excluded.email,
	is_sent = 0,
	period_type = excluded.period_type,
	period_value = excluded.period_value,
	period_desc = excluded.period_desc`
	reminderUpdateSQL          = `UPDATE reminders SET is_sent = 1 WHERE id = ?`
	reminderDeleteSQL          = `DELETE FROM reminders WHERE id = ?`
	reminderDeleteByAccountSQL = `DELETE FROM reminders WHERE account_id = ?`
	reminderGetSQL             = `SELECT id, account_id, title, remind_at, next_remind_at, email, is_sent, period_type, period_value, period_desc, created_at FROM reminders WHERE id = ?`
	reminderListSQL            = `SELECT id, account_id, title, remind_at, next_remind_at, email, is_sent, period_type, period_value, period_desc, created_at FROM reminders ORDER BY remind_at`
	reminderPendingSQL         = `SELECT id, account_id, title, remind_at, next_remind_at, email, is_sent, period_type, period_value, period_desc, created_at FROM reminders WHERE is_sent = 0 AND remind_at <= ? ORDER BY remind_at`
	reminderByAccountSQL       = `SELECT id, account_id, title, remind_at, next_remind_at, email, is_sent, period_type, period_value, period_desc, created_at FROM reminders WHERE account_id = ? ORDER BY remind_at`
	reminderDueAccountsSQL     = `SELECT r.id AS reminder_id, a.id AS account_id, a.title, a.website, a.username,
	r.title AS reminder_title, COALESCE(NULLIF(r.email, ''), NULLIF(a.reminder_email, ''), NULLIF(a.email, ''), '') AS reminder_email,
	CASE
		WHEN COALESCE(r.period_type, '') <> '' AND r.is_sent = 1 THEN r.next_remind_at
		ELSE r.remind_at
	END AS remind_at
	FROM reminders r
	JOIN accounts a ON a.id = r.account_id
	WHERE a.is_deleted = 0 AND (a.status = 'active' OR a.status IS NULL) AND (
		(COALESCE(r.period_type, '') = '' AND r.is_sent = 0 AND r.remind_at <= ?)
		OR (COALESCE(r.period_type, '') <> '' AND ((r.is_sent = 0 AND r.remind_at <= ?) OR (r.is_sent = 1 AND r.next_remind_at <= ?)))
	)
	ORDER BY remind_at, a.title`
	reminderUpdateNextSentSQL = `UPDATE reminders SET is_sent = 1, remind_at = ?, next_remind_at = ?, period_type = ?, period_value = ?, period_desc = ? WHERE id = ?`
)

type ReminderRepository struct {
	db *sqlx.DB
}

func NewReminderRepository(db *sqlx.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

func (r *ReminderRepository) Create(rem *model.Reminder) error {
	result, err := r.db.Exec(reminderCreateSQL, rem.AccountID, rem.Title, rem.RemindAt, rem.NextRemindAt, rem.Email, rem.PeriodType, rem.PeriodValue, rem.PeriodDesc)
	if err != nil {
		return err
	}
	rem.ID, _ = result.LastInsertId()
	rem.CreatedAt = model.Now()
	return nil
}

func (r *ReminderRepository) UpsertByAccount(rem *model.Reminder) error {
	_, err := r.db.Exec(reminderUpsertByAccountSQL, rem.AccountID, rem.Title, rem.RemindAt, rem.NextRemindAt, rem.Email, rem.PeriodType, rem.PeriodValue, rem.PeriodDesc)
	return err
}

func (r *ReminderRepository) MarkAsSent(id int64) error {
	_, err := r.db.Exec(reminderUpdateSQL, id)
	return err
}

func (r *ReminderRepository) Delete(id int64) error {
	_, err := r.db.Exec(reminderDeleteSQL, id)
	return err
}

func (r *ReminderRepository) DeleteByAccount(accountID int64) error {
	_, err := r.db.Exec(reminderDeleteByAccountSQL, accountID)
	return err
}

func (r *ReminderRepository) Get(id int64) (*model.Reminder, error) {
	var rem model.Reminder
	err := r.db.Get(&rem, reminderGetSQL, id)
	if err != nil {
		return nil, err
	}
	return &rem, nil
}

func (r *ReminderRepository) List() ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := r.db.Select(&reminders, reminderListSQL)
	return reminders, err
}

func (r *ReminderRepository) GetPending(now int64) ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := r.db.Select(&reminders, reminderPendingSQL, now)
	return reminders, err
}

func (r *ReminderRepository) ListByAccount(accountID int64) ([]model.Reminder, error) {
	var reminders []model.Reminder
	err := r.db.Select(&reminders, reminderByAccountSQL, accountID)
	return reminders, err
}

func (r *ReminderRepository) ListDueAccounts(now int64) ([]model.DueReminderAccount, error) {
	var reminders []model.DueReminderAccount
	err := r.db.Select(&reminders, reminderDueAccountsSQL, now, now, now)
	return reminders, err
}

func (r *ReminderRepository) UpdateAfterSent(id int64, remindAt int64, nextRemindAt int64, periodType string, periodValue int, periodDesc string) error {
	_, err := r.db.Exec(reminderUpdateNextSentSQL, remindAt, nextRemindAt, periodType, periodValue, periodDesc, id)
	return err
}
