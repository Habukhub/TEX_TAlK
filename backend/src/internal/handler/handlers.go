package handler

import (
	"tex/src/internal/handler/appconfig"
	"tex/src/internal/handler/auth"
	"tex/src/internal/handler/healthcheck"
	"tex/src/internal/handler/schedule"
	"tex/src/internal/handler/submission"
	"tex/src/internal/handler/systemsetting"
)

type Handlers struct {
	HealthcheckHandler   *healthcheck.Handler
	SubmissionHandler    *submission.Handler
	ScheduleHandler      *schedule.Handler
	AuthHandler          *auth.Handler
	AppConfigHandler     *appconfig.Handler
	SystemSettingHandler *systemsetting.Handler
}

func NewHandlers(submissionHandler *submission.Handler, scheduleHandler *schedule.Handler, authHandler *auth.Handler, appConfigHandler *appconfig.Handler, systemSettingHandler *systemsetting.Handler) *Handlers {
	return &Handlers{
		HealthcheckHandler:   healthcheck.NewHandler(),
		SubmissionHandler:    submissionHandler,
		ScheduleHandler:      scheduleHandler,
		AuthHandler:          authHandler,
		AppConfigHandler:     appConfigHandler,
		SystemSettingHandler: systemSettingHandler,
	}
}
