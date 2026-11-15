package repository_test

import (
	"context"
	"errors"
	"testing"

	"tex/src/internal/domain"
	"tex/src/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockScheduleRepo_Create(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)

	schedule := &domain.Schedule{ScheduleType: domain.ScheduleTypeSpeaker, StartTime: "09:00", EndTime: "10:00"}
	mockRepo.On("Create", mock.Anything, schedule).Return(nil)

	err := mockRepo.Create(context.Background(), schedule)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockScheduleRepo_FindAll(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)

	expected := []domain.Schedule{
		{ID: uuid.New(), ScheduleType: domain.ScheduleTypeSpeaker, StartTime: "09:00", EndTime: "10:00"},
		{ID: uuid.New(), ScheduleType: domain.ScheduleTypeBreak, StartTime: "10:00", EndTime: "10:15"},
	}
	mockRepo.On("FindAll", mock.Anything).Return(expected, nil)

	results, err := mockRepo.FindAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockScheduleRepo_Delete_Success(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	err := mockRepo.Delete(context.Background(), id)

	assert.NoError(t, err)
}

func TestMockScheduleRepo_Delete_NotFound(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(errors.New("schedule slot not found"))

	err := mockRepo.Delete(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "schedule slot not found")
}

func TestMockScheduleRepo_CountSpeakers(t *testing.T) {
	mockRepo := new(testutils.MockScheduleRepository)

	mockRepo.On("CountSpeakers", mock.Anything).Return(int64(3), nil)

	count, err := mockRepo.CountSpeakers(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}
