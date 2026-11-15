package dto

import (
	"errors"
	"net/mail"
	"strings"
)

type SubmitRequest struct {
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	TopicTitle string `json:"topic_title"`
}

func (r SubmitRequest) Validate() error {
	if strings.TrimSpace(r.FullName) == "" {
		return errors.New("full_name is required")
	}
	if len(r.FullName) > 255 {
		return errors.New("full_name must not exceed 255 characters")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("email is invalid")
	}
	if strings.TrimSpace(r.TopicTitle) == "" {
		return errors.New("topic_title is required")
	}
	if len(r.TopicTitle) > 255 {
		return errors.New("topic_title must not exceed 255 characters")
	}
	return nil
}

type PublicSubmissionResponse struct {
	ID            string  `json:"id"`
	FullName      string  `json:"full_name"`
	TopicTitle    string  `json:"topic_title"`
	Status        string  `json:"status"`
	SlideFileName *string `json:"slide_file_name,omitempty"`
}

// UpdateSubmissionInfoRequest allows admin to edit applicant name and topic title.
type UpdateSubmissionInfoRequest struct {
	FullName   string `json:"full_name"`
	TopicTitle string `json:"topic_title"`
}

func (r UpdateSubmissionInfoRequest) Validate() error {
	if strings.TrimSpace(r.FullName) == "" {
		return errors.New("full_name is required")
	}
	if len(r.FullName) > 255 {
		return errors.New("full_name must not exceed 255 characters")
	}
	if strings.TrimSpace(r.TopicTitle) == "" {
		return errors.New("topic_title is required")
	}
	if len(r.TopicTitle) > 255 {
		return errors.New("topic_title must not exceed 255 characters")
	}
	return nil
}

// DownloadZipRequest lists the submission IDs the admin wants to ZIP.
type DownloadZipRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}
