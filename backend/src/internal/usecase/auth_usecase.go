package usecase

import (
	"context"
	"fmt"

	"tex/src/internal/domain"
	"tex/src/internal/infrastructure/repository"
	"tex/src/internal/util"
)

type AuthUseCase interface {
	HandleGoogleCallback(ctx context.Context, googleID, email, name string) (*domain.User, string, error)
}

type authUseCase struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthUseCase(userRepo repository.UserRepository, jwtSecret string) AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (u *authUseCase) HandleGoogleCallback(ctx context.Context, googleID, email, name string) (*domain.User, string, error) {
	// 1. ตรวจสอบว่ามี Email นี้ในระบบหรือไม่ ถ้ามีให้อัปเดตข้อมูล ถ้าไม่มีจะ Error ทันที
	user, err := u.userRepo.VerifyAndSyncUser(ctx, googleID, email, name)
	if err != nil {
		// ส่ง Error บอกว่าไม่มีสิทธิ์เข้าใช้งาน
		return nil, "", fmt.Errorf("unauthorized: email %s is not allowed to login", email)
	}

	// 2. สร้าง JWT Token
	token, err := util.GenerateJWT(user.ID, user.Email, u.jwtSecret)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}
