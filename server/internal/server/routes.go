package server

import (
	"dcs/internal/handler"
	"dcs/internal/repository/postgres"
	"dcs/internal/service"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerRoutes(r chi.Router) {
	bsRepo := postgres.NewBusinessSystemRepo(s.pool)
	bsSvc := service.NewBusinessSystemService(bsRepo)
	bsHandler := handler.NewBusinessSystemHandler(bsSvc)

	zoneRepo := postgres.NewZoneRepo(s.pool)
	zoneSvc := service.NewZoneService(zoneRepo)
	zoneHandler := handler.NewZoneHandler(zoneSvc)

	rtRepo := postgres.NewResourceTypeRepo(s.pool)
	rtSvc := service.NewResourceSystemService(rtRepo)
	rtHandler := handler.NewResourceTypeHandler(rtSvc)

	instRepo := postgres.NewInstanceRepo(s.pool)
	instSvc := service.NewInstanceService(instRepo)
	instHandler := handler.NewInstanceHandler(instSvc)

	sinkRepo := postgres.NewDataSinkRepo(s.pool)
	sinkSvc := service.NewDataSinkService(sinkRepo)
	sinkHandler := handler.NewDataSinkHandler(sinkSvc)

	ctRepo := postgres.NewCollectionTaskRepo(s.pool)
	ctSvc := service.NewCollectionTaskService(ctRepo)
	ctHandler := handler.NewCollectionTaskHandler(ctSvc)

	cnRepo := postgres.NewCollectorNodeRepo(s.pool)
	cnSvc := service.NewCollectorService(cnRepo)
	cnHandler := handler.NewCollectorHandler(cnSvc)

	statusSvc := service.NewStatusCacheService(s.rdb)
	statusHandler := handler.NewStatusCacheHandler(statusSvc)

	userRepo := postgres.NewUserRepo(s.pool)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	ncRepo := postgres.NewNotificationChannelRepo(s.pool)
	ncSvc := service.NewNotificationChannelService(ncRepo)
	ncHandler := handler.NewNotificationChannelHandler(ncSvc)

	agRepo := postgres.NewAlertGroupRepo(s.pool)
	agSvc := service.NewAlertGroupService(agRepo)
	agHandler := handler.NewAlertGroupHandler(agSvc)

	asRepo := postgres.NewAlertStrategyRepo(s.pool)
	asSvc := service.NewAlertStrategyService(asRepo)
	asHandler := handler.NewAlertStrategyHandler(asSvc)

	stRepo := postgres.NewScheduledTaskRepo(s.pool)
	stSvc := service.NewScheduledTaskService(stRepo)
	stHandler := handler.NewScheduledTaskHandler(stSvc)

	r.Mount("/business-systems", bsHandler.Routes())
	r.Mount("/zones", zoneHandler.Routes())
	r.Mount("/resource-types", rtHandler.Routes())
	r.Mount("/instances", instHandler.Routes())
	r.Mount("/data-sinks", sinkHandler.Routes())
	r.Mount("/collection-tasks", ctHandler.Routes())
	r.Mount("/collector-nodes", cnHandler.Routes())
	r.Mount("/status", statusHandler.Routes())
	r.Mount("/users", userHandler.Routes())
	r.Mount("/notification-channels", ncHandler.Routes())
	r.Mount("/alert-groups", agHandler.Routes())
	r.Mount("/alert-strategies", asHandler.Routes())
	r.Mount("/scheduled-tasks", stHandler.Routes())
}
