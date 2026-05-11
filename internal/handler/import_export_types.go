package handler

import "passworder/internal/model"

type ExportData struct {
	Version    string            `json:"version"`
	ExportTime int64             `json:"exportTime"`
	Accounts   []ExportAccount   `json:"accounts"`
	Categories []model.Category  `json:"categories"`
	Notes      []ExportNote      `json:"notes"`
	Settings   map[string]string `json:"settings,omitempty"`
}

type ExportAccount struct {
	ID                  int64  `json:"id"`
	CategoryID          int64  `json:"categoryId"`
	Title               string `json:"title"`
	Website             string `json:"website"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	PasswordEncrypted   string `json:"passwordEncrypted,omitempty"`
	Email               string `json:"email"`
	ReminderEmail       string `json:"reminderEmail"`
	RemindAt            int64  `json:"remindAt"`
	ReminderPeriodType  string `json:"reminderPeriodType"`
	ReminderPeriodValue int    `json:"reminderPeriodValue"`
	RegistrationTime    int64  `json:"registrationTime"`
	RegistrationNotes   string `json:"registrationNotes"`
	Phone               string `json:"phone"`
	Notes               string `json:"notes"`
	Tags                string `json:"tags"`
	IsFavorite          int    `json:"isFavorite"`
	Status              string `json:"status"`
	CreatedAt           int64  `json:"createdAt"`
	UpdatedAt           int64  `json:"updatedAt"`
}

type ExportNote struct {
	ID           int64              `json:"id"`
	Title        string             `json:"title"`
	Remarks      string             `json:"remarks"`
	Body         string             `json:"body"`
	BodyFormat   string             `json:"bodyFormat"`
	OriginalName string             `json:"originalName"`
	MimeType     string             `json:"mimeType"`
	SizeBytes    int64              `json:"sizeBytes"`
	FileType     string             `json:"fileType"`
	Sha256       string             `json:"sha256"`
	Attachments  []ExportAttachment `json:"attachments"`
	CreatedAt    int64              `json:"createdAt"`
	UpdatedAt    int64              `json:"updatedAt"`
}

type ExportAttachment struct {
	ID           int64  `json:"id"`
	OriginalName string `json:"originalName"`
	MimeType     string `json:"mimeType"`
	SizeBytes    int64  `json:"sizeBytes"`
	FileType     string `json:"fileType"`
	Sha256       string `json:"sha256"`
}
