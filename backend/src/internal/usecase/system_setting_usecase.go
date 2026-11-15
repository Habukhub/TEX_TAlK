package usecase

import (
	"context"

	"tex/src/internal/domain"
	"tex/src/internal/infrastructure/repository"
)

type SystemSettingUseCase interface {
	IsPresentationUploadEnabled(ctx context.Context) (bool, error)
	SetPresentationUploadEnabled(ctx context.Context, enabled bool) error
}

type systemSettingUseCase struct {
	repo repository.SystemSettingRepository
}

func NewSystemSettingUseCase(repo repository.SystemSettingRepository) SystemSettingUseCase {
	return &systemSettingUseCase{repo: repo}
}

func (u *systemSettingUseCase) IsPresentationUploadEnabled(ctx context.Context) (bool, error) {
	setting, err := u.repo.GetByKey(ctx, domain.SettingKeyPresentationUploadEnabled)
	if err != nil {
		return false, err
	}
	return setting.Value == "true", nil
}

func (u *systemSettingUseCase) SetPresentationUploadEnabled(ctx context.Context, enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	return u.repo.UpdateValue(ctx, domain.SettingKeyPresentationUploadEnabled, val)
}
