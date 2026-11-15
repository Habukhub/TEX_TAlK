package testutils

import (
	"context"

	"tex/src/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockScheduleRepository is a mock for repository.ScheduleRepository
type MockScheduleRepository struct {
	mock.Mock
}

func (m *MockScheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockScheduleRepository) CreateWithSpeakerLimit(ctx context.Context, schedule *domain.Schedule, max int64) error {
	args := m.Called(ctx, schedule, max)
	return args.Error(0)
}

func (m *MockScheduleRepository) FindAll(ctx context.Context) ([]domain.Schedule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Schedule), args.Error(1)
}

func (m *MockScheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockScheduleRepository) CountSpeakers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
