package distribution

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log/slog"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"coordinator/config"
	"coordinator/model"
)

type Ring struct {
	partitions int
	nodes      []ringNode
}

type ringNode struct {
	hash uint32
	id   string
}

func NewRing(partitions int) *Ring {
	return &Ring{partitions: partitions}
}

func (r *Ring) Build(workerIDs []string) {
	r.nodes = make([]ringNode, 0, len(workerIDs)*r.partitions)
	for _, id := range workerIDs {
		for i := 0; i < r.partitions; i++ {
			h := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s:%d", id, i)))
			r.nodes = append(r.nodes, ringNode{hash: h, id: id})
		}
	}
	sort.Slice(r.nodes, func(i, j int) bool {
		return r.nodes[i].hash < r.nodes[j].hash
	})
}

func (r *Ring) Get(key string) string {
	if len(r.nodes) == 0 {
		return ""
	}
	h := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(r.nodes), func(i int) bool {
		return r.nodes[i].hash >= h
	})
	if idx >= len(r.nodes) {
		idx = 0
	}
	return r.nodes[idx].id
}

func (r *Ring) Size() int {
	return len(r.nodes) / max(r.partitions, 1)
}

// ModuleRoute 模块路由表条目：由 Coordinator 从各 Module.Routes() 聚合后传入。
type ModuleRoute struct {
	ModuleName string
	Routes     []TaskRoute
}

// TaskRoute 单条路由规则，对应 module.TaskRoute。
type TaskRoute struct {
	WorkerType model.WorkerType
	TaskFilter func(*model.CollectionTask) bool
	RuleFilter func(*model.AlertRule) bool
}

type Distributor struct {
	store  *model.Store
	cfg    *config.Config
	logger *slog.Logger

	mu           sync.RWMutex
	rings        map[model.WorkerType]*Ring
	plan         *model.AssignmentPlan
	version      atomic.Int64
	onChange     func()
	moduleRoutes []ModuleRoute
}

func NewDistributor(store *model.Store, cfg *config.Config, logger *slog.Logger) *Distributor {
	return &Distributor{
		store:  store,
		cfg:    cfg,
		rings:  make(map[model.WorkerType]*Ring),
		logger: logger.With("component", "distributor"),
		plan: &model.AssignmentPlan{
			ZoneID:      cfg.ZoneID,
			Assignments: make(map[string]*model.Assignment),
		},
	}
}

func (d *Distributor) SetOnChange(fn func()) {
	d.onChange = fn
}

// SetModuleRoutes 设置模块路由表。由 Coordinator 在所有模块 Init 后调用。
func (d *Distributor) SetModuleRoutes(routes []ModuleRoute) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.moduleRoutes = routes
}

func (d *Distributor) Rebalance() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. 收集所有模块声明的唯一 WorkerType 集合
	workerTypeSet := make(map[model.WorkerType]struct{})
	for _, mr := range d.moduleRoutes {
		for _, r := range mr.Routes {
			workerTypeSet[r.WorkerType] = struct{}{}
		}
	}

	// 2. 为每个 WorkerType 构建哈希环
	d.rings = make(map[model.WorkerType]*Ring, len(workerTypeSet))
	for wt := range workerTypeSet {
		workers := d.store.GetHealthyWorkers(wt)
		ids := make([]string, 0, len(workers))
		for _, w := range workers {
			ids = append(ids, w.ID)
		}
		ring := NewRing(d.cfg.HashPartitions())
		ring.Build(ids)
		d.rings[wt] = ring
	}

	// 3. 构建分配表
	newPlan := &model.AssignmentPlan{
		ZoneID:      d.cfg.ZoneID,
		Assignments: make(map[string]*model.Assignment),
		Version:     d.version.Add(1),
		UpdatedAt:   time.Now(),
	}

	// 确保所有 healthy worker 出现在分配表中
	for wt := range workerTypeSet {
		for _, w := range d.store.GetHealthyWorkers(wt) {
			newPlan.Assignments[w.ID] = &model.Assignment{WorkerID: w.ID}
		}
	}

	// 4. 按模块路由分配 task 和 rule
	for _, mr := range d.moduleRoutes {
		for _, route := range mr.Routes {
			ring := d.rings[route.WorkerType]
			if ring == nil {
				continue
			}

			if route.TaskFilter != nil {
				for _, task := range d.store.GetEnabledTasks() {
					if !route.TaskFilter(task) {
						continue
					}
					workerID := ring.Get(task.TaskID)
					if workerID == "" {
						continue
					}
					a, ok := newPlan.Assignments[workerID]
					if !ok {
						a = &model.Assignment{WorkerID: workerID}
						newPlan.Assignments[workerID] = a
					}
					a.TaskIDs = append(a.TaskIDs, task.TaskID)
				}
			}

			if route.RuleFilter != nil {
				for _, rule := range d.store.GetAllRules() {
					if !route.RuleFilter(rule) {
						continue
					}
					workerID := ring.Get(rule.RuleID)
					if workerID == "" {
						continue
					}
					a, ok := newPlan.Assignments[workerID]
					if !ok {
						a = &model.Assignment{WorkerID: workerID}
						newPlan.Assignments[workerID] = a
					}
					a.RuleIDs = append(a.RuleIDs, rule.RuleID)
				}
			}
		}
	}

	d.plan = newPlan

	// 日志统计
	stats := make(map[string]int)
	for wt := range workerTypeSet {
		stats[string(wt)] = len(d.store.GetHealthyWorkers(wt))
	}
	d.logger.Info("分配表已重算",
		"version", newPlan.Version,
		"workers", len(newPlan.Assignments),
		"worker_types", stats,
		"modules", len(d.moduleRoutes),
	)

	if d.onChange != nil {
		go d.onChange()
	}
}

func (d *Distributor) GetAssignment(workerID string) *model.Assignment {
	d.mu.RLock()
	defer d.mu.RUnlock()
	a, ok := d.plan.Assignments[workerID]
	if !ok {
		return &model.Assignment{WorkerID: workerID}
	}
	return a
}

func (d *Distributor) GetPlan() *model.AssignmentPlan {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.plan
}

// ---------------------------------------------------------------------------
// SSE EventBroker
// ---------------------------------------------------------------------------

type SSEEvent struct {
	Event string
	Data  interface{}
}

type EventBroker struct {
	mu          sync.RWMutex
	subscribers map[chan SSEEvent]struct{}
	logger      *slog.Logger
}

func NewEventBroker(logger *slog.Logger) *EventBroker {
	return &EventBroker{
		subscribers: make(map[chan SSEEvent]struct{}),
		logger:      logger.With("component", "event_broker"),
	}
}

func (b *EventBroker) Subscribe() chan SSEEvent {
	ch := make(chan SSEEvent, 64)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *EventBroker) Unsubscribe(ch chan SSEEvent) {
	b.mu.Lock()
	delete(b.subscribers, ch)
	b.mu.Unlock()
	close(ch)
}

func (b *EventBroker) Broadcast(event SSEEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			b.logger.Warn("subscriber channel full, dropping event", "event", event.Event)
		}
	}
}

func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
