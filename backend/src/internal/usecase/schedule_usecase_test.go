package usecase_test

import (
	"context"
	"errors"
	"testing"

	"tex/src/internal/domain"
	"tex/src/internal/dto"
	"tex/src/internal/infrastructure/repository"
	"tex/src/internal/testutils"
	"tex/src/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTimeline_Success(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)
	uc := usecase.NewScheduleUseCase(mockRepo)

	expected := []domain.Schedule{
		{ID: uuid.New(), ScheduleType: domain.ScheduleTypeSpeaker, StartTime: "09:00", EndTime: "10:00"},
		{ID: uuid.New(), ScheduleType: domain.ScheduleTypeBreak, StartTime: "10:00", EndTime: "10:15"},
	}
	mockRepo.On("FindAll", mock.Anything).Return(expected, nil)

	result, err := uc.GetTimeline(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetTimeline_RepoError(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)
	uc := usecase.NewScheduleUseCase(mockRepo)

	mockRepo.On("FindAll", mock.Anything).Return([]domain.Schedule{}, errors.New("db error"))

	result, err := uc.GetTimeline(context.Background())

	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestAddSlot_Speaker_Success(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)
	uc := usecase.NewScheduleUseCase(mockRepo)

	submissionID := uuid.New()
	req := dto.ScheduleSlotRequest{
		SubmissionID: &submissionID,
		ScheduleType: "SPEAKER",
		StartTime:    "09:00",
		EndTime:      "10:00",
	}

	mockRepo.On("CreateWithSpeakerLimit", mock.Anything, mock.AnythingOfType("*domain.Schedule"), int64(40)).Return(nil)

	err := uc.AddSlot(context.Background(), req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
	mockRepo.AssertNotCalled(t, "CountSpeakers")
}

func TestAddSlot_Speaker_LimitReached(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)
	uc := usecase.NewScheduleUseCase(mockRepo)

	submissionID := uuid.New()
	req := dto.ScheduleSlotRequest{
		SubmissionID: &submissionID,
		ScheduleType: "SPEAKER",
		StartTime:    "09:00",
		EndTime:      "10:00",
	}

	mockRepo.On("CreateWithSpeakerLimit", mock.Anything, mock.AnythingOfType("*domain.Schedule"), int64(40)).
		Return(repository.ErrSpeakerLimitReached)

	err := uc.AddSlot(context.Background(), req)

	assert.ErrorIs(t, err, repository.ErrSpeakerLimitReached)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestAddSlot_Break_Success(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)
	uc := usecase.NewScheduleUseCase(mockRepo)

	req := dto.ScheduleSlotRequest{
		SubmissionID: nil,
		ScheduleType: "BREAK",
		StartTime:    "10:00",
		EndTime:      "10:15",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Schedule")).Return(nil)

	err := uc.AddSlot(context.Background(), req)

	assert.NoError(t, err)
	mockRepo.AssertNotCalled(t, "CreateWithSpeakerLimit")
	mockRepo.AssertExpectations(t)
}

func TestAddSlot_Speaker_DBError(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)
	uc := usecase.NewScheduleUseCase(mockRepo)

	submissionID := uuid.New()
	req := dto.ScheduleSlotRequest{
		SubmissionID: &submissionID,
		ScheduleType: "SPEAKER",
		StartTime:    "09:00",
		EndTime:      "10:00",
	}

	mockRepo.On("CreateWithSpeakerLimit", mock.Anything, mock.AnythingOfType("*domain.Schedule"), int64(40)).
		Return(errors.New("db error"))

	err := uc.AddSlot(context.Background(), req)

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestDeleteSlot_Success(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)
	uc := usecase.NewScheduleUseCase(mockRepo)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	err := uc.DeleteSlot(context.Background(), id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteSlot_NotFound(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)
	uc := usecase.NewScheduleUseCase(mockRepo)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(errors.New("schedule slot not found"))

	err := uc.DeleteSlot(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "schedule slot not found")
}
