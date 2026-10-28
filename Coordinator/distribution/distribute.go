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

	"gopkg.in/yaml.v3"
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

type Distributor struct {
	store  *model.Store
	cfg    *config.Config
	ring   *Ring // 保留用于兼容，实际使用下面两个专用 ring
	vmagentRing  *Ring
	vmalertRing  *Ring
	logger *slog.Logger

	mu       sync.RWMutex
	plan     *model.AssignmentPlan
	version  atomic.Int64
	onChange func()
}

func NewDistributor(store *model.Store, cfg *config.Config, logger *slog.Logger) *Distributor {
	partitions := cfg.HashPartitions()
	d := &Distributor{
		store:       store,
		cfg:         cfg,
		ring:        NewRing(partitions),
		vmagentRing: NewRing(partitions),
		vmalertRing: NewRing(partitions),
		logger:      logger.With("component", "distributor"),
		plan: &model.AssignmentPlan{
			ZoneID:      cfg.ZoneID,
			Assignments: make(map[string]*model.Assignment),
		},
	}
	return d
}

func (d *Distributor) SetOnChange(fn func()) {
	d.onChange = fn
}

func (d *Distributor) Rebalance() {
	d.mu.Lock()
	defer d.mu.Unlock()

	vmagentWorkers := d.store.GetHealthyWorkers(model.WorkerVMAgent)
	vmalertWorkers := d.store.GetHealthyWorkers(model.WorkerVMAlert)

	// 为 vmagent 和 vmalert 分别构建独立的哈希环
	vmagentIDs := make([]string, 0, len(vmagentWorkers))
	for _, w := range vmagentWorkers {
		vmagentIDs = append(vmagentIDs, w.ID)
	}
	d.vmagentRing.Build(vmagentIDs)

	vmalertIDs := make([]string, 0, len(vmalertWorkers))
	for _, w := range vmalertWorkers {
		vmalertIDs = append(vmalertIDs, w.ID)
	}
	d.vmalertRing.Build(vmalertIDs)

	newPlan := &model.AssignmentPlan{
		ZoneID:      d.cfg.ZoneID,
		Assignments: make(map[string]*model.Assignment),
		Version:     d.version.Add(1),
		UpdatedAt:   time.Now(),
	}

	for _, w := range vmagentWorkers {
		newPlan.Assignments[w.ID] = &model.Assignment{WorkerID: w.ID}
	}
	for _, w := range vmalertWorkers {
		newPlan.Assignments[w.ID] = &model.Assignment{WorkerID: w.ID}
	}

	// 采集任务只分配给 vmagent workers
	tasks := d.store.GetEnabledTasks()
	for _, task := range tasks {
		workerID := d.vmagentRing.Get(task.TaskID)
		if workerID == "" {
			continue
		}
		assignment, ok := newPlan.Assignments[workerID]
		if !ok {
			assignment = &model.Assignment{WorkerID: workerID}
			newPlan.Assignments[workerID] = assignment
		}
		assignment.TaskIDs = append(assignment.TaskIDs, task.TaskID)
	}

	// 告警规则只分配给 vmalert workers
	rules := d.store.GetAllRules()
	for _, rule := range rules {
		workerID := d.vmalertRing.Get(rule.RuleID)
		if workerID == "" {
			continue
		}
		assignment, ok := newPlan.Assignments[workerID]
		if !ok {
			assignment = &model.Assignment{WorkerID: workerID}
			newPlan.Assignments[workerID] = assignment
		}
		assignment.RuleIDs = append(assignment.RuleIDs, rule.RuleID)
	}

	d.plan = newPlan
	d.logger.Info("分配表已重算",
		"version", newPlan.Version,
		"workers", len(newPlan.Assignments),
		"vmagent_workers", len(vmagentWorkers),
		"vmalert_workers", len(vmalertWorkers),
		"tasks", len(tasks),
		"rules", len(rules),
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

func (d *Distributor) GenerateScrapeConfig(workerID string) ([]byte, error) {
	assignment := d.GetAssignment(workerID)
	tasks := d.store.GetEnabledTasks()

	taskSet := make(map[string]bool, len(assignment.TaskIDs))
	for _, id := range assignment.TaskIDs {
		taskSet[id] = true
	}

	var scrapeConfigs []map[string]interface{}
	for _, task := range tasks {
		if !taskSet[task.TaskID] {
			continue
		}
		sc := d.buildScrapeConfig(task)
		scrapeConfigs = append(scrapeConfigs, sc)
	}

	w := d.store.GetWorker(workerID)
	if w == nil {
		return yaml.Marshal(map[string]interface{}{"scrape_configs": scrapeConfigs})
	}

	zoneResources := d.store.GetResourcesByType(model.ResourceVM)
	var remoteWriteURL string
	if len(zoneResources) > 0 {
		remoteWriteURL = zoneResources[0].Address + "/api/v1/write"
	}

	result := map[string]interface{}{
		"scrape_configs": scrapeConfigs,
	}
	if remoteWriteURL != "" {
		result["remote_write"] = []map[string]interface{}{
			{"url": remoteWriteURL},
		}
	}

	return yaml.Marshal(result)
}

func (d *Distributor) buildScrapeConfig(task *model.CollectionTask) map[string]interface{} {
	sc := map[string]interface{}{
		"job_name": task.JobName,
	}

	if task.ScrapeInterval != "" {
		sc["scrape_interval"] = task.ScrapeInterval
	}
	if task.MetricsPath != "" && task.MetricsPath != "/metrics" {
		sc["metrics_path"] = task.MetricsPath
	}

	labels := make(map[string]string)
	for k, v := range task.Labels {
		labels[k] = v
	}
	labels["zone"] = task.ZoneID
	labels["__meta_task_id"] = task.TaskID

	staticConfig := map[string]interface{}{
		"targets": []string{task.Target},
		"labels":  labels,
	}
	sc["static_configs"] = []interface{}{staticConfig}

	if task.Auth != nil && task.Auth.Username != "" {
		sc["basic_auth"] = map[string]interface{}{
			"username": task.Auth.Username,
			"password": task.Auth.Password,
		}
	}

	return sc
}

func (d *Distributor) GenerateAlertRules(workerID string) ([]byte, error) {
	assignment := d.GetAssignment(workerID)
	allRules := d.store.GetAllRules()

	ruleSet := make(map[string]bool, len(assignment.RuleIDs))
	for _, id := range assignment.RuleIDs {
		ruleSet[id] = true
	}

	groupMap := make(map[string][]map[string]interface{})

	for _, rule := range allRules {
		if !ruleSet[rule.RuleID] {
			continue
		}

		groupName := rule.GroupName
		if groupName == "" {
			groupName = fmt.Sprintf("zone-%s-default", rule.ZoneID)
		}

		ruleDef := map[string]interface{}{
			"alert": rule.AlertName,
			"expr":  rule.Expr,
		}
		if rule.Duration != "" {
			ruleDef["for"] = rule.Duration
		}

		allLabels := make(map[string]string)
		allLabels["zone"] = rule.ZoneID
		if rule.Severity != "" {
			allLabels["severity"] = rule.Severity
		}
		for k, v := range rule.Labels {
			allLabels[k] = v
		}
		ruleDef["labels"] = allLabels

		if len(rule.Annotations) > 0 {
			ruleDef["annotations"] = rule.Annotations
		}

		groupMap[groupName] = append(groupMap[groupName], ruleDef)
	}

	var groups []map[string]interface{}
	for name, rules := range groupMap {
		group := map[string]interface{}{
			"name":  name,
			"rules": rules,
		}
		groups = append(groups, group)
	}

	result := map[string]interface{}{
		"groups": groups,
	}

	return yaml.Marshal(result)
}

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
