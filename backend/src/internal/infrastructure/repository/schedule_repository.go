package repository

import (
	"context"
	"errors"

	"tex/src/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrSpeakerLimitReached = errors.New("speaker limit reached")

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.Schedule) error
	CreateWithSpeakerLimit(ctx context.Context, schedule *domain.Schedule, max int64) error
	FindAll(ctx context.Context) ([]domain.Schedule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountSpeakers(ctx context.Context) (int64, error)
}

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	return r.db.WithContext(ctx).Create(schedule).Error
}

func (r *scheduleRepository) CreateWithSpeakerLimit(ctx context.Context, schedule *domain.Schedule, max int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&domain.Schedule{}).
			Where("schedule_type = ?", domain.ScheduleTypeSpeaker).
			Count(&count).Error; err != nil {
			return err
		}
		if count >= max {
			return ErrSpeakerLimitReached
		}
		return tx.Create(schedule).Error
	})
}

func (r *scheduleRepository) FindAll(ctx context.Context) ([]domain.Schedule, error) {
	var schedules []domain.Schedule
	// Preload Submission data AND Order by start_time ASC so the timeline is sorted
	err := r.db.WithContext(ctx).
		Preload("Submission").
		Order("start_time asc").
		Find(&schedules).Error
	return schedules, err
}

func (r *scheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Schedule{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("schedule slot not found")
	}
	return nil
}

func (r *scheduleRepository) CountSpeakers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Schedule{}).
		Where("schedule_type = ?", domain.ScheduleTypeSpeaker).
		Count(&count).Error
	return count, err
}
