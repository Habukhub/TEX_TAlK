package repository

import (
	"context"

	"tex/src/internal/domain"

	"gorm.io/gorm"
)

type SystemSettingRepository interface {
	GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error)
	UpdateValue(ctx context.Context, key string, value string) error
}

type systemSettingRepository struct {
	db *gorm.DB
}

func NewSystemSettingRepository(db *gorm.DB) SystemSettingRepository {
	return &systemSettingRepository{db: db}
}

func (r *systemSettingRepository) GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error) {
	var setting domain.SystemSetting
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *systemSettingRepository) UpdateValue(ctx context.Context, key string, value string) error {
	// Use Save or Updates to ensure we can update or insert (if using Clauses)
	// Assuming the migration already inserts the keys, we can just update.
	return r.db.WithContext(ctx).Model(&domain.SystemSetting{}).Where("key = ?", key).Update("value", value).Error
}
