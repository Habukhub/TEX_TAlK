package usecase

import (
	"context"
	"errors"

	"tex/src/internal/domain"
	"tex/src/internal/dto"
	"tex/src/internal/infrastructure/repository"

	"github.com/google/uuid"
)

var ErrSubmissionNotApproved = errors.New("submission is not approved")

type SubmissionUseCase interface {
	Submit(ctx context.Context, req dto.SubmitRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Submission, error)
	GetPending(ctx context.Context) ([]domain.Submission, error)
	GetApproved(ctx context.Context) ([]domain.Submission, error)
	Approve(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
	UpdateSlide(ctx context.Context, id uuid.UUID, slideURL string, slideFileName string) error
	UpdateInfo(ctx context.Context, id uuid.UUID, req dto.UpdateSubmissionInfoRequest) error
}

type submissionUseCase struct {
	repo repository.SubmissionRepository
}

func NewSubmissionUseCase(repo repository.SubmissionRepository) SubmissionUseCase {
	return &submissionUseCase{repo: repo}
}

func (u *submissionUseCase) Submit(ctx context.Context, req dto.SubmitRequest) error {
	submission := &domain.Submission{
		FullName:   req.FullName,
		Email:      req.Email,
		TopicTitle: req.TopicTitle,
		Status:     domain.StatusPending,
	}
	return u.repo.Create(ctx, submission)
}

func (u *submissionUseCase) GetPending(ctx context.Context) ([]domain.Submission, error) {
	return u.repo.GetPending(ctx)
}

func (u *submissionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *submissionUseCase) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Submission, error) {
	return u.repo.GetByIDs(ctx, ids)
}

func (u *submissionUseCase) GetApproved(ctx context.Context) ([]domain.Submission, error) {
	return u.repo.GetApproved(ctx)
}

func (u *submissionUseCase) Approve(ctx context.Context, id uuid.UUID) error {
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.repo.UpdateStatus(ctx, id, domain.StatusApproved)
}

func (u *submissionUseCase) Reject(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func (u *submissionUseCase) UpdateSlide(ctx context.Context, id uuid.UUID, slideURL string, slideFileName string) error {
	sub, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if sub.Status != domain.StatusApproved {
		return ErrSubmissionNotApproved
	}
	return u.repo.UpdateSlide(ctx, id, slideURL, slideFileName)
}

func (u *submissionUseCase) UpdateInfo(ctx context.Context, id uuid.UUID, req dto.UpdateSubmissionInfoRequest) error {
	_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.repo.UpdateInfo(ctx, id, req.FullName, req.TopicTitle)
}
