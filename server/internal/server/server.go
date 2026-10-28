package server

import (
	"context"
	"dcs/internal/config"
	"dcs/internal/handler"
	"dcs/internal/middleware"
	"dcs/internal/repository/postgres"
	"dcs/internal/service"
	"dcs/pkg/response"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	cfg    *config.Config
	logger *slog.Logger
	pool   *pgxpool.Pool
	rdb    *redis.Client
	router chi.Router
}

func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		pool.Close()
		return nil, err
	}

	s := &Server{
		cfg:    cfg,
		logger: logger,
		pool:   pool,
		rdb:    rdb,
	}

	s.setupRouter()
	return s, nil
}

func (s *Server) setupRouter() {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(middleware.Logging(s.logger))
	r.Use(middleware.CORS(middleware.DefaultCORSConfig()))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		authHandler, jwtMiddleware := s.setupAuth(r)

		// Public auth routes (no JWT required)
		r.Mount("/auth", authHandler.Routes())

		// Protected routes (JWT required)
		r.Group(func(r chi.Router) {
			r.Use(jwtMiddleware)
			s.registerRoutes(r)
		})
	})

	s.router = r
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:         s.cfg.Server.Addr,
		Handler:      s.router,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	s.logger.Info("server starting", "addr", s.cfg.Server.Addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Close() {
	s.pool.Close()
	s.rdb.Close()
}

func (s *Server) setupAuth(r chi.Router) (*handler.AuthHandler, func(http.Handler) http.Handler) {
	userRepo := postgres.NewUserRepo(s.pool)
	authSrcRepo := postgres.NewUserAuthSourceRepo(s.pool)
	authSvc := service.NewAuthService(userRepo, authSrcRepo, &s.cfg.JWT)
	authHandler := handler.NewAuthHandler(authSvc)
	jwtMiddleware := middleware.JWT(authSvc, &s.cfg.JWT)
	return authHandler, jwtMiddleware
}
