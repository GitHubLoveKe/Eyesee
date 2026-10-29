package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"coordinator/config"
	"coordinator/distribution"
	"coordinator/health"
	"coordinator/model"
	"coordinator/module"
)

type Server struct {
	cfg         *config.Config
	store       *model.Store
	distributor *distribution.Distributor
	broker      *distribution.EventBroker
	healthChk   *health.Checker
	logger      *slog.Logger
	mux         *http.ServeMux
	httpServer  *http.Server

	configGenerators  map[model.WorkerType]module.ConfigGenerator
	validWorkerTypes  map[model.WorkerType]struct{}
}

func NewServer(
	cfg *config.Config,
	store *model.Store,
	dist *distribution.Distributor,
	broker *distribution.EventBroker,
	hc *health.Checker,
	modules []module.Module,
	logger *slog.Logger,
) *Server {
	s := &Server{
		cfg:                cfg,
		store:              store,
		distributor:        dist,
		broker:             broker,
		healthChk:          hc,
		logger:             logger.With("component", "api"),
		mux:                http.NewServeMux(),
		configGenerators:   make(map[model.WorkerType]module.ConfigGenerator),
		validWorkerTypes:   make(map[model.WorkerType]struct{}),
	}

	for _, m := range modules {
		for _, route := range m.Routes() {
			s.validWorkerTypes[route.WorkerType] = struct{}{}
		}
		if cg, ok := m.(module.ConfigGenerator); ok {
			for _, route := range m.Routes() {
				s.configGenerators[route.WorkerType] = cg
			}
		}
		if ra, ok := m.(module.RouteRegistrar); ok {
			ra.RegisterRoutes(s)
		}
	}

	s.registerRoutes()
	return s
}

// Handle 实现 module.RouteRegistry 接口，供模块注册自定义路由。
func (s *Server) Handle(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/v1/workers/register", s.handleWorkerRegister)
	s.mux.HandleFunc("/api/v1/workers/", s.routeWorker)
	s.mux.HandleFunc("/api/v1/push/resources", s.handlePushResources)
	s.mux.HandleFunc("/api/v1/push/tasks", s.handlePushTasks)
	s.mux.HandleFunc("/api/v1/push/rules", s.handlePushRules)
	s.mux.HandleFunc("/api/v1/status", s.handleStatus)
	s.mux.HandleFunc("/api/v1/assignments", s.handleAssignments)
	s.mux.HandleFunc("/api/v1/stream/assignments", s.handleSSEStream)
}

func (s *Server) routeWorker(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/workers/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	workerID := parts[0]
	subPath := strings.Join(parts[1:], "/")

	switch {
	case subPath == "heartbeat" && r.Method == http.MethodPost:
		s.handleWorkerHeartbeat(w, r, workerID)
	case subPath == "assignments" && r.Method == http.MethodGet:
		s.handleWorkerAssignments(w, r, workerID)
	case subPath == "config" && r.Method == http.MethodGet:
		s.handleWorkerConfig(w, r, workerID)
	case subPath == "scrape-config" && r.Method == http.MethodGet:
		s.handleWorkerConfig(w, r, workerID)
	case subPath == "rules" && r.Method == http.MethodGet:
		s.handleWorkerConfig(w, r, workerID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) Listen() (net.Listener, error) {
	ln, err := net.Listen("tcp", s.cfg.Server.HTTPAddr)
	if err != nil {
		return nil, fmt.Errorf("监听 %s 失败: %w", s.cfg.Server.HTTPAddr, err)
	}
	s.httpServer = &http.Server{
		Handler: s.mux,
	}
	return ln, nil
}

func (s *Server) Serve(ctx context.Context, ln net.Listener) {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(shutdownCtx)
	}()

	s.logger.Info("HTTP API 已启动", "addr", ln.Addr().String())
	if err := s.httpServer.Serve(ln); err != http.ErrServerClosed {
		s.logger.Error("HTTP API 异常退出", "error", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (s *Server) handleWorkerRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reg struct {
		WorkerID       string            `json:"worker_id"`
		WorkerType     string            `json:"worker_type"`
		Address        string            `json:"address"`
		HealthEndpoint string            `json:"health_endpoint"`
		Capacity       int               `json:"capacity"`
		Labels         map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if reg.WorkerID == "" || reg.WorkerType == "" || reg.Address == "" {
		writeError(w, http.StatusBadRequest, "worker_id, worker_type, address are required")
		return
	}

	wt := model.WorkerType(reg.WorkerType)
	if _, ok := s.validWorkerTypes[wt]; !ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported worker_type: %s", reg.WorkerType))
		return
	}

	if err := validateWorkerAddress(reg.Address); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid address: %v", err))
		return
	}

	worker := &model.Worker{
		ID:             reg.WorkerID,
		Type:           wt,
		Address:        reg.Address,
		HealthEndpoint: reg.HealthEndpoint,
		Capacity:       reg.Capacity,
		Labels:         reg.Labels,
		Status:         model.StatusHealthy,
	}
	s.store.UpsertWorker(worker)

	s.logger.Info("worker 已注册", "worker_id", reg.WorkerID, "type", reg.WorkerType, "address", reg.Address)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"worker_id":  reg.WorkerID,
			"registered": true,
			"zone_id":    s.cfg.ZoneID,
		},
	})
}

func (s *Server) handleWorkerHeartbeat(w http.ResponseWriter, r *http.Request, workerID string) {
	var hb struct {
		Timestamp time.Time              `json:"timestamp"`
		Status    string                 `json:"status"`
		Metrics   map[string]interface{} `json:"metrics"`
	}
	if err := json.NewDecoder(r.Body).Decode(&hb); err != nil {
		writeError(w, http.StatusBadRequest, "invalid heartbeat body")
		return
	}

	s.healthChk.RecordHeartbeat(workerID)

	assignment := s.distributor.GetAssignment(workerID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"worker_id":   workerID,
			"task_count":  len(assignment.TaskIDs),
			"rule_count":  len(assignment.RuleIDs),
			"plan_version": s.distributor.GetPlan().Version,
		},
	})
}

func (s *Server) handleWorkerAssignments(w http.ResponseWriter, _ *http.Request, workerID string) {
	assignment := s.distributor.GetAssignment(workerID)
	plan := s.distributor.GetPlan()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"worker_id":    workerID,
			"task_ids":     assignment.TaskIDs,
			"rule_ids":     assignment.RuleIDs,
			"plan_version": plan.Version,
			"updated_at":   plan.UpdatedAt,
		},
	})
}

// handleWorkerConfig 通用配置下发：根据 worker 类型查找对应模块的 ConfigGenerator。
func (s *Server) handleWorkerConfig(w http.ResponseWriter, _ *http.Request, workerID string) {
	worker := s.store.GetWorker(workerID)
	if worker == nil {
		writeError(w, http.StatusNotFound, "worker not found")
		return
	}

	cg, ok := s.configGenerators[worker.Type]
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("no config generator for worker type: %s", worker.Type))
		return
	}

	assignment := s.distributor.GetAssignment(workerID)
	data, err := cg.GenerateWorkerConfig(workerID, assignment)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("generate config failed: %v", err))
		return
	}

	w.Header().Set("Content-Type", cg.ConfigContentType())
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) handlePushResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body failed")
		return
	}

	var resources []*model.Resource
	if err := json.Unmarshal(body, &resources); err != nil {
		writeError(w, http.StatusBadRequest, "invalid resources JSON")
		return
	}

	s.store.ReplaceResources(resources)
	s.distributor.Rebalance()
	s.broker.Broadcast(distribution.SSEEvent{
		Event: "resources_updated",
		Data:  map[string]interface{}{"count": len(resources)},
	})

	s.logger.Info("资源列表已更新", "count", len(resources))
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{"updated": len(resources)},
	})
}

func (s *Server) handlePushTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body failed")
		return
	}

	var tasks []*model.CollectionTask
	if err := json.Unmarshal(body, &tasks); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tasks JSON")
		return
	}

	s.store.ReplaceTasks(tasks)
	s.distributor.Rebalance()
	s.broker.Broadcast(distribution.SSEEvent{
		Event: "tasks_updated",
		Data:  map[string]interface{}{"count": len(tasks)},
	})

	s.logger.Info("采集任务已更新", "count", len(tasks))
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{"updated": len(tasks)},
	})
}

func (s *Server) handlePushRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body failed")
		return
	}

	var rules []*model.AlertRule
	if err := json.Unmarshal(body, &rules); err != nil {
		writeError(w, http.StatusBadRequest, "invalid rules JSON")
		return
	}

	s.store.ReplaceRules(rules)
	s.distributor.Rebalance()
	s.broker.Broadcast(distribution.SSEEvent{
		Event: "rules_updated",
		Data:  map[string]interface{}{"count": len(rules)},
	})

	s.logger.Info("告警规则已更新", "count", len(rules))
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{"updated": len(rules)},
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	plan := s.distributor.GetPlan()
	workers := s.store.GetAllWorkers()
	resources := s.store.GetResources()

	workerStats := make(map[string]int)
	for _, w := range workers {
		workerStats[string(w.Status)]++
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": map[string]interface{}{
			"zone_id":          s.cfg.ZoneID,
			"plan_version":     plan.Version,
			"total_workers":    len(workers),
			"worker_stats":     workerStats,
			"total_resources":  len(resources),
			"total_tasks":      len(s.store.GetEnabledTasks()),
			"total_rules":      len(s.store.GetAllRules()),
			"last_rebalance":   plan.UpdatedAt,
		},
	})
}

func (s *Server) handleAssignments(w http.ResponseWriter, _ *http.Request) {
	plan := s.distributor.GetPlan()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code": 0,
		"data": plan,
	})
}

func (s *Server) handleSSEStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := s.broker.Subscribe()
	defer s.broker.Unsubscribe(ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, data)
			flusher.Flush()
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"code":    status,
		"message": message,
	})
}

func validateWorkerAddress(addr string) error {
	u, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("scheme must be http or https, got %s", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("empty hostname")
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("link-local addresses are not allowed")
		}
		if ip.String() == "169.254.169.254" {
			return fmt.Errorf("cloud metadata address is not allowed")
		}
	}
	return nil
}
