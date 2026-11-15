package usecase

import (
	"context"

	"tex/src/internal/domain"
	"tex/src/internal/dto"
	"tex/src/internal/infrastructure/repository"

	"github.com/google/uuid"
)

// ErrSpeakerLimitReached is re-exported from repository for handler-layer error detection.
var ErrSpeakerLimitReached = repository.ErrSpeakerLimitReached

type ScheduleUseCase interface {
	GetTimeline(ctx context.Context) ([]domain.Schedule, error)
	AddSlot(ctx context.Context, req dto.ScheduleSlotRequest) error
	DeleteSlot(ctx context.Context, id uuid.UUID) error
}

type scheduleUseCase struct {
	scheduleRepo repository.ScheduleRepository
}

func NewScheduleUseCase(scheduleRepo repository.ScheduleRepository) ScheduleUseCase {
	return &scheduleUseCase{scheduleRepo: scheduleRepo}
}

func (u *scheduleUseCase) GetTimeline(ctx context.Context) ([]domain.Schedule, error) {
	return u.scheduleRepo.FindAll(ctx)
}

func (u *scheduleUseCase) AddSlot(ctx context.Context, req dto.ScheduleSlotRequest) error {
	schedule := &domain.Schedule{
		SubmissionID: req.SubmissionID,
		ScheduleType: domain.ScheduleType(req.ScheduleType),
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	}
	if req.ScheduleType == string(domain.ScheduleTypeSpeaker) {
		return u.scheduleRepo.CreateWithSpeakerLimit(ctx, schedule, 40)
	}
	return u.scheduleRepo.Create(ctx, schedule)
}

func (u *scheduleUseCase) DeleteSlot(ctx context.Context, id uuid.UUID) error {
	return u.scheduleRepo.Delete(ctx, id)
}
