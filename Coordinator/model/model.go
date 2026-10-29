package model

import (
	"encoding/json"
	"sync"
	"time"
)

type ResourceType string

const (
	ResourceVMAgent        ResourceType = "vmagent"
	ResourceVMAlert        ResourceType = "vmalert"
	ResourceVM             ResourceType = "victoriametrics"
	ResourcePrometheus     ResourceType = "prometheus"
	ResourceOpenSearch     ResourceType = "opensearch"
	ResourceElasticsearch  ResourceType = "elasticsearch"
	ResourceOpenObserve    ResourceType = "openobserve"
	ResourceOTelCollector  ResourceType = "otel_collector"
	ResourceS3             ResourceType = "s3"
	ResourceExporter       ResourceType = "exporter"
)

type WorkerType string

const (
	WorkerVMAgent       WorkerType = "vmagent"
	WorkerVMAlert       WorkerType = "vmalert"
	WorkerOTelCollector WorkerType = "otel-collector"
	WorkerExporter      WorkerType = "exporter"
)

type NodeStatus string

const (
	StatusHealthy   NodeStatus = "healthy"
	StatusDegraded  NodeStatus = "degraded"
	StatusOffline   NodeStatus = "offline"
	StatusUnknown   NodeStatus = "unknown"
)

type Resource struct {
	ID             string          `json:"id"`
	Type           ResourceType    `json:"type"`
	ZoneID         string          `json:"zone_id"`
	Address        string          `json:"address"`
	HealthEndpoint string          `json:"health_endpoint"`
	Labels         map[string]string `json:"labels"`
	Status         NodeStatus      `json:"status"`
	LastHeartbeat  time.Time       `json:"last_heartbeat"`
	LastProbeOK    bool            `json:"-"`
	ProbeFailCount int             `json:"-"`
}

type Worker struct {
	ID             string          `json:"id"`
	Type           WorkerType      `json:"type"`
	ZoneID         string          `json:"zone_id"`
	Address        string          `json:"address"`
	HealthEndpoint string          `json:"health_endpoint"`
	Capacity       int             `json:"capacity"`
	Labels         map[string]string `json:"labels"`
	Status         NodeStatus      `json:"status"`
	LastHeartbeat  time.Time       `json:"last_heartbeat"`
	ProbeFailCount int             `json:"-"`
}

type CollectionTask struct {
	TaskID         string            `json:"task_id"`
	Module         string            `json:"module"`
	InstanceID     string            `json:"instance_id"`
	Name           string            `json:"name"`
	ZoneID         string            `json:"zone_id"`
	Config         json.RawMessage   `json:"config"`
	Enabled        bool              `json:"enabled"`
	Target         string            `json:"target"`
	JobName        string            `json:"job_name"`
	MetricsPath    string            `json:"metrics_path"`
	ScrapeInterval string            `json:"scrape_interval"`
	Labels         map[string]string `json:"labels"`
	Auth           *AuthConfig       `json:"auth,omitempty"`
}

// ModuleName 返回任务所属模块名。为空时默认 "metrics"（向后兼容）。
func (t *CollectionTask) ModuleName() string {
	if t.Module == "" {
		return "metrics"
	}
	return t.Module
}

type AuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type AlertRule struct {
	RuleID        string            `json:"rule_id"`
	Module        string            `json:"module"`
	TaskID        string            `json:"task_id"`
	ZoneID        string            `json:"zone_id"`
	AlertName     string            `json:"alert_name"`
	Expr          string            `json:"expr"`
	Duration      string            `json:"duration"`
	Severity      string            `json:"severity"`
	Labels        map[string]string `json:"labels"`
	Annotations   map[string]string `json:"annotations"`
	GroupName     string            `json:"group_name"`
	GroupInterval string            `json:"group_interval"`
}

// ModuleName 返回规则所属模块名。为空时默认 "metrics"（向后兼容）。
func (r *AlertRule) ModuleName() string {
	if r.Module == "" {
		return "metrics"
	}
	return r.Module
}

type Store struct {
	mu        sync.RWMutex
	zoneID    string
	resources map[string]*Resource
	workers   map[string]*Worker
	tasks     map[string]*CollectionTask
	rules     map[string]*AlertRule
}

func NewStore(zoneID string) *Store {
	return &Store{
		zoneID:    zoneID,
		resources: make(map[string]*Resource),
		workers:   make(map[string]*Worker),
		tasks:     make(map[string]*CollectionTask),
		rules:     make(map[string]*AlertRule),
	}
}

func (s *Store) UpsertResource(r *Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r.ZoneID = s.zoneID
	s.resources[r.ID] = r
}

func (s *Store) RemoveResource(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.resources, id)
}

func (s *Store) GetResources() []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Resource, 0, len(s.resources))
	for _, r := range s.resources {
		out = append(out, r)
	}
	return out
}

func (s *Store) GetResourcesByType(t ResourceType) []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Resource
	for _, r := range s.resources {
		if r.Type == t {
			out = append(out, r)
		}
	}
	return out
}

func (s *Store) UpsertWorker(w *Worker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.ZoneID = s.zoneID
	if w.HealthEndpoint == "" {
		w.HealthEndpoint = "/-/healthy"
	}
	if w.Capacity <= 0 {
		w.Capacity = 500
	}
	if w.Status == "" {
		w.Status = StatusUnknown
	}
	s.workers[w.ID] = w
}

func (s *Store) GetWorker(id string) *Worker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workers[id]
}

func (s *Store) GetHealthyWorkers(typ WorkerType) []*Worker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Worker
	for _, w := range s.workers {
		if w.Type == typ && w.Status == StatusHealthy {
			out = append(out, w)
		}
	}
	return out
}

func (s *Store) GetAllWorkers() []*Worker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Worker, 0, len(s.workers))
	for _, w := range s.workers {
		out = append(out, w)
	}
	return out
}

func (s *Store) RemoveWorker(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.workers, id)
}

func (s *Store) UpsertTask(t *CollectionTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ZoneID = s.zoneID
	s.tasks[t.TaskID] = t
}

func (s *Store) RemoveTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
}

func (s *Store) GetEnabledTasks() []*CollectionTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*CollectionTask
	for _, t := range s.tasks {
		if t.Enabled {
			out = append(out, t)
		}
	}
	return out
}

func (s *Store) UpsertRule(r *AlertRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r.ZoneID = s.zoneID
	s.rules[r.RuleID] = r
}

func (s *Store) RemoveRule(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rules, id)
}

func (s *Store) GetAllRules() []*AlertRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*AlertRule, 0, len(s.rules))
	for _, r := range s.rules {
		out = append(out, r)
	}
	return out
}

func (s *Store) GetTasksByModule(moduleName string) []*CollectionTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*CollectionTask
	for _, t := range s.tasks {
		if t.Enabled && t.ModuleName() == moduleName {
			out = append(out, t)
		}
	}
	return out
}

func (s *Store) GetRulesByModule(moduleName string) []*AlertRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*AlertRule
	for _, r := range s.rules {
		if r.ModuleName() == moduleName {
			out = append(out, r)
		}
	}
	return out
}

func (s *Store) ReplaceTasks(tasks []*CollectionTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = make(map[string]*CollectionTask, len(tasks))
	for _, t := range tasks {
		t.ZoneID = s.zoneID
		s.tasks[t.TaskID] = t
	}
}

func (s *Store) ReplaceRules(rules []*AlertRule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules = make(map[string]*AlertRule, len(rules))
	for _, r := range rules {
		r.ZoneID = s.zoneID
		s.rules[r.RuleID] = r
	}
}

func (s *Store) ReplaceResources(resources []*Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources = make(map[string]*Resource, len(resources))
	for _, r := range resources {
		r.ZoneID = s.zoneID
		s.resources[r.ID] = r
	}
}

type Assignment struct {
	WorkerID string   `json:"worker_id"`
	TaskIDs  []string `json:"task_ids"`
	RuleIDs  []string `json:"rule_ids"`
}

type AssignmentPlan struct {
	ZoneID      string                 `json:"zone_id"`
	Assignments map[string]*Assignment `json:"assignments"`
	Version     int64                  `json:"version"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
