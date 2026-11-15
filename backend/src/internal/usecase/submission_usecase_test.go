package usecase_test

import (
	"context"
	"errors"
	"testing"

	"tex/src/internal/domain"
	"tex/src/internal/dto"
	"tex/src/internal/testutils"
	"tex/src/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubmit_Success(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	req := dto.SubmitRequest{
		FullName:   "สมชาย ใจดี",
		Email:      "somchai@example.com",
		TopicTitle: "หัวข้อทดสอบ",
	}

	// กำหนดให้ repo.Create ตอบกลับ nil (สำเร็จ)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Submission) bool {
		return s.FullName == req.FullName &&
			s.Email == req.Email &&
			s.TopicTitle == req.TopicTitle &&
			s.Status == domain.StatusPending
	})).Return(nil)

	err := uc.Submit(context.Background(), req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSubmit_RepoError(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	req := dto.SubmitRequest{FullName: "Test", Email: "test@example.com", TopicTitle: "Topic"}
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Submission")).Return(errors.New("db error"))

	err := uc.Submit(context.Background(), req)

	assert.Error(t, err)
	assert.EqualError(t, err, "db error")
}

func TestGetPending_Success(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	expected := []domain.Submission{
		{ID: uuid.New(), FullName: "คนที่ 1", Status: domain.StatusPending},
		{ID: uuid.New(), FullName: "คนที่ 2", Status: domain.StatusPending},
	}
	mockRepo.On("GetPending", mock.Anything).Return(expected, nil)

	result, err := uc.GetPending(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetApproved_Success(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	expected := []domain.Submission{
		{ID: uuid.New(), FullName: "คนที่ 1", Status: domain.StatusApproved},
	}
	mockRepo.On("GetApproved", mock.Anything).Return(expected, nil)

	result, err := uc.GetApproved(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, domain.StatusApproved, result[0].Status)
}

func TestApprove_Success(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	id := uuid.New()
	submission := &domain.Submission{ID: id, Status: domain.StatusPending}

	// Approve ต้อง GetByID ก่อน แล้วค่อย UpdateStatus
	mockRepo.On("GetByID", mock.Anything, id).Return(submission, nil)
	mockRepo.On("UpdateStatus", mock.Anything, id, domain.StatusApproved).Return(nil)

	err := uc.Approve(context.Background(), id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestApprove_NotFound(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	id := uuid.New()
	mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("record not found"))

	err := uc.Approve(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "record not found")
	// UpdateStatus ต้องไม่ถูกเรียกเลย
	mockRepo.AssertNotCalled(t, "UpdateStatus")
}

func TestReject_Success(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	err := uc.Reject(context.Background(), id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestReject_RepoError(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(errors.New("db error"))

	err := uc.Reject(context.Background(), id)

	assert.Error(t, err)
}

func TestUpdateInfo_Success(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	id := uuid.New()
	submission := &domain.Submission{ID: id, Status: domain.StatusApproved}
	req := dto.UpdateSubmissionInfoRequest{FullName: "New Name", TopicTitle: "New Topic"}

	mockRepo.On("GetByID", mock.Anything, id).Return(submission, nil)
	mockRepo.On("UpdateInfo", mock.Anything, id, "New Name", "New Topic").Return(nil)

	err := uc.UpdateInfo(context.Background(), id, req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateInfo_NotFound(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	id := uuid.New()
	req := dto.UpdateSubmissionInfoRequest{FullName: "Name", TopicTitle: "Topic"}
	mockRepo.On("GetByID", mock.Anything, id).Return(nil, errors.New("record not found"))

	err := uc.UpdateInfo(context.Background(), id, req)

	assert.Error(t, err)
	assert.EqualError(t, err, "record not found")
	mockRepo.AssertNotCalled(t, "UpdateInfo")
}

func TestUpdateInfo_RepoError(t *testing.T) {
	mockRepo := new(testutils.MockSubmissionRepository)
	uc := usecase.NewSubmissionUseCase(mockRepo)

	id := uuid.New()
	submission := &domain.Submission{ID: id, Status: domain.StatusPending}
	req := dto.UpdateSubmissionInfoRequest{FullName: "Name", TopicTitle: "Topic"}

	mockRepo.On("GetByID", mock.Anything, id).Return(submission, nil)
	mockRepo.On("UpdateInfo", mock.Anything, id, "Name", "Topic").Return(errors.New("db error"))

	err := uc.UpdateInfo(context.Background(), id, req)

	assert.Error(t, err)
	assert.EqualError(t, err, "db error")
}
