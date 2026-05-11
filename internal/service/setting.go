package service

import (
	"fmt"

	"passworder/internal/model"
	"passworder/internal/repository"
)

type SettingService struct {
	repo *repository.SettingRepository
}

func NewSettingService(repo *repository.SettingRepository) *SettingService {
	return &SettingService{repo: repo}
}

func (s *SettingService) Get(key string) (string, error) {
	setting, err := s.repo.Get(key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (s *SettingService) Set(key, value string) error {
	return s.repo.Set(key, value)
}

func (s *SettingService) GetBool(key string, defaultValue bool) bool {
	val, err := s.Get(key)
	if err != nil {
		return defaultValue
	}
	return val == "true" || val == "1"
}

func (s *SettingService) GetInt(key string, defaultValue int) int {
	val, err := s.Get(key)
	if err != nil {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(val, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}

func (s *SettingService) List() ([]model.Setting, error) {
	return s.repo.List()
}

func (s *SettingService) GetMailSenderSettings() model.MailSenderSettings {
	return model.MailSenderSettings{
		SMTPHost:        s.getString("mail.smtp_host"),
		SMTPPort:        s.GetInt("mail.smtp_port", 587),
		SMTPUsername:    s.getString("mail.smtp_username"),
		SMTPPassword:    s.getString("mail.smtp_password"),
		SMTPFromAddress: s.getString("mail.from_address"),
		SMTPFromName:    s.getString("mail.from_name"),
	}
}

func (s *SettingService) GetServerConfig() model.ServerConfig {
	return model.ServerConfig{
		Host:                  s.getString("server.host"),
		Port:                  s.GetInt("server.port", 0),
		DBPath:                s.getString("server.db_path"),
		StorageDir:            s.getString("server.storage_dir"),
		ReminderCheckInterval: s.GetInt("server.reminder_interval", 0),
	}
}

func (s *SettingService) SetServerConfig(cfg model.ServerConfig) error {
	if cfg.Host != "" {
		if err := s.Set("server.host", cfg.Host); err != nil {
			return err
		}
	}
	if cfg.Port > 0 {
		if err := s.Set("server.port", fmt.Sprintf("%d", cfg.Port)); err != nil {
			return err
		}
	}
	if cfg.DBPath != "" {
		if err := s.Set("server.db_path", cfg.DBPath); err != nil {
			return err
		}
	}
	if cfg.StorageDir != "" {
		if err := s.Set("server.storage_dir", cfg.StorageDir); err != nil {
			return err
		}
	}
	if cfg.ReminderCheckInterval > 0 {
		if err := s.Set("server.reminder_interval", fmt.Sprintf("%d", cfg.ReminderCheckInterval)); err != nil {
			return err
		}
	}
	return nil
}

func (s *SettingService) getString(key string) string {
	val, err := s.Get(key)
	if err != nil {
		return ""
	}
	return val
}
