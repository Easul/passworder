package model

import "time"

type Auth struct {
	ID           int64  `db:"id" json:"id"`
	PasswordHash string `db:"password_hash" json:"-"`
	KDFSalt      []byte `db:"kdf_salt" json:"-"`
	CreatedAt    int64  `db:"created_at" json:"createdAt"`
	UpdatedAt    int64  `db:"updated_at" json:"updatedAt"`
}

type Category struct {
	ID        int64  `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	Icon      string `db:"icon" json:"icon"`
	SortOrder int    `db:"sort_order" json:"sortOrder"`
	CreatedAt int64  `db:"created_at" json:"createdAt"`
	UpdatedAt int64  `db:"updated_at" json:"updatedAt"`
	IsDeleted int    `db:"is_deleted" json:"isDeleted"`
}

type Account struct {
	ID                  int64  `db:"id" json:"id"`
	CategoryID          int64  `db:"category_id" json:"categoryId"`
	Title               string `db:"title" json:"title"`
	Website             string `db:"website" json:"website"`
	Username            string `db:"username" json:"username"`
	PasswordEncrypted   []byte `db:"password_encrypted" json:"-"`
	Password            string `db:"-" json:"password,omitempty"`
	Email               string `db:"email" json:"email"`
	ReminderEmail       string `db:"reminder_email" json:"reminderEmail"`
	RemindAt            int64  `db:"remind_at" json:"remindAt"`
	ReminderPeriodType  string `db:"reminder_period_type" json:"reminderPeriodType,omitempty"`
	ReminderPeriodValue int    `db:"reminder_period_value" json:"reminderPeriodValue,omitempty"`
	RegistrationTime    int64  `db:"registration_time" json:"registrationTime"`
	RegistrationNotes   string `db:"registration_notes" json:"registrationNotes"`
	Phone               string `db:"phone" json:"phone"`
	Notes               string `db:"notes" json:"notes"`
	Tags                string `db:"tags" json:"tags"`
	IsFavorite          int    `db:"is_favorite" json:"isFavorite"`
	Status              string `db:"status" json:"status"`
	CreatedAt           int64  `db:"created_at" json:"createdAt"`
	UpdatedAt           int64  `db:"updated_at" json:"updatedAt"`
	IsDeleted           int    `db:"is_deleted" json:"isDeleted"`
	ReminderStatus      string `db:"reminder_status" json:"reminderStatus"`
}

type Attachment struct {
	ID           int64  `db:"id" json:"id"`
	AccountID    int64  `db:"account_id" json:"accountId"`
	OriginalName string `db:"original_name" json:"originalName"`
	StoredName   string `db:"stored_name" json:"storedName"`
	MimeType     string `db:"mime_type" json:"mimeType"`
	SizeBytes    int64  `db:"size_bytes" json:"sizeBytes"`
	Sha256       string `db:"sha256" json:"sha256"`
	CreatedAt    int64  `db:"created_at" json:"createdAt"`
	IsDeleted    int    `db:"is_deleted" json:"isDeleted"`
}

type PersonalFile struct {
	ID              int64  `db:"id" json:"id"`
	Title           string `db:"title" json:"title"`
	Remarks         string `db:"remarks" json:"remarks"`
	Body            string `db:"body" json:"body"`
	BodyFormat      string `db:"body_format" json:"bodyFormat"`
	OriginalName    string `db:"original_name" json:"originalName"`
	StoredName      string `db:"stored_name" json:"storedName"`
	MimeType        string `db:"mime_type" json:"mimeType"`
	SizeBytes       int64  `db:"size_bytes" json:"sizeBytes"`
	Sha256          string `db:"sha256" json:"sha256"`
	FileType        string `db:"file_type" json:"fileType"`
	CreatedAt       int64  `db:"created_at" json:"createdAt"`
	UpdatedAt       int64  `db:"updated_at" json:"updatedAt"`
	DeletedAt       *int64 `db:"deleted_at" json:"deletedAt,omitempty"`
	IsDeleted       int    `db:"is_deleted" json:"isDeleted"`
	AttachmentCount int    `db:"attachment_count" json:"attachmentCount"`
}

type NoteAttachment struct {
	ID             int64  `db:"id" json:"id"`
	PersonalFileID int64  `db:"personal_file_id" json:"personalFileId"`
	OriginalName   string `db:"original_name" json:"originalName"`
	StoredName     string `db:"stored_name" json:"storedName"`
	MimeType       string `db:"mime_type" json:"mimeType"`
	SizeBytes      int64  `db:"size_bytes" json:"sizeBytes"`
	Sha256         string `db:"sha256" json:"sha256"`
	FileType       string `db:"file_type" json:"fileType"`
	CreatedAt      int64  `db:"created_at" json:"createdAt"`
	IsDeleted      int    `db:"is_deleted" json:"isDeleted"`
}

type Setting struct {
	Key       string `db:"key" json:"key"`
	Value     string `db:"value" json:"value"`
	UpdatedAt int64  `db:"updated_at" json:"updatedAt"`
}

type Reminder struct {
	ID           int64  `db:"id" json:"id"`
	AccountID    int64  `db:"account_id" json:"accountId"`
	Title        string `db:"title" json:"title"`
	RemindAt     int64  `db:"remind_at" json:"remindAt"`
	NextRemindAt int64  `db:"next_remind_at" json:"nextRemindAt"`
	Email        string `db:"email" json:"email"`
	IsSent       int    `db:"is_sent" json:"isSent"`
	PeriodType   string `db:"period_type" json:"periodType"`
	PeriodValue  int    `db:"period_value" json:"periodValue"`
	PeriodDesc   string `db:"period_desc" json:"periodDesc"`
	CreatedAt    int64  `db:"created_at" json:"createdAt"`
}

type MailSenderSettings struct {
	SMTPHost        string `json:"smtpHost"`
	SMTPPort        int    `json:"smtpPort"`
	SMTPUsername    string `json:"smtpUsername"`
	SMTPPassword    string `json:"smtpPassword"`
	SMTPFromAddress string `json:"fromAddress"`
	SMTPFromName    string `json:"fromName"`
}

type ServerConfig struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	DBPath                string `json:"dbPath"`
	StorageDir            string `json:"storageDir"`
	ReminderCheckInterval int    `json:"reminderCheckInterval"`
}

type DueReminderAccount struct {
	ReminderID    int64  `db:"reminder_id" json:"reminderId"`
	AccountID     int64  `db:"account_id" json:"accountId"`
	Title         string `db:"title" json:"title"`
	Website       string `db:"website" json:"website"`
	Username      string `db:"username" json:"username"`
	ReminderTitle string `db:"reminder_title" json:"reminderTitle"`
	ReminderEmail string `db:"reminder_email" json:"reminderEmail"`
	RemindAt      int64  `db:"remind_at" json:"remindAt"`
}

type Response struct {
	Type    int         `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	ResponseSuccess = 0
	ResponseFail    = 2
	ResponseNoData  = 3
	ResponseNoAuth  = 4
	ResponseError   = 5
)

func Now() int64 {
	return time.Now().Unix()
}
