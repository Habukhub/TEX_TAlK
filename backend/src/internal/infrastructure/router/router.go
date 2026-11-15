package router

import (
	"strings"
	"time"

	"tex/src/internal/configs"
	"tex/src/internal/constant"
	"tex/src/internal/handler"
	"tex/src/internal/infrastructure/middleware"

	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	_ "tex/docs"
)

func SetupRoutes(app *fiber.App, handlers *handler.Handlers, conf *configs.Config) {
	if !strings.EqualFold(conf.App.Environment, "prod") {
		app.Get("/swagger/*", swaggo.HandlerDefault)
	}

	api := app.Group("/api/v1")
	api.Use(limiter.New(limiter.Config{
		Max:        60,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests, please slow down"})
		},
	}))

	// Public routes
	api.Get(constant.APIPathHealthz, handlers.HealthcheckHandler.GetHealthz)
	api.Get("/config", handlers.AppConfigHandler.GetConfig)
	api.Post("/submissions", handlers.SubmissionHandler.Submit)
	api.Get("/schedule", handlers.ScheduleHandler.GetTimeline)
	api.Get("/speakers", handlers.SubmissionHandler.GetPublicApproved)
	api.Put("/submissions/:id/slide", handlers.SubmissionHandler.UploadSlide)
	api.Get("/settings", handlers.SystemSettingHandler.GetSettings)

	// Auth routes
	auth := api.Group("/auth")
	auth.Get("/google", handlers.AuthHandler.GoogleLogin)
	auth.Get("/google/callback", handlers.AuthHandler.GoogleCallback)
	auth.Post("/logout", handlers.AuthHandler.Logout)

	// Admin routes (Protected by JWT)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(conf.Auth.JWTSecret))

	// Submissions
	admin.Get("/submissions/pending", handlers.SubmissionHandler.GetPending)
	admin.Get("/submissions/approved", handlers.SubmissionHandler.GetApproved)
	admin.Put("/submissions/:id/approve", handlers.SubmissionHandler.Approve)
	admin.Put("/submissions/:id/info", handlers.SubmissionHandler.UpdateInfo)
	admin.Delete("/submissions/:id", handlers.SubmissionHandler.Reject)
	admin.Get("/submissions/:id/slide", handlers.SubmissionHandler.DownloadSlide)
	admin.Post("/submissions/slides/zip", handlers.SubmissionHandler.DownloadZip)

	// Schedule
	admin.Post("/schedule", handlers.ScheduleHandler.AddSlot)
	admin.Delete("/schedule/:id", handlers.ScheduleHandler.DeleteSlot)

	// System Settings
	admin.Put("/settings/presentation_upload_enabled", handlers.SystemSettingHandler.UpdatePresentationUploadEnabled)
}
