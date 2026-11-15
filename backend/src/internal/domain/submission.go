package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubmissionStatus string

const (
	StatusPending  SubmissionStatus = "PENDING"
	StatusApproved SubmissionStatus = "APPROVED"
	StatusRejected SubmissionStatus = "REJECTED"
)

type Submission struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FullName      string           `gorm:"type:varchar(255);not null"`
	Email         string           `gorm:"type:varchar(255);not null"`
	TopicTitle    string           `gorm:"type:varchar(255);not null"`
	Status        SubmissionStatus `gorm:"type:varchar(20);default:'PENDING'"`
	ApprovedAt    *time.Time       `gorm:"default:null"`
	SlideURL      *string          `gorm:"type:text;default:null"`
	SlideFileName *string          `gorm:"type:varchar(255);default:null"`
	CreatedAt     time.Time        `gorm:"autoCreateTime"`

	Schedules []Schedule `gorm:"foreignKey:SubmissionID" json:"Schedules,omitempty"`
}
