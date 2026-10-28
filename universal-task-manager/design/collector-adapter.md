# 采集器适配器设计

> 描述采集器兼容层的详细设计，支持 Alloy、vmagent、Prometheus 等采集器的统一接入。

---

## 一、设计目标

| 目标 | 描述 |
|------|------|
| 统一抽象 | 不同采集器通过统一接口接入 |
| 配置转换 | 任务配置自动转换为采集器原生格式 |
| 热更新 | 配置变更无需重启采集器 |
| 可扩展 | 新增采集器只需实现适配器接口 |

---

## 二、适配器接口

```go
// CollectorAdapter 采集器适配器
type CollectorAdapter interface {
    // Type 采集器类型标识
    Type() string

    // ValidateConfig 校验任务配置
    ValidateConfig(task *CollectionTask, source *DataSource) error

    // GenerateConfig 生成采集器原生配置
    // 输入: 任务 + 实例 + 任务类型定义
    // 输出: 采集器原生配置字节流
    GenerateConfig(ctx *ConfigContext) ([]byte, error)

    // GenerateSD 生成服务发现配置
    // 用于 http_sd / file_sd 场景
    GenerateSD(tasks []*CollectionTask) ([]byte, error)

    // ApplyConfig 将配置应用到采集器
    // 支持热更新的采集器可直接推送
    ApplyConfig(zone *Zone, config []byte) error

    // HealthCheck 检查采集器健康状态
    HealthCheck(zone *Zone) (*HealthStatus, error)
}

// ConfigContext 配置生成上下文
type ConfigContext struct {
    Task     *CollectionTask
    Source   *DataSource
    TaskType *TaskType
    Zone     *Zone
}
```

---

## 三、内置适配器实现

### 3.1 vmagent 适配器

vmagent 使用 Prometheus 兼容的 scrape_config 格式。

```go
type VmagentAdapter struct{}

func (a *VmagentAdapter) GenerateConfig(ctx *ConfigContext) ([]byte, error) {
    // 从任务类型的 config_template.vmagent 获取模板
    template := ctx.TaskType.ConfigTemplate["vmagent"]

    // 渲染模板变量
    rendered := renderTemplate(template, map[string]interface{}{
        "task_id":     ctx.Task.TaskID,
        "source_id":   ctx.Source.SourceID,
        "host":        ctx.Source.Host,
        "port":        ctx.Source.Port,
        "job_name":    fmt.Sprintf("%s-%s", ctx.TaskType.TypeID, ctx.Source.SourceID),
        // 任务表单中的配置值
        "config":      ctx.Task.Config,
    })

    return json.Marshal(rendered)
}

func (a *VmagentAdapter) GenerateSD(tasks []*CollectionTask) ([]byte, error) {
    // 生成 http_sd 格式的 target 列表
    var targets []httpSDTarget
    for _, task := range tasks {
        targets = append(targets, httpSDTarget{
            Targets: []string{fmt.Sprintf("%s:%d", task.Host, task.Port)},
            Labels: map[string]string{
                "job":        task.Name,
                "instance":   task.SourceID,
                "__meta_task_id": task.TaskID,
            },
        })
    }
    return json.Marshal(targets)
}

func (a *VmagentAdapter) ApplyConfig(zone *Zone, config []byte) error {
    // 通过 vmagent API 热加载配置
    // POST /-/reload 或写入配置目录
    return http.Post(zone.RemoteWriteURL+"/-/reload", "application/json", bytes.NewReader(config))
}
```

**vmagent 配置输出示例：**

```json
{
  "job_name": "mysql-standard-inst-001",
  "scrape_interval": "15s",
  "scrape_timeout": "10s",
  "metrics_path": "/metrics",
  "static_configs": [
    {
      "targets": ["10.0.1.5:9104"],
      "labels": {
        "instance": "inst-001",
        "zone": "zone-east-1"
      }
    }
  ],
  "basic_auth": {
    "username": "monitor",
    "password": "***"
  }
}
```

---

### 3.2 Alloy 适配器

Alloy 使用 River 配置语言。

```go
type AlloyAdapter struct{}

func (a *AlloyAdapter) GenerateConfig(ctx *ConfigContext) ([]byte, error) {
    // 从任务类型的 config_template.alloy 获取模板
    template := ctx.TaskType.ConfigTemplate["alloy"]

    // 生成 River 格式配置
    river := generateRiverConfig(template, ctx)

    return []byte(river), nil
}

func (a *AlloyAdapter) ApplyConfig(zone *Zone, config []byte) error {
    // 通过 Alloy API 热更新
    // PUT /-/config
    return http.Put(zone.AlloyURL+"/-/config", "application/json", bytes.NewReader(config))
}
```

**Alloy 配置输出示例：**

```river
prometheus.scrape "mysql_standard_inst_001" {
  job_name        = "mysql-standard-inst-001"
  scrape_interval = "15s"
  targets         = [{
    __address__ = "10.0.1.5:9104",
  }]

  forward_to = [prometheus.remote_write.default.receiver]
}

prometheus.remote_write "default" {
  endpoint {
    url = "http://vmagent:8429/api/v1/write"
  }
}
```

---

### 3.3 Prometheus 适配器

Prometheus 使用 YAML 格式的 scrape_config。

```go
type PrometheusAdapter struct{}

func (a *PrometheusAdapter) GenerateConfig(ctx *ConfigContext) ([]byte, error) {
    // 生成 Prometheus 原生 YAML 配置
    scrapeConfig := &ScrapeConfig{
        JobName:        fmt.Sprintf("%s-%s", ctx.TaskType.TypeID, ctx.Source.SourceID),
        ScrapeInterval: ctx.Task.Config["scrape_interval"].(string),
        ScrapeTimeout:  "10s",
        MetricsPath:    "/metrics",
        StaticConfigs: []StaticConfig{
            {
                Targets: []string{fmt.Sprintf("%s:%d", ctx.Source.Host, ctx.Source.Port)},
                Labels: map[string]string{
                    "instance": ctx.Source.SourceID,
                    "zone":     ctx.Zone.ZoneID,
                },
            },
        },
    }

    return yaml.Marshal(scrapeConfig)
}

func (a *PrometheusAdapter) ApplyConfig(zone *Zone, config []byte) error {
    // 写入配置文件并触发 reload
    // 1. 写入 /etc/prometheus/scrape_configs/{task_id}.yaml
    // 2. POST /-/reload
    return writeAndReload(zone, config)
}
```

---

## 四、服务发现集成

### 4.1 http_sd 模式

适用于 vmagent 和 Alloy，通过 HTTP 接口动态获取 target 列表。

```
采集器 → GET /api/v1/sd/{zone_id}/targets → 管理平台 → 返回 target 列表
```

**管理平台 http_sd 接口：**

```go
func (s *Server) handleHTTPSD(c *gin.Context) {
    zoneID := c.Param("zone_id")

    // 查询该区域所有启用的任务
    tasks, _ := s.taskRepo.FindEnabledByZone(zoneID)

    // 转换为 http_sd 格式
    var targets []httpSDTarget
    for _, task := range tasks {
        source, _ := s.sourceRepo.FindByID(task.SourceID)
        taskType, _ := s.taskTypeRepo.FindByID(task.TypeID)

        targets = append(targets, httpSDTarget{
            Targets: []string{fmt.Sprintf("%s:%d", source.Host, source.Port)},
            Labels: map[string]string{
                "job":              task.Name,
                "instance":         source.SourceID,
                "source_type":      source.SourceType,
                "__meta_task_id":   task.TaskID,
                "__meta_type_id":   task.TypeID,
            },
        })
    }

    c.JSON(200, targets)
}
```

**http_sd 响应格式：**

```json
[
  {
    "targets": ["10.0.1.5:9104"],
    "labels": {
      "job": "mysql-standard-inst-001",
      "instance": "inst-001",
      "source_type": "mysql",
      "__meta_task_id": "task-001",
      "__meta_type_id": "mysql-standard"
    }
  },
  {
    "targets": ["10.0.1.6:9100"],
    "labels": {
      "job": "linux-node-inst-002",
      "instance": "inst-002",
      "source_type": "linux",
      "__meta_task_id": "task-002",
      "__meta_type_id": "linux-node"
    }
  }
]
```

### 4.2 file_sd 模式

适用于 Prometheus，写入文件由 Prometheus 定期读取。

```go
func (s *Server) generateFileSD(zoneID string) error {
    tasks, _ := s.taskRepo.FindEnabledByZone(zoneID)

    var targets []fileSDTarget
    for _, task := range tasks {
        // ... 构建 target
    }

    // 写入文件
    path := fmt.Sprintf("/etc/prometheus/sd/%s.json", zoneID)
    data, _ := json.Marshal(targets)
    return os.WriteFile(path, data, 0644)
}
```

---

## 五、配置热更新

### 5.1 更新流程

```
任务配置变更
    │
    ▼
管理平台检测变更
    │
    ├── 查询任务绑定的采集器类型
    │
    ▼
调用对应适配器 GenerateConfig
    │
    ▼
调用适配器 ApplyConfig
    │
    ├── vmagent: POST /-/reload
    ├── Alloy:   PUT /-/config
    └── Prometheus: 写入文件 + POST /-/reload
    │
    ▼
采集器热加载新配置（不中断采集）
```

### 5.2 批量更新优化

当多个任务同时变更时，合并为一次采集器 reload：

```go
func (s *Server) batchApplyChanges(zoneID string) error {
    // 收集该区域所有变更的任务
    tasks := s.changeBuffer.Drain(zoneID)

    // 按采集器类型分组
    byCollector := groupByCollector(tasks)

    for collectorType, collectorTasks := range byCollector {
        adapter := s.collectorFactory.Get(collectorType)

        // 生成完整配置（而非增量）
        fullConfig := adapter.GenerateFullConfig(zoneID, collectorTasks)

        // 一次 apply
        adapter.ApplyConfig(zone, fullConfig)
    }

    return nil
}
```

---

## 六、健康检查

```go
type HealthStatus struct {
    CollectorType string
    Status        string // healthy / degraded / down
    Uptime        time.Duration
    TargetCount   int
    LastReload    time.Time
    Errors        []string
}

func (a *VmagentAdapter) HealthCheck(zone *Zone) (*HealthStatus, error) {
    // GET /api/v1/status/buildinfo
    // GET /api/v1/targets
    resp, err := http.Get(zone.VmagentURL + "/api/v1/status/buildinfo")
    if err != nil {
        return &HealthStatus{Status: "down"}, err
    }

    // 解析响应，构建健康状态
    return &HealthStatus{
        CollectorType: "vmagent",
        Status:        "healthy",
        Uptime:        parseUptime(resp),
        TargetCount:   countTargets(zone),
        LastReload:    parseLastReload(resp),
    }, nil
}
```

---

## 七、扩展新采集器

### 7.1 实现步骤

1. **创建适配器结构体**

```go
type MyCollectorAdapter struct{}
```

2. **实现接口方法**

```go
func (a *MyCollectorAdapter) Type() string { return "my-collector" }
func (a *MyCollectorAdapter) ValidateConfig(task *CollectionTask, source *DataSource) error { ... }
func (a *MyCollectorAdapter) GenerateConfig(ctx *ConfigContext) ([]byte, error) { ... }
func (a *MyCollectorAdapter) GenerateSD(tasks []*CollectionTask) ([]byte, error) { ... }
func (a *MyCollectorAdapter) ApplyConfig(zone *Zone, config []byte) error { ... }
func (a *MyCollectorAdapter) HealthCheck(zone *Zone) (*HealthStatus, error) { ... }
```

3. **注册到工厂**

```go
func init() {
    collectorFactory.Register("my-collector", &MyCollectorAdapter{})
}
```

4. **在任务类型中添加配置模板**

```json
{
  "config_template": {
    "my-collector": {
      "custom_field": "{{config.my_option}}"
    }
  }
}
```

---

## 八、错误处理

| 场景 | 处理策略 |
|------|----------|
| 采集器不可达 | 记录错误，标记区域状态为 degraded，继续重试 |
| 配置生成失败 | 回滚到上一版本配置，告警通知 |
| 热更新失败 | 重试 3 次，失败后标记任务状态为 error |
| 采集器类型不支持 | 返回明确错误，提示用户选择支持的采集器 |
