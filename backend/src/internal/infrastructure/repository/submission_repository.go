package repository

import (
	"context"
	"time"

	"tex/src/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubmissionRepository interface {
	Create(ctx context.Context, submission *domain.Submission) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Submission, error)
	GetPending(ctx context.Context) ([]domain.Submission, error)
	GetApproved(ctx context.Context) ([]domain.Submission, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SubmissionStatus) error
	UpdateSlide(ctx context.Context, id uuid.UUID, slideURL string, slideFileName string) error
	UpdateInfo(ctx context.Context, id uuid.UUID, fullName string, topicTitle string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type submissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) SubmissionRepository {
	return &submissionRepository{db: db}
}

func (r *submissionRepository) Create(ctx context.Context, submission *domain.Submission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

func (r *submissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	var submission domain.Submission
	if err := r.db.WithContext(ctx).First(&submission, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *submissionRepository) GetPending(ctx context.Context) ([]domain.Submission, error) {
	var submissions []domain.Submission
	err := r.db.WithContext(ctx).Where("status = ?", domain.StatusPending).Order("created_at desc").Find(&submissions).Error
	return submissions, err
}

func (r *submissionRepository) GetApproved(ctx context.Context) ([]domain.Submission, error) {
	var submissions []domain.Submission
	err := r.db.WithContext(ctx).Where("status = ?", domain.StatusApproved).Order("created_at desc").Find(&submissions).Error
	return submissions, err
}

func (r *submissionRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Submission, error) {
	var submissions []domain.Submission
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&submissions).Error
	return submissions, err
}

func (r *submissionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SubmissionStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == domain.StatusApproved {
		now := time.Now()
		updates["approved_at"] = now
	}
	return r.db.WithContext(ctx).Model(&domain.Submission{}).Where("id = ?", id).Updates(updates).Error
}

func (r *submissionRepository) UpdateSlide(ctx context.Context, id uuid.UUID, slideURL string, slideFileName string) error {
	return r.db.WithContext(ctx).Model(&domain.Submission{}).Where("id = ?", id).Updates(map[string]interface{}{
		"slide_url":       slideURL,
		"slide_file_name": slideFileName,
	}).Error
}

func (r *submissionRepository) UpdateInfo(ctx context.Context, id uuid.UUID, fullName string, topicTitle string) error {
	return r.db.WithContext(ctx).Model(&domain.Submission{}).Where("id = ?", id).Updates(map[string]interface{}{
		"full_name":   fullName,
		"topic_title": topicTitle,
	}).Error
}

func (r *submissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Submission{}).Error
}
