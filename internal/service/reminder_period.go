package service

import (
	"fmt"
	"time"
)

type ReminderPeriod struct {
	Type  string `json:"type"`  // yearly, monthly, weekly, daily, hourly, days
	Value int    `json:"value"` // 用于间隔（如每3天，每周几等）
	Desc  string `json:"desc"`  // 描述文本
}

func (rp *ReminderPeriod) CalculateNext(fromTime time.Time) time.Time {
	switch rp.Type {
	case "yearly":
		return fromTime.AddDate(1, 0, 0)
	case "monthly":
		if rp.Value <= 0 {
			rp.Value = 1
		}
		return fromTime.AddDate(0, rp.Value, 0)
	case "weekly":
		return fromTime.AddDate(0, 0, 7)
	case "daily":
		return fromTime.AddDate(0, 0, 1)
	case "hourly":
		if rp.Value <= 0 {
			rp.Value = 1
		}
		return fromTime.Add(time.Duration(rp.Value) * time.Hour)
	case "days":
		if rp.Value <= 0 {
			rp.Value = 1
		}
		return fromTime.AddDate(0, 0, rp.Value)
	default:
		return fromTime.AddDate(0, 1, 0)
	}
}

func (rp *ReminderPeriod) GetDescription() string {
	switch rp.Type {
	case "yearly":
		return "每年"
	case "monthly":
		if rp.Value > 1 {
			return fmt.Sprintf("每%d个月", rp.Value)
		}
		return "每月"
	case "weekly":
		return "每周"
	case "daily":
		return "每天"
	case "hourly":
		if rp.Value <= 0 {
			rp.Value = 1
		}
		return fmt.Sprintf("每%d小时", rp.Value)
	case "days":
		if rp.Value <= 0 {
			rp.Value = 1
		}
		return fmt.Sprintf("每%d天", rp.Value)
	default:
		return "不重复"
	}
}

func (rp *ReminderPeriod) IsValid() bool {
	validTypes := []string{"yearly", "monthly", "weekly", "daily", "hourly", "days", ""}
	for _, t := range validTypes {
		if rp.Type == t {
			return true
		}
	}
	return false
}
