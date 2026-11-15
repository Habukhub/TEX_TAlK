package domain

import "time"

const (
	SettingKeyPresentationUploadEnabled = "presentation_upload_enabled"
)

type SystemSetting struct {
	Key       string    `gorm:"primaryKey;column:key"`
	Value     string    `gorm:"column:value;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}
