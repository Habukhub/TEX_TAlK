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

// ทดสอบผ่าน MockSubmissionRepository ว่า interface ทำงานถูกต้อง
// (Repository layer จริงทดสอบผ่าน Integration Test ที่ต้องการ DB จริง)

func TestMockSubmissionRepo_Create(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)

	submission := &domain.Submission{
		FullName: "Test", Email: "test@e.com", TopicTitle: "Topic", Status: domain.StatusPending,
	}
	mockRepo.On("Create", mock.Anything, submission).Return(nil)

	err := mockRepo.Create(context.Background(), submission)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMockSubmissionRepo_Create_Error(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)

	submission := &domain.Submission{}
	mockRepo.On("Create", mock.Anything, submission).Return(errors.New("unique constraint violation"))

	err := mockRepo.Create(context.Background(), submission)

	assert.Error(t, err)
	assert.EqualError(t, err, "unique constraint violation")
}

func TestMockSubmissionRepo_GetByID_Found(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)

	id := uuid.New()
	expected := &domain.Submission{ID: id, FullName: "Found User", Status: domain.StatusPending}
	mockRepo.On("GetByID", mock.Anything, id).Return(expected, nil)

	result, err := mockRepo.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, id, result.ID)
}

func TestMockSubmissionRepo_GetByID_NotFound(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)

	id := uuid.New()
	mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("record not found"))

	result, err := mockRepo.GetByID(context.Background(), id)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMockSubmissionRepo_GetPending(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)

	expected := []domain.Submission{
		{ID: uuid.New(), Status: domain.StatusPending},
		{ID: uuid.New(), Status: domain.StatusPending},
	}
	mockRepo.On("GetPending", mock.Anything).Return(expected, nil)

	results, err := mockRepo.GetPending(context.Background())

	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestMockSubmissionRepo_GetApproved(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)

	expected := []domain.Submission{
		{ID: uuid.New(), Status: domain.StatusApproved},
	}
	mockRepo.On("GetApproved", mock.Anything).Return(expected, nil)

	results, err := mockRepo.GetApproved(context.Background())

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, domain.StatusApproved, results[0].Status)
}

func TestMockSubmissionRepo_UpdateStatus(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)

	id := uuid.New()
	mockRepo.On("UpdateStatus", mock.Anything, id, domain.StatusApproved).Return(nil)

	err := mockRepo.UpdateStatus(context.Background(), id, domain.StatusApproved)

	assert.NoError(t, err)
}

func TestMockSubmissionRepo_Delete(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	err := mockRepo.Delete(context.Background(), id)

	assert.NoError(t, err)
}
