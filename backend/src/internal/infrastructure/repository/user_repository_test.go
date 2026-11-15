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

func TestMockUserRepo_VerifyAndSyncUser_Found(t *testing.T) {
	mockRepo := new(testutils.MockUserRepository)

	expected := &domain.User{ID: uuid.New(), GoogleID: "g-123", Email: "admin@e.com", Name: "Admin"}
	mockRepo.On("VerifyAndSyncUser", mock.Anything, "g-123", "admin@e.com", "Admin").Return(expected, nil)

	user, err := mockRepo.VerifyAndSyncUser(context.Background(), "g-123", "admin@e.com", "Admin")

	assert.NoError(t, err)
	assert.Equal(t, "admin@e.com", user.Email)
}

func TestMockUserRepo_VerifyAndSyncUser_NotFound(t *testing.T) {
	mockRepo := new(testutils.MockUserRepository)

	mockRepo.On("VerifyAndSyncUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("record not found"))

	user, err := mockRepo.VerifyAndSyncUser(context.Background(), "id", "email", "name")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestMockUserRepo_FindByGoogleID_Found(t *testing.T) {
	mockRepo := new(testutils.MockUserRepository)

	expected := &domain.User{ID: uuid.New(), GoogleID: "g-456", Email: "found@e.com"}
	mockRepo.On("FindByGoogleID", mock.Anything, "g-456").Return(expected, nil)

	user, err := mockRepo.FindByGoogleID(context.Background(), "g-456")

	assert.NoError(t, err)
	assert.Equal(t, "g-456", user.GoogleID)
}

func TestMockUserRepo_FindByGoogleID_NotFound(t *testing.T) {
	mockRepo := new(testutils.MockUserRepository)

	mockRepo.On("FindByGoogleID", mock.Anything, "not-exist").Return(nil, errors.New("record not found"))

	user, err := mockRepo.FindByGoogleID(context.Background(), "not-exist")

	assert.Error(t, err)
	assert.Nil(t, user)
}
