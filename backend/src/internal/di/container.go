package di

import (
	"context"
	"log"

	"tex/src/internal/configs"
	"tex/src/internal/handler"
	"tex/src/internal/handler/appconfig"
	"tex/src/internal/handler/auth"
	"tex/src/internal/handler/schedule"
	"tex/src/internal/handler/submission"
	"tex/src/internal/handler/systemsetting"
	"tex/src/internal/infrastructure/firebase"
	"tex/src/internal/infrastructure/repository"
	"tex/src/internal/usecase"

	"gorm.io/gorm"
)

func BuildHandlers(db *gorm.DB, conf *configs.Config) *handler.Handlers {
	// 1. Repositories
	submissionRepo := repository.NewSubmissionRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	userRepo := repository.NewUserRepository(db)
	systemSettingRepo := repository.NewSystemSettingRepository(db)

	// 2. Use Cases
	submissionUC := usecase.NewSubmissionUseCase(submissionRepo)
	scheduleUC := usecase.NewScheduleUseCase(scheduleRepo)
	authUC := usecase.NewAuthUseCase(userRepo, conf.Auth.JWTSecret)
	systemSettingUC := usecase.NewSystemSettingUseCase(systemSettingRepo)

	// 3. Firebase Storage (optional — skipped when storage bucket is not configured)
	// credentialFile is optional: empty = use ADC (Application Default Credentials) on Cloud Run
	var firebaser *firebase.StorageService
	if conf.Firebase.StorageBucket != "" {
		svc, err := firebase.NewStorageService(context.Background(), conf.Firebase.CredentialFile, conf.Firebase.StorageBucket)
		if err != nil {
			log.Printf("[WARN] Firebase storage init failed: %v — slide upload/download will not work", err)
		} else {
			firebaser = svc
		}
	}

	// 4. Handlers
	submissionHandler := submission.NewHandler(submissionUC, systemSettingUC, firebaser)
	scheduleHandler := schedule.NewHandler(scheduleUC)
	authHandler := auth.NewHandler(authUC, conf)
	appConfigHandler := appconfig.NewHandler(conf)
	systemSettingHandler := systemsetting.NewHandler(systemSettingUC)

	// 5. Return combined handlers
	return handler.NewHandlers(submissionHandler, scheduleHandler, authHandler, appConfigHandler, systemSettingHandler)
}
