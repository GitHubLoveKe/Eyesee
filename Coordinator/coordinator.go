package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"coordinator/api"
	"coordinator/config"
	"coordinator/distribution"
	"coordinator/gossip"
	"coordinator/health"
	"coordinator/model"
	"coordinator/module"
)

type Coordinator struct {
	cfg         *config.Config
	store       *model.Store
	distributor *distribution.Distributor
	broker      *distribution.EventBroker
	healthChk   *health.Checker
	gossip      *gossip.GossipCoordinator
	apiServer   *api.Server
	logger      *slog.Logger
	httpClient  *http.Client

	activeModules []module.Module
}

func New(cfg *config.Config, logger *slog.Logger) *Coordinator {
	store := model.NewStore(cfg.ZoneID)
	broker := distribution.NewEventBroker(logger)
	dist := distribution.NewDistributor(store, cfg, logger)

	c := &Coordinator{
		cfg:         cfg,
		store:       store,
		broker:      broker,
		distributor: dist,
		logger:      logger.With("component", "coordinator"),
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}

	// 初始化启用的模块
	c.activeModules = c.initModules(cfg, store, logger)

	// 构建模块路由表传给 Distributor
	distRoutes := make([]distribution.ModuleRoute, 0, len(c.activeModules))
	for _, m := range c.activeModules {
		routes := m.Routes()
		dr := distribution.ModuleRoute{
			ModuleName: m.Name(),
			Routes:     make([]distribution.TaskRoute, len(routes)),
		}
		for i, r := range routes {
			dr.Routes[i] = distribution.TaskRoute{
				WorkerType: r.WorkerType,
				TaskFilter: r.TaskFilter,
				RuleFilter: r.RuleFilter,
			}
		}
		distRoutes = append(distRoutes, dr)
	}
	dist.SetModuleRoutes(distRoutes)

	// 构建现有子系统
	c.healthChk = health.NewChecker(store, cfg, c.onWorkerStatusChange, logger)
	c.gossip = gossip.New(cfg, &gossip.DelegateCallbacks{
		OnWorkerUpdate: c.onGossipWorkerUpdate,
		OnRebalanceReq: c.onGossipRebalance,
	}, logger)
	c.apiServer = api.NewServer(cfg, store, dist, broker, c.healthChk, c.activeModules, logger)

	dist.SetOnChange(func() {
		c.broker.Broadcast(distribution.SSEEvent{
			Event: "rebalance_complete",
			Data: map[string]interface{}{
				"version": c.distributor.GetPlan().Version,
			},
		})
	})

	return c
}

func (c *Coordinator) initModules(cfg *config.Config, store *model.Store, logger *slog.Logger) []module.Module {
	var active []module.Module
	for _, name := range cfg.EnabledModules {
		m, ok := module.Get(name)
		if !ok {
			logger.Warn("未知模块, 跳过", "name", name)
			continue
		}

		var rawCfg map[string]interface{}
		if cfg.ModuleConfigs != nil {
			if rc, ok := cfg.ModuleConfigs[name]; ok {
				rawCfg, _ = rc.(map[string]interface{})
			}
		}

		mctx := module.ModuleContext{
			Store:     store,
			Config:    cfg,
			RawConfig: rawCfg,
			Logger:    logger.With("module", name),
		}

		if err := m.Init(mctx); err != nil {
			logger.Error("模块初始化失败", "name", name, "error", err)
			continue
		}

		active = append(active, m)
		logger.Info("模块已加载", "name", name)
	}
	return active
}

func (c *Coordinator) Run(ctx context.Context) error {
	c.logger.Info("Zone Coordinator 启动中",
		"zone_id", c.cfg.ZoneID,
		"http_addr", c.cfg.Server.HTTPAddr,
		"gossip_port", c.cfg.Gossip.BindPort,
		"modules", c.moduleNames(),
	)

	// 1. 绑定 HTTP 监听端口（同步，失败立即退出）
	ln, err := c.apiServer.Listen()
	if err != nil {
		return fmt.Errorf("HTTP 监听失败: %w", err)
	}
	c.logger.Info("HTTP API 端口已绑定", "addr", ln.Addr().String())

	// 2. 启动 gossip 协议
	if err := c.gossip.Start(ctx); err != nil {
		c.logger.Error("gossip 启动失败, 将以单节点模式运行", "error", err)
	} else {
		c.logger.Info("gossip 协议已启动")
	}

	// 3. 启动 HTTP 服务（在已绑定的 listener 上）
	go c.apiServer.Serve(ctx, ln)

	// 4. 启动健康检测
	c.healthChk.Run(ctx)
	c.logger.Info("健康检测已启动")

	// 5. 启动模块
	for _, m := range c.activeModules {
		if err := m.Start(ctx); err != nil {
			c.logger.Error("模块启动失败", "name", m.Name(), "error", err)
		}
	}

	// 6. 启动资源同步
	go c.resourceSyncLoop(ctx)
	c.logger.Info("资源同步循环已启动", "interval", c.cfg.SyncInterval())

	// 等待退出信号
	<-ctx.Done()
	c.logger.Info("Zone Coordinator 正在关闭...")

	// 逆序停止模块
	for i := len(c.activeModules) - 1; i >= 0; i-- {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		c.activeModules[i].Stop(shutdownCtx)
		cancel()
	}

	c.gossip.Stop()
	return nil
}

func (c *Coordinator) moduleNames() []string {
	names := make([]string, len(c.activeModules))
	for i, m := range c.activeModules {
		names[i] = m.Name()
	}
	return names
}

func (c *Coordinator) resourceSyncLoop(ctx context.Context) {
	c.pullResources(ctx)

	ticker := time.NewTicker(c.cfg.SyncInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.pullResources(ctx)
		}
	}
}

func (c *Coordinator) pullResources(ctx context.Context) {
	if c.cfg.Management.URL == "" {
		return
	}

	url := fmt.Sprintf("%s/api/v1/zones/%s/resources", c.cfg.Management.URL, c.cfg.ZoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Warn("构建资源拉取请求失败", "error", err)
		return
	}
	if c.cfg.Management.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.Management.AuthToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Debug("资源拉取失败, 等待下次重试", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("资源拉取返回非 200", "status", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Warn("读取资源响应体失败", "error", err)
		return
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			Resources []*model.Resource       `json:"resources"`
			Tasks     []*model.CollectionTask `json:"tasks"`
			Rules     []*model.AlertRule      `json:"rules"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Warn("解析资源响应失败", "error", err)
		return
	}

	if result.Code != 0 {
		c.logger.Warn("管理平台返回错误码", "code", result.Code)
		return
	}

	if result.Data.Resources != nil {
		c.store.ReplaceResources(result.Data.Resources)
	}
	if result.Data.Tasks != nil {
		c.store.ReplaceTasks(result.Data.Tasks)
	}
	if result.Data.Rules != nil {
		c.store.ReplaceRules(result.Data.Rules)
	}

	c.distributor.Rebalance()
	c.logger.Info("资源同步完成",
		"resources", len(result.Data.Resources),
		"tasks", len(result.Data.Tasks),
		"rules", len(result.Data.Rules),
	)
}

func (c *Coordinator) onWorkerStatusChange(workerID string, oldStatus, newStatus model.NodeStatus) {
	c.logger.Info("worker 状态变更",
		"worker_id", workerID,
		"old", oldStatus,
		"new", newStatus,
	)

	if newStatus == model.StatusHealthy || oldStatus == model.StatusHealthy || oldStatus == model.StatusDegraded {
		debounce := c.cfg.RebalanceDebounce()
		time.AfterFunc(debounce, func() {
			c.distributor.Rebalance()
		})
	}

	c.reportHealthToManagement()
}

func (c *Coordinator) onGossipWorkerUpdate(update gossip.WorkerUpdate) {
	c.logger.Debug("收到 gossip worker 更新", "worker_id", update.WorkerID, "action", update.Action)

	if update.Action == "remove" {
		c.store.RemoveWorker(update.WorkerID)
		c.distributor.Rebalance()
	} else if update.Status != "" {
		w := c.store.GetWorker(update.WorkerID)
		if w != nil {
			w.Status = model.NodeStatus(update.Status)
		}
	}
}

func (c *Coordinator) onGossipRebalance(reason string) {
	c.logger.Info("收到 gossip rebalance 请求", "reason", reason)
	c.distributor.Rebalance()
}

func (c *Coordinator) reportHealthToManagement() {
	if c.cfg.Management.URL == "" {
		return
	}

	workers := c.store.GetAllWorkers()
	resources := c.store.GetResources()

	type healthEntry struct {
		ID     string          `json:"id"`
		Type   string          `json:"type"`
		Status model.NodeStatus `json:"status"`
	}

	var entries []healthEntry
	for _, w := range workers {
		entries = append(entries, healthEntry{
			ID:     w.ID,
			Type:   string(w.Type),
			Status: w.Status,
		})
	}
	for _, r := range resources {
		entries = append(entries, healthEntry{
			ID:     r.ID,
			Type:   string(r.Type),
			Status: r.Status,
		})
	}

	payload := map[string]interface{}{
		"zone_id":   c.cfg.ZoneID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"entries":   entries,
	}

	data, _ := json.Marshal(payload)
	reportURL := fmt.Sprintf("%s/api/v1/zones/%s/health-report", c.cfg.Management.URL, c.cfg.ZoneID)

	req, err := http.NewRequest(http.MethodPost, reportURL, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.Management.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.Management.AuthToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Debug("健康上报失败", "error", err)
		return
	}
	resp.Body.Close()
}
