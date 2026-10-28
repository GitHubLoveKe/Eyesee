package health

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"coordinator/config"
	"coordinator/model"
)

type StatusChangeFunc func(workerID string, oldStatus, newStatus model.NodeStatus)

type Checker struct {
	store        *model.Store
	cfg          *config.Config
	client       *http.Client
	onStatusChange StatusChangeFunc
	logger       *slog.Logger
	mu           sync.RWMutex
	lastHeartbeats map[string]time.Time
}

func NewChecker(store *model.Store, cfg *config.Config, onChange StatusChangeFunc, logger *slog.Logger) *Checker {
	return &Checker{
		store: store,
		cfg:   cfg,
		client: &http.Client{
			Timeout: cfg.ProbeTimeout(),
		},
		onStatusChange: onChange,
		logger:         logger.With("component", "health"),
		lastHeartbeats: make(map[string]time.Time),
	}
}

func (c *Checker) Run(ctx context.Context) {
	go c.heartbeatMonitor(ctx)
	go c.activeProber(ctx)
	c.logger.Info("health checker started",
		"probe_interval", c.cfg.ProbeInterval(),
		"heartbeat_timeout", c.cfg.HeartbeatTimeout(),
	)
}

func (c *Checker) heartbeatMonitor(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.ProbeInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.checkHeartbeatTimeouts()
		}
	}
}

func (c *Checker) checkHeartbeatTimeouts() {
	timeout := c.cfg.HeartbeatTimeout()
	now := time.Now()

	workers := c.store.GetAllWorkers()
	for _, w := range workers {
		c.mu.RLock()
		lastHB, ok := c.lastHeartbeats[w.ID]
		c.mu.RUnlock()

		if !ok {
			continue
		}

		if now.Sub(lastHB) > timeout && w.Status != model.StatusOffline {
			old := w.Status
			w.Status = model.StatusOffline
			c.logger.Warn("worker 心跳超时, 标记 offline", "worker_id", w.ID, "timeout", timeout)
			c.fireStatusChange(w.ID, old, w.Status)
		}
	}
}

func (c *Checker) activeProber(ctx context.Context) {
	ticker := time.NewTicker(c.cfg.ProbeInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.probeAllWorkers()
		}
	}
}

func (c *Checker) probeAllWorkers() {
	workers := c.store.GetAllWorkers()
	var wg sync.WaitGroup
	for _, w := range workers {
		wg.Add(1)
		go func(w *model.Worker) {
			defer wg.Done()
			c.probeWorker(w)
		}(w)
	}
	wg.Wait()
}

func (c *Checker) probeWorker(w *model.Worker) {
	url := fmt.Sprintf("%s%s", w.Address, w.HealthEndpoint)

	resp, err := c.client.Get(url)
	if err != nil {
		w.ProbeFailCount++
		c.logger.Debug("worker 探测失败", "worker_id", w.ID, "url", url, "fail_count", w.ProbeFailCount, "error", err)

		if w.ProbeFailCount >= c.cfg.Health.FailThreshold {
			c.transitionStatus(w, model.StatusDegraded)
		}
		return
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		w.ProbeFailCount = 0
		c.transitionStatus(w, model.StatusHealthy)
	} else {
		w.ProbeFailCount++
		if w.ProbeFailCount >= c.cfg.Health.FailThreshold {
			c.transitionStatus(w, model.StatusDegraded)
		}
	}
}

func (c *Checker) transitionStatus(w *model.Worker, newStatus model.NodeStatus) {
	if w.Status == newStatus {
		return
	}
	old := w.Status
	w.Status = newStatus
	c.fireStatusChange(w.ID, old, newStatus)
}

func (c *Checker) fireStatusChange(workerID string, old, new model.NodeStatus) {
	if c.onStatusChange != nil && old != new {
		c.onStatusChange(workerID, old, new)
	}
}

func (c *Checker) RecordHeartbeat(workerID string) {
	c.mu.Lock()
	c.lastHeartbeats[workerID] = time.Now()
	c.mu.Unlock()

	w := c.store.GetWorker(workerID)
	if w != nil {
		w.LastHeartbeat = time.Now()
		if w.Status == model.StatusOffline || w.Status == model.StatusUnknown {
			c.transitionStatus(w, model.StatusHealthy)
		}
	}
}

func (c *Checker) ProbeResult(workerID string) (model.NodeStatus, int) {
	w := c.store.GetWorker(workerID)
	if w == nil {
		return model.StatusUnknown, 0
	}
	return w.Status, w.ProbeFailCount
}
