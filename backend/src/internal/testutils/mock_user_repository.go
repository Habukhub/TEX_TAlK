package testutils

import (
	"context"

	"tex/src/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock for repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) VerifyAndSyncUser(ctx context.Context, googleID, email, name string) (*domain.User, error) {
	args := m.Called(ctx, googleID, email, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	args := m.Called(ctx, googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
