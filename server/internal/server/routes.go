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

	r.Mount("/business-systems", bsHandler.Routes())
	r.Mount("/zones", zoneHandler.Routes())
	r.Mount("/resource-types", rtHandler.Routes())
	r.Mount("/instances", instHandler.Routes())
	r.Mount("/data-sinks", sinkHandler.Routes())
	r.Mount("/collection-tasks", ctHandler.Routes())
	r.Mount("/collector-nodes", cnHandler.Routes())
	r.Mount("/status", statusHandler.Routes())
}
