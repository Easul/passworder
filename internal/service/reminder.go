package service

import (
	"fmt"
	"strings"
	"time"

	"passworder/internal/model"
	"passworder/internal/repository"
)

type ReminderService struct {
	repo           *repository.ReminderRepository
	settingService *SettingService
	sender         MailSender
}

func NewReminderService(repo *repository.ReminderRepository, settingService *SettingService, sender MailSender) *ReminderService {
	return &ReminderService{repo: repo, settingService: settingService, sender: sender}
}

func (s *ReminderService) Create(rem *model.Reminder) error {
	return s.repo.Create(rem)
}

func (s *ReminderService) Delete(id int64) error {
	return s.repo.Delete(id)
}

func (s *ReminderService) Get(id int64) (*model.Reminder, error) {
	return s.repo.Get(id)
}

func (s *ReminderService) List() ([]model.Reminder, error) {
	return s.repo.List()
}

func (s *ReminderService) GetPending() ([]model.Reminder, error) {
	return s.repo.GetPending(time.Now().Unix())
}

func (s *ReminderService) MarkAsSent(id int64) error {
	return s.repo.MarkAsSent(id)
}

func (s *ReminderService) ListByAccount(accountID int64) ([]model.Reminder, error) {
	return s.repo.ListByAccount(accountID)
}

func (s *ReminderService) NormalizeSchedule(remindAt int64, periodType string, periodValue int) (int64, int64) {
	if remindAt <= 0 || periodType == "" {
		return remindAt, 0
	}
	period := &ReminderPeriod{Type: periodType, Value: periodValue}
	current := time.Unix(remindAt, 0)
	now := time.Now()
	for !current.After(now) {
		current = period.CalculateNext(current)
	}
	next := period.CalculateNext(current)
	return current.Unix(), next.Unix()
}

func (s *ReminderService) SyncAccountReminder(account *model.Account, periodType string, periodValue int) error {
	if account == nil || account.ID == 0 {
		return nil
	}
	if account.RemindAt <= 0 {
		return s.repo.DeleteByAccount(account.ID)
	}

	normalizedRemindAt, nextRemindAt := s.NormalizeSchedule(account.RemindAt, periodType, periodValue)
	account.RemindAt = normalizedRemindAt
	period := &ReminderPeriod{Type: periodType, Value: periodValue}
	reminder := &model.Reminder{
		AccountID:    account.ID,
		Title:        fmt.Sprintf("登录提醒: %s", account.Title),
		RemindAt:     normalizedRemindAt,
		NextRemindAt: nextRemindAt,
		Email:        strings.TrimSpace(account.ReminderEmail),
		PeriodType:   periodType,
		PeriodValue:  periodValue,
		PeriodDesc:   period.GetDescription(),
	}
	if reminder.Email == "" {
		reminder.Email = strings.TrimSpace(account.Email)
	}
	return s.repo.UpsertByAccount(reminder)
}

func (s *ReminderService) SendDue() (map[string]int, error) {
	settings := s.settingService.GetMailSenderSettings()
	if settings.SMTPHost == "" || settings.SMTPFromAddress == "" {
		return nil, fmt.Errorf("mail sender settings incomplete")
	}

	dueAccounts, err := s.repo.ListDueAccounts(time.Now().Unix())
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]model.DueReminderAccount)
	for _, item := range dueAccounts {
		to := strings.TrimSpace(item.ReminderEmail)
		if to == "" {
			continue
		}
		grouped[to] = append(grouped[to], item)
	}

	sentGroups := 0
	sentReminders := 0
	failedGroups := 0

	for recipient, items := range grouped {
		subject := fmt.Sprintf("%d 个账号登录提醒", len(items))
		body := s.buildReminderBody(items)
		if err := s.sender.Send(settings, recipient, subject, body); err != nil {
			failedGroups++
			continue
		}
		sentGroups++
		for _, item := range items {
			reminder, err := s.repo.Get(item.ReminderID)
			if err != nil {
				continue
			}
			if reminder.PeriodType == "" {
				_ = s.repo.MarkAsSent(item.ReminderID)
			} else {
				nextRemindAt := s.calculateNextRemindAt(item.RemindAt, reminder.PeriodType, reminder.PeriodValue)
				_ = s.repo.UpdateAfterSent(item.ReminderID, item.RemindAt, nextRemindAt, reminder.PeriodType, reminder.PeriodValue, reminder.PeriodDesc)
			}
			sentReminders++
		}
	}

	return map[string]int{
		"dueAccounts":   len(dueAccounts),
		"sentGroups":    sentGroups,
		"failedGroups":  failedGroups,
		"sentReminders": sentReminders,
	}, nil
}

func (s *ReminderService) calculateNextRemindAt(remindAt int64, periodType string, periodValue int) int64 {
	if periodType == "" {
		return 0
	}
	period := &ReminderPeriod{Type: periodType, Value: periodValue}
	return period.CalculateNext(time.Unix(remindAt, 0)).Unix()
}

func (s *ReminderService) buildReminderBody(items []model.DueReminderAccount) string {
	lines := []string{"以下账号需要登录：", ""}
	for _, item := range items {
		line := fmt.Sprintf("- %s | 用户名: %s", item.Title, item.Username)
		if strings.TrimSpace(item.Website) != "" {
			line = fmt.Sprintf("%s | 网站: %s", line, item.Website)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
