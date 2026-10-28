# A4 告警收敛模块

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：A4 | 职责：提供告警引擎的可替换抽象（AlertEngine），管理告警事件的消费、落库、收敛、认领与静默，支持默认实现与多种替换路径。

---

## 一、概述

告警收敛模块是告警层的引擎抽象层。它定义了 `AlertEngine` 接口，将告警规则的推送/撤回、告警事件的查询、告警认领与静默等操作抽象为统一契约。默认实现基于 vmalert + alertmanager + Kafka 的流水线，但整个引擎层可替换为 Prometheus rules / Thanos Ruler / Cortex 等方案。

模块同时负责告警事件（alert_event）的 Kafka 消费与落库，确保所有告警事件有持久化记录，支持历史查询与审计。

**默认告警流水线（两级 Alertmanager）：**

```
vmalert × N（各区，规则评估）
    │
    ▼
区AM（每区单实例，轻量去重：fingerprint消重）
    │
    ▼
核心AM（中心部署，策略路由：分组/抑制/静默/升级/分发）
    │
    ▼
Kafka（消息队列）
    │
    ▼
消费者（告警事件落库 → alert_event 表）
    │
    ▼
NotificationDispatcher（通知投递）
```

**两级 AM 职责划分：**

| 层级 | 组件 | 部署 | 职责 | 不做什么 |
|------|------|------|------|----------|
| 区级 | alertmanager | 每区单实例，不集群 | fingerprint 去重，消除同区多 vmalert 的重复告警 | 不配置 group/inhibit/silence |
| 核心 | alertmanager | 中心部署，可集群 | 全局分组(group_by)、抑制(inhibit)、静默(silence)、路由分发 | 不做 fingerprint 级去重 |

> 区 AM 保持轻量——只转发去重后的告警到核心 AM。核心 AM 拥有全局视图，可跨区关联告警并执行策略。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| AlertEngine 接口定义 | 定义告警引擎的抽象契约 |
| 规则推送与撤回 | 通过 AlertEngine.PushRules / RemoveRules 管理告警规则（推送到区 vmalert） |
| 两级 AM 管理 | 区 AM 配置去重转发；核心 AM 配置分组/抑制/静默策略 |
| Kafka 消费与落库 | 从核心 AM 输出的 Kafka 消费告警事件，写入 alert_event 表 |
| 告警事件查询 | 提供告警事件的多维查询能力 |
| 告警认领 | 支持用户对 firing 状态的告警进行认领 |
| 告警静默 | 支持对指定告警设置静默期（核心 AM 级别） |
| 默认引擎实现 | 提供 vmalert + 区AM(去重) + 核心AM(策略) 的默认适配 |
| 替换路径文档 | 记录各组件的替换方案 |

### 本模块不负责

| 职责 | 归属 |
|------|------|
| 告警策略的定义 | A1 告警策略模块 |
| 告警路由与通知投递 | A2 告警组模块 / A3 通知渠道模块 |
| 告警规则的评估执行 | 底层告警引擎（vmalert / Prometheus 等） |
| 告警的分组/抑制/静默（引擎侧） | alertmanager 或替换组件 |

---

## 三、功能清单

| 功能编号 | 功能名称 | 优先级 | 说明 |
|----------|----------|--------|------|
| A4-F01 | AlertEngine 接口定义 | P0 | 定义引擎抽象接口 |
| A4-F02 | 规则推送（PushRules） | P0 | 将告警规则推送至指定网区的告警引擎 |
| A4-F03 | 规则撤回（RemoveRules） | P0 | 从告警引擎撤回指定规则 |
| A4-F04 | 告警查询（QueryAlerts） | P0 | 多维查询告警事件 |
| A4-F05 | 告警认领（Acknowledge） | P0 | 用户对 firing 告警进行认领 |
| A4-F06 | 告警静默（Silence） | P0 | 对指定告警设置静默期 |
| A4-F07 | Kafka 消费 | P0 | 从 Kafka 消费告警事件并落库 |
| A4-F08 | 告警事件落库 | P0 | 将消费的告警事件写入 alert_event 表 |
| A4-F09 | vmalert 默认适配 | P0 | 实现 vmalert 的规则推送/撤回 |
| A4-F10 | alertmanager 默认适配 | P0 | 实现 alertmanager 的静默/查询 |
| A4-F11 | 告警统计 | P1 | 按网区、严重级别、状态等维度统计告警 |
| A4-F12 | 告警事件详情 | P1 | 查看告警事件的完整信息（labels、annotations、时间线） |
| A4-F13 | 引擎健康检查 | P1 | 检查底层告警引擎的运行状态 |
| A4-F14 | 替换引擎适配 | P2 | 实现 Prometheus rules / Thanos Ruler / Cortex 适配 |

---

## 四、核心数据模型

### 4.1 alert_event（告警事件表）

> 引用 data-model.md 4.6 节定义。

```sql
CREATE TABLE alert_event (
    event_id        BIGINT AUTO_INCREMENT PRIMARY KEY,
    alert_name      VARCHAR(256) NOT NULL COMMENT '告警规则名称',
    zone_id         VARCHAR(64) COMMENT '所属网区',
    instance_id     VARCHAR(64) COMMENT '关联实例',
    severity        VARCHAR(32) COMMENT 'critical/warning/info',
    status          VARCHAR(32) COMMENT 'firing/resolved',
    labels          JSON COMMENT '告警标签',
    annotations     JSON COMMENT '告警注释',
    starts_at       TIMESTAMP COMMENT '告警触发时间',
    ends_at         TIMESTAMP NULL COMMENT '告警恢复时间',
    fingerprint     VARCHAR(128) COMMENT '告警指纹，用于去重',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ae_zone ON alert_event(zone_id);
CREATE INDEX idx_ae_status ON alert_event(status);
CREATE INDEX idx_ae_time ON alert_event(starts_at);
CREATE INDEX idx_ae_fingerprint ON alert_event(fingerprint);
```

**关键字段说明：**

| 字段 | 说明 |
|------|------|
| `fingerprint` | 告警指纹，由 alertmanager 基于 labels 计算。相同 fingerprint 的告警视为同一告警的不同状态（firing → resolved） |
| `status` | `firing` = 告警触发中；`resolved` = 告警已恢复 |
| `labels` | 包含告警的所有标签，如 `instance`、`zone`、`severity`、`alertname` 等 |
| `annotations` | 包含告警的描述信息，如 `summary`、`description` 等 |

### 4.2 告警认领记录表（建议新增）

```sql
CREATE TABLE alert_acknowledgement (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id        BIGINT NOT NULL COMMENT '关联告警事件ID',
    user_id         VARCHAR(64) NOT NULL COMMENT '认领人',
    acknowledged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '认领时间',
    note            TEXT COMMENT '认领备注',
    FOREIGN KEY (event_id) REFERENCES alert_event(event_id),
    FOREIGN KEY (user_id) REFERENCES user(user_id)
);

CREATE INDEX idx_ack_event ON alert_acknowledgement(event_id);
CREATE INDEX idx_ack_user ON alert_acknowledgement(user_id);
```

### 4.3 告警静默记录表（建议新增）

```sql
CREATE TABLE alert_silence (
    silence_id      VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) COMMENT '适用网区，NULL表示全局',
    matchers        JSON NOT NULL COMMENT '匹配条件，如 [{"name":"alertname","value":"MySQL.*","isRegex":true}]',
    starts_at       TIMESTAMP NOT NULL COMMENT '静默开始时间',
    ends_at         TIMESTAMP NOT NULL COMMENT '静默结束时间',
    duration        VARCHAR(32) COMMENT '静默时长，如 1h/30m',
    reason          TEXT COMMENT '静默原因',
    created_by      VARCHAR(64) NOT NULL COMMENT '创建人',
    status          VARCHAR(32) DEFAULT 'active' COMMENT 'active/expired/revoked',
    engine_silence_id VARCHAR(128) COMMENT '底层引擎中的静默ID',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES user(user_id)
);

CREATE INDEX idx_silence_zone ON alert_silence(zone_id);
CREATE INDEX idx_silence_status ON alert_silence(status);
CREATE INDEX idx_silence_time ON alert_silence(starts_at, ends_at);
```

---

## 五、接口与交互

### 5.1 上游依赖

| 上游模块 | 依赖接口 | 用途 |
|----------|----------|------|
| A1 告警策略模块 | 规则生成结果 | 获取需要推送的告警规则（RCRule） |
| Kafka | 告警事件消息 | 消费 alertmanager 发出的告警事件 |

### 5.2 下游提供

| 下游模块 | 提供接口 | 用途 |
|----------|----------|------|
| A3 通知渠道模块 | 告警事件数据 | 触发通知投递流程 |
| A2 告警组模块 | 告警认领状态 | 升级超时检测依赖认领状态 |
| 前端展示 | 告警事件查询 | 告警列表、告警详情、告警统计 |

### 5.3 AlertEngine 接口定义

```go
// AlertEngine 可替换的告警引擎抽象接口
type AlertEngine interface {
    // PushRules 向指定网区推送告警规则
    // zoneID: 目标网区
    // rules: 告警规则列表
    PushRules(ctx context.Context, zoneID string, rules []*AlertRule) error

    // RemoveRules 从指定网区撤回告警规则
    // zoneID: 目标网区
    // ruleIDs: 要撤回的规则ID列表
    RemoveRules(ctx context.Context, zoneID string, ruleIDs []string) error

    // QueryAlerts 查询告警事件
    // filter: 查询过滤条件
    QueryAlerts(ctx context.Context, filter *AlertFilter) (*AlertQueryResult, error)

    // Acknowledge 认领告警
    // alertID: 告警事件标识
    // userID: 认领人
    Acknowledge(ctx context.Context, alertID string, userID string) error

    // Silence 设置告警静默
    // filter: 静默匹配条件
    // duration: 静默时长
    // reason: 静默原因
    Silence(ctx context.Context, filter *SilenceFilter, duration time.Duration, reason string) (*SilenceResult, error)

    // HealthCheck 检查引擎健康状态
    HealthCheck(ctx context.Context, zoneID string) (*EngineHealth, error)

    // Type 返回引擎类型标识
    Type() string
}
```

**数据结构定义：**

```go
// AlertRule 告警规则
type AlertRule struct {
    RuleID      string            `json:"rule_id"`
    AlertName   string            `json:"alert_name"`
    Expr        string            `json:"expr"`
    Duration    string            `json:"duration"`
    Severity    string            `json:"severity"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
}

// AlertFilter 告警查询过滤条件
type AlertFilter struct {
    ZoneID     string   `json:"zone_id"`
    AlertName  string   `json:"alert_name"`
    Severity   []string `json:"severity"`
    Status     []string `json:"status"`
    InstanceID string   `json:"instance_id"`
    StartTime  time.Time `json:"start_time"`
    EndTime    time.Time `json:"end_time"`
    Fingerprint string  `json:"fingerprint"`
    Page       int      `json:"page"`
    PageSize   int      `json:"page_size"`
}

// AlertQueryResult 告警查询结果
type AlertQueryResult struct {
    Items    []*AlertEvent `json:"items"`
    Total    int64         `json:"total"`
    Page     int           `json:"page"`
    PageSize int           `json:"page_size"`
}

// AlertEvent 告警事件
type AlertEvent struct {
    EventID      int64             `json:"event_id"`
    AlertName    string            `json:"alert_name"`
    ZoneID       string            `json:"zone_id"`
    InstanceID   string            `json:"instance_id"`
    Severity     string            `json:"severity"`
    Status       string            `json:"status"`
    Labels       map[string]string `json:"labels"`
    Annotations  map[string]string `json:"annotations"`
    StartsAt     time.Time         `json:"starts_at"`
    EndsAt       *time.Time        `json:"ends_at"`
    Fingerprint  string            `json:"fingerprint"`
    Acknowledged bool              `json:"acknowledged"`
    AckUser      string            `json:"ack_user"`
}

// SilenceFilter 静默匹配条件
type SilenceFilter struct {
    ZoneID      string            `json:"zone_id"`
    AlertName   string            `json:"alert_name"`
    Labels      map[string]string `json:"labels"`
    InstanceID  string            `json:"instance_id"`
}

// SilenceResult 静默操作结果
type SilenceResult struct {
    SilenceID       string    `json:"silence_id"`
    EngineSilenceID string    `json:"engine_silence_id"`
    StartsAt        time.Time `json:"starts_at"`
    EndsAt          time.Time `json:"ends_at"`
}

// EngineHealth 引擎健康状态
type EngineHealth struct {
    EngineType string `json:"engine_type"`
    ZoneID     string `json:"zone_id"`
    Status     string `json:"status"`
    RuleCount  int    `json:"rule_count"`
    Uptime     string `json:"uptime"`
    Message    string `json:"message"`
}
```

### 5.4 默认实现：vmalert + alertmanager

#### 5.4.1 vmalert 规则推送

```go
type VmalertEngine struct {
    vmalertURLs map[string]string // zoneID → vmalert URL
}

func (e *VmalertEngine) PushRules(ctx context.Context, zoneID string, rules []*AlertRule) error {
    vmalertURL := e.vmalertURLs[zoneID]

    // 将规则转换为 vmalert rule group 格式
    ruleGroup := &VmalertRuleGroup{
        Name:     fmt.Sprintf("multdc-%s", zoneID),
        Interval: "1m",
        Rules:    make([]VmalertRule, 0, len(rules)),
    }

    for _, rule := range rules {
        ruleGroup.Rules = append(ruleGroup.Rules, VmalertRule{
            Alert:  rule.AlertName,
            Expr:   rule.Expr,
            For:    rule.Duration,
            Labels: mergeLabels(rule.Labels, map[string]string{
                "severity": rule.Severity,
                "zone_id":  zoneID,
            }),
            Annotations: rule.Annotations,
        })
    }

    // 通过 vmalert API 推送规则
    // PUT /api/v1/rules/{group_name}
    body, _ := json.Marshal(ruleGroup)
    req, _ := http.NewRequest("PUT",
        fmt.Sprintf("%s/api/v1/rules/%s", vmalertURL, ruleGroup.Name),
        bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("push rules to vmalert failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("vmalert returned status %d", resp.StatusCode)
    }

    return nil
}

func (e *VmalertEngine) RemoveRules(ctx context.Context, zoneID string, ruleIDs []string) error {
    vmalertURL := e.vmalertURLs[zoneID]

    // 通过 vmalert API 删除指定规则
    // DELETE /api/v1/rules/{group_name}?rules=rule1,rule2
    params := url.Values{}
    params.Set("rules", strings.Join(ruleIDs, ","))

    req, _ := http.NewRequest("DELETE",
        fmt.Sprintf("%s/api/v1/rules/multdc-%s?%s", vmalertURL, zoneID, params.Encode()),
        nil)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("remove rules from vmalert failed: %w", err)
    }
    defer resp.Body.Close()

    return nil
}
```

#### 5.4.2 alertmanager 静默与查询

```go
func (e *VmalertEngine) Acknowledge(ctx context.Context, alertID string, userID string) error {
    // 认领操作：
    // 1. 在 alert_acknowledgement 表中记录认领
    // 2. 可选：在 alertmanager 中添加 annotation 标记
    // vmalert 本身不支持认领，认领状态由管理平台维护
    return nil
}

func (e *VmalertEngine) Silence(ctx context.Context, filter *SilenceFilter, duration time.Duration, reason string) (*SilenceResult, error) {
    // 调用 alertmanager API 创建静默
    // POST /api/v2/silences
    silence := map[string]interface{}{
        "matchers": buildMatchers(filter),
        "startsAt": time.Now().UTC().Format(time.RFC3339),
        "endsAt":   time.Now().Add(duration).UTC().Format(time.RFC3339),
        "createdBy": "multdc-platform",
        "comment":  reason,
    }

    body, _ := json.Marshal(silence)
    resp, err := http.Post(e.alertmanagerURL+"/api/v2/silences",
        "application/json", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create silence failed: %w", err)
    }
    defer resp.Body.Close()

    var result struct { SilenceID string `json:"silenceID"` }
    json.NewDecoder(resp.Body).Decode(&result)

    return &SilenceResult{
        SilenceID:       generateSilenceID(),
        EngineSilenceID: result.SilenceID,
        StartsAt:        time.Now(),
        EndsAt:          time.Now().Add(duration),
    }, nil
}
```

#### 5.4.3 Kafka 消费与告警事件落库

```go
// AlertEventConsumer 告警事件消费者
type AlertEventConsumer struct {
    consumer  kafka.Consumer
    eventRepo AlertEventRepository
    notifier  NotificationDispatcher
}

func (c *AlertEventConsumer) Start(ctx context.Context) error {
    // 订阅 alertmanager 输出的 Kafka topic
    c.consumer.Subscribe("alertmanager-alerts")

    for {
        msg, err := c.consumer.Receive(ctx)
        if err != nil {
            if ctx.Err() != nil {
                return nil
            }
            log.Errorf("kafka receive error: %v", err)
            continue
        }

        // 解析 alertmanager webhook 格式
        var alert AlertmanagerAlert
        if err := json.Unmarshal(msg.Value, &alert); err != nil {
            log.Errorf("unmarshal alert failed: %v", err)
            continue
        }

        // 落库
        event := &AlertEvent{
            AlertName:   alert.Labels["alertname"],
            ZoneID:      alert.Labels["zone_id"],
            InstanceID:  alert.Labels["instance"],
            Severity:    alert.Labels["severity"],
            Status:      alert.Status,
            Labels:      alert.Labels,
            Annotations: alert.Annotations,
            StartsAt:    alert.StartsAt,
            EndsAt:      alert.EndsAt,
            Fingerprint: alert.Fingerprint,
        }

        savedEvent, err := c.eventRepo.Save(ctx, event)
        if err != nil {
            log.Errorf("save alert event failed: %v", err)
            continue
        }

        // 如果是 firing 状态且未被静默，触发通知
        if alert.Status == "firing" && !c.isSilenced(alert) {
            c.dispatchNotification(ctx, savedEvent)
        }

        c.consumer.Ack(msg)
    }
}
```

### 5.5 对外 API

#### 5.5.1 查询告警事件列表

```
GET /api/v1/alert-events
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| zone_id | string | 否 | 按网区过滤 |
| alert_name | string | 否 | 按告警名称过滤 |
| severity | string | 否 | 按严重级别过滤，可多选：critical,warning,info |
| status | string | 否 | 按状态过滤：firing/resolved |
| instance_id | string | 否 | 按实例 ID 过滤 |
| fingerprint | string | 否 | 按告警指纹过滤 |
| start_time | timestamp | 否 | 起始时间 |
| end_time | timestamp | 否 | 结束时间 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "event_id": 12345,
        "alert_name": "MySQL连接数过高",
        "zone_id": "zone-east-1",
        "instance_id": "inst-mysql-001",
        "severity": "warning",
        "status": "firing",
        "labels": {
          "alertname": "MySQL连接数过高",
          "instance": "inst-mysql-001",
          "zone_id": "zone-east-1",
          "severity": "warning"
        },
        "annotations": {
          "summary": "MySQL inst-mysql-001 连接数过高",
          "description": "当前连接数 480，超过阈值的 90%"
        },
        "starts_at": "2026-10-27T10:25:00Z",
        "ends_at": null,
        "fingerprint": "abc123def456",
        "acknowledged": false,
        "created_at": "2026-10-27T10:25:01Z"
      }
    ],
    "total": 42,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.5.2 查询告警事件详情

```
GET /api/v1/alert-events/{event_id}
```

**响应：** 返回告警事件的完整信息，包括认领记录、静默状态、关联策略等。

#### 5.5.3 认领告警

```
POST /api/v1/alert-events/{event_id}/acknowledge
```

**请求体：**

```json
{
  "user_id": "user-001",
  "note": "已收到，正在排查"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "event_id": 12345,
    "acknowledged": true,
    "ack_user": "user-001",
    "acknowledged_at": "2026-10-27T10:30:00Z"
  }
}
```

**行为：**

1. 在 alert_acknowledgement 表中记录认领信息。
2. 停止该告警的升级计时（通知 A2 模块）。
3. 若底层引擎支持认领（如 alertmanager），同步更新引擎侧状态。

#### 5.5.4 设置告警静默

```
POST /api/v1/alert-silences
```

**请求体：**

```json
{
  "zone_id": "zone-east-1",
  "matchers": [
    { "name": "alertname", "value": "MySQL连接数过高", "is_regex": false },
    { "name": "instance", "value": "inst-mysql-001", "is_regex": false }
  ],
  "duration": "2h",
  "reason": "计划维护窗口，已通知 DBA 团队"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "silence_id": "silence-001",
    "engine_silence_id": "am-silence-abc123",
    "starts_at": "2026-10-27T10:35:00Z",
    "ends_at": "2026-10-27T12:35:00Z",
    "status": "active"
  }
}
```

#### 5.5.5 查询静默列表

```
GET /api/v1/alert-silences
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| zone_id | string | 否 | 按网区过滤 |
| status | string | 否 | 按状态过滤：active/expired/revoked |

#### 5.5.6 撤销静默

```
POST /api/v1/alert-silences/{silence_id}/revoke
```

#### 5.5.7 告警统计

```
GET /api/v1/alert-events/statistics
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| zone_id | string | 否 | 按网区过滤 |
| period | string | 否 | 统计周期：1h/6h/24h/7d |

**响应：**

```json
{
  "code": 0,
  "data": {
    "period": "24h",
    "total_firing": 12,
    "total_resolved": 45,
    "by_severity": {
      "critical": 3,
      "warning": 7,
      "info": 2
    },
    "by_zone": [
      { "zone_id": "zone-east-1", "firing": 5, "resolved": 20 },
      { "zone_id": "zone-west-1", "firing": 7, "resolved": 25 }
    ],
    "top_alert_names": [
      { "alert_name": "MySQL连接数过高", "count": 15 },
      { "alert_name": "磁盘空间不足", "count": 10 },
      { "alert_name": "CPU使用率过高", "count": 8 }
    ]
  }
}
```

#### 5.5.8 引擎健康检查

```
GET /api/v1/alert-engines/health
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "engines": [
      {
        "engine_type": "vmalert",
        "zone_id": "zone-east-1",
        "status": "healthy",
        "rule_count": 35,
        "uptime": "72h",
        "message": ""
      },
      {
        "engine_type": "vmalert",
        "zone_id": "zone-west-1",
        "status": "healthy",
        "rule_count": 28,
        "uptime": "72h",
        "message": ""
      }
    ]
  }
}
```

---

## 六、设计决策

### 6.1 AlertEngine 可替换设计

**决策：** 通过 `AlertEngine` 接口抽象告警引擎能力，默认实现可被替换。

**接口方法与设计意图：**

| 方法 | 设计意图 |
|------|----------|
| `PushRules(zoneID, rules)` | 将告警规则推送至指定网区的引擎。不同引擎的规则格式不同，由实现层负责转换 |
| `RemoveRules(zoneID, ruleIDs)` | 从引擎撤回指定规则。用于策略禁用/删除时清理规则 |
| `QueryAlerts(filter)` | 查询告警事件。默认实现从 alert_event 表查询，也可代理引擎原生查询 API |
| `Acknowledge(alertID, userID)` | 认领告警。部分引擎原生支持认领（如 PagerDuty），不支持的实现由管理平台维护状态 |
| `Silence(filter, duration, reason)` | 设置静默。默认实现调用 alertmanager API，也可由管理平台在消费侧过滤 |
| `HealthCheck(zoneID)` | 检查引擎健康状态 |

### 6.2 替换路径

#### 6.2.1 告警规则评估引擎替换

| 默认实现 | 替换方案 | 适配要点 |
|----------|----------|----------|
| vmalert | Prometheus rules | 通过 Prometheus API 管理 rule files，或使用 file_sd 写入规则文件 |
| vmalert | Thanos Ruler | 通过 Thanos Ruler API 推送规则，支持跨网区全局规则 |
| vmalert | Grafana Mimir | 通过 Mimir Ruler API 管理规则，兼容 Prometheus 规则格式 |

**Prometheus rules 替换实现示例：**

```go
type PrometheusRulesEngine struct {
    prometheusURLs map[string]string
    ruleDir        string // 规则文件目录
}

func (e *PrometheusRulesEngine) PushRules(ctx context.Context, zoneID string, rules []*AlertRule) error {
    // 生成 Prometheus 格式的 rule file
    ruleGroup := generatePrometheusRuleGroup(zoneID, rules)
    yamlData, _ := yaml.Marshal(ruleGroup)

    // 写入规则文件
    filePath := filepath.Join(e.ruleDir, fmt.Sprintf("multdc-%s.yaml", zoneID))
    os.WriteFile(filePath, yamlData, 0644)

    // 触发 Prometheus reload
    return http.Post(e.prometheusURLs[zoneID]+"/-/reload", "", nil)
}
```

#### 6.2.2 告警收敛组件替换

| 默认实现 | 替换方案 | 适配要点 |
|----------|----------|----------|
| alertmanager | Cortex alertmanager | 通过 Cortex API 管理静默/抑制，多租户支持 |
| alertmanager | 自定义收敛逻辑 | 在 Kafka 消费侧实现分组/抑制/静默，不依赖外部组件 |
| alertmanager | Grafana Alerting | 使用 Grafana 内置的告警收敛能力 |

#### 6.2.3 消息队列替换

| 默认实现 | 替换方案 | 适配要点 |
|----------|----------|----------|
| Kafka | RabbitMQ | 实现 RabbitMQ 消费者，订阅 alertmanager 的 exchange |
| Kafka | Apache Pulsar | 实现 Pulsar 消费者，支持 Pulsar 的分区与订阅模式 |
| Kafka | Redis Streams | 轻量级方案，适合小规模部署 |

**消息队列抽象：**

```go
// MessageQueue 消息队列抽象
type MessageQueue interface {
    Subscribe(topic string) (<-Message, error)
    Publish(topic string, msg []byte) error
    Ack(msg Message) error
    Close() error
}
```

### 6.3 Kafka 消费与落库设计

**决策：** 告警事件从 Kafka 消费后先落库，再触发通知投递。

**流程：**

```
alertmanager webhook → Kafka producer → Kafka topic
                                              │
                                              ▼
                                   AlertEventConsumer
                                              │
                                   ┌──────────┴──────────┐
                                   │                     │
                                   ▼                     ▼
                              落库 alert_event     检查静默规则
                                   │                     │
                                   │              ┌──────┴──────┐
                                   │              │             │
                                   │           未静默         已静默
                                   │              │             │
                                   │              ▼             ▼
                                   │     解析路由(A2)      记录日志
                                   │              │
                                   │              ▼
                                   │     通知投递(A3)
                                   │              │
                                   └──────────────┘
```

**理由：**

- 先落库确保告警事件不丢失，即使通知投递失败也可追溯。
- 落库后再检查静默规则，避免静默的告警触发通知但仍保留记录。
- 消费与投递解耦，通知投递失败不影响消费进度。

### 6.4 告警指纹与去重

**决策：** 使用 alertmanager 生成的 `fingerprint` 作为告警去重依据。

**指纹计算：**

```
fingerprint = SHA256(sorted_labels)
```

其中 `sorted_labels` 为按 key 排序后的 labels 键值对拼接。

**去重逻辑：**

- 收到 `firing` 状态的告警时，查找同 fingerprint 的最近一条记录。
- 若最近记录为 `firing` 且未 `resolved`，更新该记录（而非新建）。
- 收到 `resolved` 状态的告警时，更新同 fingerprint 的 firing 记录，设置 `ends_at` 和 `status=resolved`。

### 6.5 告警认领机制

**决策：** 认领状态由管理平台维护，而非依赖底层引擎。

**理由：**

- 并非所有告警引擎原生支持认领（vmalert 不支持，alertmanager 的 silence 语义不同）。
- 管理平台维护认领状态，确保无论底层引擎如何替换，认领行为一致。
- 认领状态直接影响升级链的超时检测（A2 模块），由平台统一管理更可靠。

**认领与升级链的交互：**

```
告警触发 → 分发到 L1 组
              │
              ├── 有人认领 → 停止升级计时
              │
              └── 无人认领 → 达到 escalation_minutes
                                │
                                ▼
                            升级到 L2 组
                                │
                                ├── 有人认领 → 停止升级
                                │
                                └── 无人认领 → 继续升级
```

### 6.6 告警静默的双层实现

**决策：** 静默在两个层面实现 —— 引擎层（alertmanager）和平台层（Kafka 消费侧）。

| 层面 | 实现方式 | 适用场景 |
|------|----------|----------|
| 引擎层 | 调用 alertmanager API 创建 silence | 阻止 alertmanager 发送 webhook 到 Kafka |
| 平台层 | 消费 Kafka 消息时检查 alert_silence 表 | 作为兜底，即使引擎层静默失败也能阻止通知 |

**理由：**

- 双层静默互为备份，提高可靠性。
- 平台层静默支持更灵活的管理（如按网区、按策略批量静默）。
- 引擎层静默减少不必要的 Kafka 消息量，降低系统负载。
