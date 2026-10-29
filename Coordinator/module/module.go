package module

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"coordinator/config"
	"coordinator/model"
)

// Module 可插拔功能模块的接口契约。
// 每个模块通过 init() 自注册到全局 Registry，由 Coordinator 按 config 启用。
type Module interface {
	Name() string
	Init(ctx ModuleContext) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Routes 返回该模块的任务路由规则：每种 WorkerType 对应一组 task/rule 过滤器。
	// Distributor 在 Rebalance 时据此构建哈希环并分配任务。
	Routes() []TaskRoute
}

// TaskRoute 定义一条路由规则：匹配的任务/规则分配给指定 WorkerType。
type TaskRoute struct {
	WorkerType model.WorkerType
	TaskFilter func(*model.CollectionTask) bool // nil = 该路由不处理 task
	RuleFilter func(*model.AlertRule) bool       // nil = 该路由不处理 rule
}

// ConfigGenerator 可选接口：模块为 worker 生成下发配置。
// 实现此接口的模块可通过 API 向 worker 推送配置。
type ConfigGenerator interface {
	GenerateWorkerConfig(workerID string, assignment *model.Assignment) ([]byte, error)
	ConfigContentType() string // e.g. "application/x-yaml"
}

// RouteRegistrar 可选接口：模块注册自定义 HTTP 路由到 API Server。
type RouteRegistrar interface {
	RegisterRoutes(registry RouteRegistry)
}

// RouteRegistry API Server 暴露给模块的路由注册句柄。
type RouteRegistry interface {
	Handle(pattern string, handler http.HandlerFunc)
}

// ModuleContext 模块初始化时收到的共享依赖。
// 注意：不包含 Distributor/Broker，避免循环依赖。
type ModuleContext struct {
	Store     *model.Store
	Config    *config.Config
	RawConfig map[string]interface{} // 该模块在 config.yaml 中的专属配置段
	Logger    *slog.Logger
}

// ---------------------------------------------------------------------------
// 全局注册表
// ---------------------------------------------------------------------------

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Module)
)

// Register 将模块注册到全局注册表。通常在模块包的 init() 中调用。
// 重复注册同名模块会 panic。
func Register(m Module) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, dup := registry[m.Name()]; dup {
		panic("module: duplicate registration: " + m.Name())
	}
	registry[m.Name()] = m
}

// Get 按名称查找已注册模块。
func Get(name string) (Module, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	m, ok := registry[name]
	return m, ok
}

// All 返回所有已注册模块的快照。
func All() map[string]Module {
	registryMu.RLock()
	defer registryMu.RUnlock()
	snap := make(map[string]Module, len(registry))
	for k, v := range registry {
		snap[k] = v
	}
	return snap
}
