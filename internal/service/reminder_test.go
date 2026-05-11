package service

import (
	"testing"

	"passworder/internal/model"
	"passworder/internal/repository"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type fakeMailSender struct {
	calls []mailCall
}

type mailCall struct {
	to      string
	subject string
	body    string
}

func (f *fakeMailSender) Send(settings model.MailSenderSettings, to string, subject string, body string) error {
	f.calls = append(f.calls, mailCall{to: to, subject: subject, body: body})
	return nil
}

func TestReminderServiceSyncAccountReminder(t *testing.T) {
	db := openReminderTestDB(t)
	settingService := NewSettingService(repository.NewSettingRepository(db))
	service := NewReminderService(repository.NewReminderRepository(db), settingService, &fakeMailSender{})

	account := &model.Account{ID: 1, Title: "GitHub", Email: "owner@example.com", ReminderEmail: "notify@example.com", RemindAt: 1893456000}
	if err := service.SyncAccountReminder(account, "", 0); err != nil {
		t.Fatalf("sync reminder: %v", err)
	}

	items, err := service.ListByAccount(1)
	if err != nil {
		t.Fatalf("list reminder: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 reminder, got %d", len(items))
	}
	if items[0].Email != "notify@example.com" {
		t.Fatalf("unexpected reminder email: %s", items[0].Email)
	}

	account.RemindAt = 0
	if err := service.SyncAccountReminder(account, "", 0); err != nil {
		t.Fatalf("delete reminder: %v", err)
	}
	items, err = service.ListByAccount(1)
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 reminders after delete, got %d", len(items))
	}
}

func TestReminderServiceSendDueGroupsByRecipient(t *testing.T) {
	db := openReminderTestDB(t)
	seedReminderTestData(t, db)
	settingService := NewSettingService(repository.NewSettingRepository(db))
	fakeSender := &fakeMailSender{}
	service := NewReminderService(repository.NewReminderRepository(db), settingService, fakeSender)

	result, err := service.SendDue()
	if err != nil {
		t.Fatalf("send due: %v", err)
	}
	if result["dueAccounts"] != 3 {
		t.Fatalf("expected 3 due accounts, got %d", result["dueAccounts"])
	}
	if result["sentGroups"] != 2 {
		t.Fatalf("expected 2 sent groups, got %d", result["sentGroups"])
	}
	if result["sentReminders"] != 3 {
		t.Fatalf("expected 3 sent reminders, got %d", result["sentReminders"])
	}
	if len(fakeSender.calls) != 2 {
		t.Fatalf("expected 2 send calls, got %d", len(fakeSender.calls))
	}

	items, err := repository.NewReminderRepository(db).List()
	if err != nil {
		t.Fatalf("list reminders: %v", err)
	}
	for _, item := range items {
		if item.IsSent != 1 {
			t.Fatalf("expected reminder %d to be marked sent", item.ID)
		}
	}
}

func TestReminderServiceExcludesInactiveAccounts(t *testing.T) {
	db := openReminderTestDB(t)
	settingService := NewSettingService(repository.NewSettingRepository(db))
	fakeSender := &fakeMailSender{}
	service := NewReminderService(repository.NewReminderRepository(db), settingService, fakeSender)

	statements := []string{
		`INSERT INTO settings(key, value) VALUES ('mail.smtp_host', 'smtp.example.com');`,
		`INSERT INTO settings(key, value) VALUES ('mail.smtp_port', '587');`,
		`INSERT INTO settings(key, value) VALUES ('mail.smtp_username', 'mailer');`,
		`INSERT INTO settings(key, value) VALUES ('mail.smtp_password', 'secret');`,
		`INSERT INTO settings(key, value) VALUES ('mail.from_address', 'noreply@example.com');`,
		`INSERT INTO settings(key, value) VALUES ('mail.from_name', 'Passworder');`,
		`INSERT INTO accounts(id, title, website, username, email, reminder_email, remind_at, status, password_encrypted) VALUES (1, 'Active Account', 'https://active.com', 'user1', 'user1@example.com', 'notify@example.com', 1, 'active', 'x');`,
		`INSERT INTO accounts(id, title, website, username, email, reminder_email, remind_at, status, password_encrypted) VALUES (2, 'Inactive Account', 'https://inactive.com', 'user2', 'user2@example.com', 'notify@example.com', 1, 'inactive', 'x');`,
		`INSERT INTO reminders(id, account_id, title, remind_at, next_remind_at, email, is_sent) VALUES (1, 1, '登录提醒: Active Account', 1, 0, 'notify@example.com', 0);`,
		`INSERT INTO reminders(id, account_id, title, remind_at, next_remind_at, email, is_sent) VALUES (2, 2, '登录提醒: Inactive Account', 1, 0, 'notify@example.com', 0);`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed data: %v", err)
		}
	}

	result, err := service.SendDue()
	if err != nil {
		t.Fatalf("send due: %v", err)
	}
	if result["dueAccounts"] != 1 {
		t.Fatalf("expected 1 due account (only active), got %d", result["dueAccounts"])
	}
	if result["sentReminders"] != 1 {
		t.Fatalf("expected 1 sent reminder (only active), got %d", result["sentReminders"])
	}
	if len(fakeSender.calls) != 1 {
		t.Fatalf("expected 1 send call, got %d", len(fakeSender.calls))
	}
}

func openReminderTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlx.Connect("sqlite3", ":memory:?_fk=1")
	if err != nil {
		t.Fatalf("connect sqlite: %v", err)
	}
	schema := []string{
		`CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL, updated_at INTEGER DEFAULT 0);`,
		`CREATE TABLE accounts (id INTEGER PRIMARY KEY AUTOINCREMENT, category_id INTEGER NOT NULL DEFAULT 1, title TEXT NOT NULL, website TEXT, username TEXT NOT NULL, password_encrypted BLOB NOT NULL DEFAULT '', email TEXT, reminder_email TEXT NOT NULL DEFAULT '', remind_at INTEGER NOT NULL DEFAULT 0, phone TEXT, notes TEXT, tags TEXT DEFAULT '[]', is_favorite INTEGER DEFAULT 0, status TEXT NOT NULL DEFAULT 'active', created_at INTEGER DEFAULT 0, updated_at INTEGER DEFAULT 0, is_deleted INTEGER DEFAULT 0);`,
		`CREATE TABLE reminders (id INTEGER PRIMARY KEY AUTOINCREMENT, account_id INTEGER NOT NULL, title TEXT NOT NULL, remind_at INTEGER NOT NULL, next_remind_at INTEGER NOT NULL DEFAULT 0, email TEXT, is_sent INTEGER DEFAULT 0, period_type TEXT NOT NULL DEFAULT '', period_value INTEGER NOT NULL DEFAULT 0, period_desc TEXT NOT NULL DEFAULT '', created_at INTEGER DEFAULT 0);`,
		`CREATE UNIQUE INDEX idx_reminders_account_unique ON reminders(account_id);`,
	}
	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("apply schema: %v", err)
		}
	}
	return db
}

func seedReminderTestData(t *testing.T, db *sqlx.DB) {
	t.Helper()
	statements := []string{
		`INSERT INTO settings(key, value) VALUES ('mail.smtp_host', 'smtp.example.com');`,
		`INSERT INTO settings(key, value) VALUES ('mail.smtp_port', '587');`,
		`INSERT INTO settings(key, value) VALUES ('mail.smtp_username', 'mailer');`,
		`INSERT INTO settings(key, value) VALUES ('mail.smtp_password', 'secret');`,
		`INSERT INTO settings(key, value) VALUES ('mail.from_address', 'noreply@example.com');`,
		`INSERT INTO settings(key, value) VALUES ('mail.from_name', 'Passworder');`,
		`INSERT INTO accounts(id, title, website, username, email, reminder_email, remind_at, password_encrypted) VALUES (1, 'GitHub', 'https://github.com', 'alice', 'alice@example.com', 'notify@example.com', 1, 'x');`,
		`INSERT INTO accounts(id, title, website, username, email, reminder_email, remind_at, password_encrypted) VALUES (2, 'GitLab', 'https://gitlab.com', 'bob', 'bob@example.com', 'notify@example.com', 1, 'x');`,
		`INSERT INTO accounts(id, title, website, username, email, reminder_email, remind_at, password_encrypted) VALUES (3, 'Jira', 'https://jira.example.com', 'carol', 'carol@example.com', 'ops@example.com', 1, 'x');`,
		`INSERT INTO reminders(id, account_id, title, remind_at, next_remind_at, email, is_sent) VALUES (1, 1, '登录提醒: GitHub', 1, 0, 'notify@example.com', 0);`,
		`INSERT INTO reminders(id, account_id, title, remind_at, next_remind_at, email, is_sent) VALUES (2, 2, '登录提醒: GitLab', 1, 0, 'notify@example.com', 0);`,
		`INSERT INTO reminders(id, account_id, title, remind_at, next_remind_at, email, is_sent) VALUES (3, 3, '登录提醒: Jira', 1, 0, 'ops@example.com', 0);`,
	}
	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("seed data: %v", err)
		}
	}
}
