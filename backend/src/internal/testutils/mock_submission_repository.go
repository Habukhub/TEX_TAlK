package testutils

import (
	"context"

	"tex/src/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockSubmissionRepository is a mock for repository.SubmissionRepository
type MockSubmissionRepository struct {
	mock.Mock
}

func (m *MockSubmissionRepository) Create(ctx context.Context, submission *domain.Submission) error {
	args := m.Called(ctx, submission)
	return args.Error(0)
}

func (m *MockSubmissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Submission, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]domain.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) GetPending(ctx context.Context) ([]domain.Submission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) GetApproved(ctx context.Context) ([]domain.Submission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SubmissionStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockSubmissionRepository) UpdateSlide(ctx context.Context, id uuid.UUID, slideURL string, slideFileName string) error {
	args := m.Called(ctx, id, slideURL, slideFileName)
	return args.Error(0)
}

func (m *MockSubmissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubmissionRepository) UpdateInfo(ctx context.Context, id uuid.UUID, fullName string, topicTitle string) error {
	args := m.Called(ctx, id, fullName, topicTitle)
	return args.Error(0)
}
