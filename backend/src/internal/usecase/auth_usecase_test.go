package usecase_test

import (
	"context"
	"errors"
	"testing"

	"tex/src/internal/domain"
	"tex/src/internal/testutils"
	"tex/src/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleGoogleCallback_Success(t *testing.T) {
	mockRepo := new(testutils.MockUserRepository)
	jwtSecret := "test-secret"
	uc := usecase.NewAuthUseCase(mockRepo, jwtSecret)

	googleID := "google-123"
	email := "admin@example.com"
	name := "Admin User"

	expectedUser := &domain.User{
		ID:       uuid.New(),
		GoogleID: googleID,
		Email:    email,
		Name:     name,
	}

	// จำลองว่าหา Email เจอใน DB
	mockRepo.On("VerifyAndSyncUser", mock.Anything, googleID, email, name).Return(expectedUser, nil)

	user, token, err := uc.HandleGoogleCallback(context.Background(), googleID, email, name)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.NotEmpty(t, token) // ต้องได้ JWT Token กลับมา
	mockRepo.AssertExpectations(t)
}

func TestHandleGoogleCallback_UnauthorizedEmail(t *testing.T) {
	mockRepo := new(testutils.MockUserRepository)
	uc := usecase.NewAuthUseCase(mockRepo, "test-secret")

	googleID := "google-456"
	email := "unauthorized@example.com"
	name := "Unknown User"

	// จำลองว่าหา Email ไม่เจอใน DB
	mockRepo.On("VerifyAndSyncUser", mock.Anything, googleID, email, name).
		Return(nil, errors.New("record not found"))

	user, token, err := uc.HandleGoogleCallback(context.Background(), googleID, email, name)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	assert.Nil(t, user)
	assert.Empty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestHandleGoogleCallback_RepoError(t *testing.T) {
	mockRepo := new(testutils.MockUserRepository)
	uc := usecase.NewAuthUseCase(mockRepo, "test-secret")

	mockRepo.On("VerifyAndSyncUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db connection failed"))

	user, token, err := uc.HandleGoogleCallback(context.Background(), "id", "email", "name")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Empty(t, token)
}
