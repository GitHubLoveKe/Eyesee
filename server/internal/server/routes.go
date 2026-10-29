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

	rcRepo := postgres.NewRCTaskRepo(s.pool)
	rcSvc := service.NewRCTaskService(rcRepo)
	rcHandler := handler.NewRCTaskHandler(rcSvc)

	mdRepo := postgres.NewMetricDefinitionRepo(s.pool)
	mdSvc := service.NewMetricDefinitionService(mdRepo)
	mdHandler := handler.NewMetricDefinitionHandler(mdSvc)

	credRepo := postgres.NewCredentialRepo(s.pool)
	credSvc := service.NewCredentialService(credRepo)
	credHandler := handler.NewCredentialHandler(credSvc)

	sdSvc := service.NewServiceDiscoveryService(ctRepo, instRepo)
	sdHandler := handler.NewServiceDiscoveryHandler(sdSvc)

	eventRepo := postgres.NewAlertEventRepo(s.pool)
	ackRepo := postgres.NewAlertAckRepo(s.pool)
	silenceRepo := postgres.NewAlertSilenceRepo(s.pool)
	acSvc := service.NewAlertConvergenceService(eventRepo, ackRepo, silenceRepo)
	acHandler := handler.NewAlertConvergenceHandler(acSvc)

	untRepo := postgres.NewUserNotificationTargetRepo(s.pool)
	uaSvc := service.NewUserAuthService(userRepo, untRepo)
	uaHandler := handler.NewUserAuthHandler(uaSvc)

	zcRepo := postgres.NewZoneCoordinatorRepo(s.pool)
	zcSvc := service.NewZoneCoordinatorService(zcRepo)
	zcHandler := handler.NewZoneCoordinatorHandler(zcSvc)

	sdcRepo := postgres.NewServiceDiscoveryConfigRepo(s.pool)
	sdcSvc := service.NewServiceDiscoveryConfigService(sdcRepo)
	sdcHandler := handler.NewServiceDiscoveryConfigHandler(sdcSvc)

	ttRepo := postgres.NewTaskTypeRepo(s.pool)
	ttSvc := service.NewTaskTypeService(ttRepo)
	ttHandler := handler.NewTaskTypeHandler(ttSvc)

	boRepo := postgres.NewBusinessOverviewRepo(s.pool)
	boSvc := service.NewBusinessOverviewService(boRepo)
	boHandler := handler.NewBusinessOverviewHandler(boSvc)

	nsRepo := postgres.NewNetworkSegmentRepo(s.pool)
	nsSvc := service.NewNetworkSegmentService(nsRepo)
	nsHandler := handler.NewNetworkSegmentHandler(nsSvc)

	r.Mount("/business-systems", bsHandler.Routes())
	r.Route("/zones", func(r chi.Router) {
		// Zone routes
		r.Post("/", zoneHandler.Create)
		r.Get("/", zoneHandler.List)
		r.Get("/{id}", zoneHandler.GetByID)
		r.Put("/{id}", zoneHandler.Update)
		r.Delete("/{id}", zoneHandler.Delete)

		// Network segment routes (zone-scoped)
		r.Mount("/{zoneId}/segments", nsHandler.Routes())
	})
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
	r.Mount("/rc-tasks", rcHandler.Routes())
	r.Mount("/metric-definitions", mdHandler.Routes())
	r.Mount("/credentials", credHandler.Routes())
	r.Mount("/service-discovery", sdHandler.Routes())
	r.Mount("/alert-events", acHandler.EventRoutes())
	r.Mount("/alert-silences", acHandler.SilenceRoutes())
	r.Mount("/user-auth", uaHandler.Routes())
	r.Mount("/zone-coordinators", zcHandler.Routes())
	r.Mount("/sd-configs", sdcHandler.Routes())
	r.Mount("/business-overview", boHandler.Routes())
	r.Mount("/task-types", ttHandler.Routes())
}
