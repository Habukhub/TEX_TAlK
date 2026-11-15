package domain

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleType string

const (
	ScheduleTypeSpeaker ScheduleType = "SPEAKER"
	ScheduleTypeBreak   ScheduleType = "BREAK"
)

type Schedule struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SubmissionID *uuid.UUID   `gorm:"type:uuid;index"`
	ScheduleType ScheduleType `gorm:"type:varchar(20);default:'SPEAKER'"`
	StartTime    string       `gorm:"type:varchar(5);not null"` // format HH:mm
	EndTime      string       `gorm:"type:varchar(5);not null"` // format HH:mm
	CreatedAt    time.Time    `gorm:"autoCreateTime"`

	Submission *Submission `gorm:"foreignKey:SubmissionID;constraint:OnDelete:SET NULL;"`
}
