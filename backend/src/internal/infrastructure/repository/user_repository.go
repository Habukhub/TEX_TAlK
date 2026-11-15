package repository

import (
	"context"
	"errors"
	"strings"

	"tex/src/internal/domain"

	"gorm.io/gorm"
)

type UserRepository interface {
	VerifyAndSyncUser(ctx context.Context, googleID, email, name string) (*domain.User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func EnsureAdminUsers(ctx context.Context, db *gorm.DB, adminEmails string) error {
	for _, rawEmail := range strings.Split(adminEmails, ",") {
		email := strings.TrimSpace(rawEmail)
		if email == "" {
			continue
		}

		var user domain.User
		err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		name := strings.Split(email, "@")[0]
		if err := db.WithContext(ctx).Create(&domain.User{
			Email:    email,
			Name:     name,
			GoogleID: "",
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *userRepository) VerifyAndSyncUser(ctx context.Context, googleID, email, name string) (*domain.User, error) {
	var user domain.User

	// 1. ค้นหาด้วย Email เท่านั้น
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		// คืนค่า error (มักจะเป็น gorm.ErrRecordNotFound ถ้าไม่มีอีเมลนี้)
		return nil, err
	}

	// 2. ถ้าเจอ ให้อัปเดตข้อมูลจาก Google (GoogleID เผื่อมีการเชื่อมใหม่, Name เผื่อมีการเปลี่ยนชื่อ)
	user.GoogleID = googleID
	user.Name = name
	err = r.db.WithContext(ctx).Save(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
