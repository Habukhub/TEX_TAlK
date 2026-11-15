package dto

import (
	"github.com/google/uuid"
)

type ScheduleSlotRequest struct {
	SubmissionID *uuid.UUID `json:"submission_id"` // ใส่ pointer เผื่อกรณีเป็น nil (เช่น Break)
	ScheduleType string     `json:"schedule_type" validate:"required,oneof=SPEAKER BREAK"`
	StartTime    string     `json:"start_time" validate:"required"` // format HH:mm
	EndTime      string     `json:"end_time" validate:"required"`   // format HH:mm
}
