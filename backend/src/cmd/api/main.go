package main

// @title TEX TALK API
// @version 1.0
// @description This is the backend API for TEX TALK
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tex/src/internal/configs"
	"tex/src/internal/constant"
	"tex/src/internal/di"
	"tex/src/internal/handler"
	"tex/src/internal/infrastructure/database"
	appmiddleware "tex/src/internal/infrastructure/middleware"
	"tex/src/internal/infrastructure/repository"
	"tex/src/internal/infrastructure/router"
	"tex/src/internal/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"gorm.io/gorm"
)

func main() {
	setVersionFromFile()

	conf, err := configs.LoadConfig(configs.Configuration{
		ConfigPath: string(constant.ConfigsPath),
		FileName:   string(constant.ConfigsFile),
		FileType:   string(constant.ConfigsType),
	})
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	if err := conf.ValidateRuntime(); err != nil {
		log.Fatalf("invalid runtime configuration: %v", err)
	}

	setLocation(conf.App.Timezone)

	appLogger := logger.NewLogger(conf.App.ServiceName, conf.App.Environment)

	db, err := database.NewPostgresDB(conf)
	if err != nil {
		if strings.EqualFold(conf.App.Environment, "prod") {
			log.Fatalf("failed to connect database: %v", err)
		}
		appLogger.Warn("database connection skipped", "error", err.Error())
	}
	if db != nil && strings.TrimSpace(conf.Auth.AdminEmails) != "" {
		if err := repository.EnsureAdminUsers(context.Background(), db, conf.Auth.AdminEmails); err != nil {
			log.Fatalf("failed to bootstrap admin users: %v", err)
		}
	}

	handlers := di.BuildHandlers(db, conf)

	app := buildApp(conf, appLogger, handlers)
	setupGracefulShutdown(app, db, appLogger)

	appLogger.Info("starting server", "port", conf.App.Port)
	if err := app.Listen(":" + conf.App.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func buildApp(conf *configs.Config, appLogger *logger.Logger, handlers *handler.Handlers) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      conf.App.ServiceName,
		UnescapePath: true,
		BodyLimit:    50 * 1024 * 1024, // 50MB limit
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:  strings.Split(conf.App.CORSAllowedOrigins, ","),
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Trace-ID"},
		ExposeHeaders: []string{"Content-Disposition"},
		MaxAge:        3600,
	}))
	app.Use(recover.New())
	app.Use(appmiddleware.AccessLogMiddleware(appLogger))

	router.SetupRoutes(app, handlers, conf)

	return app
}

func setLocation(name string) {
	if name == "" {
		name = string(constant.Bangkok)
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		log.Printf("failed to load timezone %q: %v", name, err)
		return
	}

	time.Local = location
}

func setVersionFromFile() {
	if os.Getenv(constant.WordVersion) != "" {
		return
	}

	versionBytes, err := os.ReadFile("VERSION")
	if err != nil {
		return
	}

	version := strings.TrimSpace(string(versionBytes))
	if version == "" {
		return
	}

	_ = os.Setenv(constant.WordVersion, version)
}

func setupGracefulShutdown(app *fiber.App, db *gorm.DB, appLogger *logger.Logger) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if db != nil {
			sqlDB, err := db.DB()
			if err == nil {
				_ = sqlDB.Close()
			}
		}

		if err := app.ShutdownWithContext(ctx); err != nil {
			appLogger.Error("graceful shutdown failed", "error", err.Error())
			return
		}

		appLogger.Info("server stopped")
	}()
}
