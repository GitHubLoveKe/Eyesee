# I3 状态缓存

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：I3 | 职责：定时聚合平台状态到 Redis 缓存，为前端仪表盘提供高性能实时数据查询

---

## 一、概述

状态缓存模块是平台数据读取的性能层，负责将分散在数据库各表中的状态信息（网区概览、节点状态、实例状态、告警统计等）定时聚合到 Redis 中，为前端仪表盘和 API 查询提供毫秒级响应。模块由 T5 定时任务调度执行聚合，采用 TTL + 定时刷新的缓存更新策略，并具备 Redis 不可用时的数据库降级查询能力。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| Redis 数据结构设计与维护 | 定义缓存 key 规范与数据结构 |
| 状态聚合执行 | 从数据库聚合各维度状态数据写入 Redis |
| 缓存更新调度 | 响应 T5 定时任务的聚合触发 |
| 仪表盘数据提供 | 为前端轮询提供高性能缓存读取接口 |
| 降级查询 | Redis 不可用时回退到数据库直查 |
| 缓存生命周期管理 | TTL 设置、过期清理、预热 |

### 本模块不负责

| 不负责项 | 归属模块 |
|----------|----------|
| 定时调度触发 | T5 定时任务 |
| 节点/实例状态的原始采集 | I1 采集器管理 / 各采集器节点 |
| 告警事件的产生 | A1 告警策略 / A4 告警收敛 |
| 数据库持久化 | 各模块自身的 Repository 层 |

---

## 三、功能清单

| # | 功能 | 优先级 | 阶段 | 说明 |
|---|------|--------|------|------|
| F1 | 全局概览聚合 | P0 | Phase 1 | 聚合 total_zones/instances/tasks/alerts 到 global:summary |
| F2 | 网区概览聚合 | P0 | Phase 1 | 按网区聚合实例数、任务数、告警数到 zone:{id}:overview |
| F3 | 节点状态缓存 | P0 | Phase 1 | 缓存各节点状态、负载、心跳到 node:{id}:status |
| F4 | 实例状态缓存 | P1 | Phase 2 | 缓存各实例状态、采集情况到 instance:{id}:status |
| F5 | 仪表盘数据读取 | P0 | Phase 1 | 提供缓存数据的读取 API |
| F6 | 缓存 TTL 管理 | P0 | Phase 1 | 自动过期 + 定时刷新双保险 |
| F7 | 降级查询 | P1 | Phase 2 | Redis 不可用时回退数据库直查 |
| F8 | 缓存预热 | P2 | Phase 3 | 服务启动时预加载关键数据到 Redis |
| F9 | 聚合执行日志 | P1 | Phase 2 | 记录每次聚合的执行时间、影响行数 |

---

## 四、核心数据模型

### 4.1 Redis 数据结构设计

#### 4.1.1 全局概览：`global:summary`

存储平台全局统计摘要，使用 Redis Hash 类型。

```
HSET global:summary
  total_zones       "5"          -- 网区总数
  total_instances   "1200"       -- 实例总数
  total_tasks       "3600"       -- 采集任务总数
  tasks_running     "3580"       -- 运行中任务
  tasks_error       "20"         -- 异常任务
  active_alerts     "12"         -- 当前活跃告警数
  critical_alerts   "3"          -- 严重告警数
  healthy_nodes     "15"         -- 健康节点数
  degraded_nodes    "1"          -- 降级节点数
  offline_nodes     "1"          -- 离线节点数
  last_aggregated   "1727000100" -- 最后聚合时间戳(Unix)
```

**TTL：** 120s（2分钟）
**刷新频率：** 每 30s 由 T5 定时任务触发聚合

#### 4.1.2 网区概览：`zone:{zone_id}:overview`

按网区维度聚合的状态摘要。

```
HSET zone:{zone_id}:overview
  zone_name         "华东区"       -- 网区名称
  instance_count    "240"          -- 该网区实例数
  instance_active   "235"          -- 活跃实例数
  instance_suspended "5"           -- 暂停实例数
  task_running      "720"          -- 运行中任务
  task_error        "3"            -- 异常任务
  task_pending      "0"            -- 待启动任务
  active_alerts     "2"            -- 活跃告警数
  critical_alerts   "0"            -- 严重告警数
  healthy_nodes     "3"            -- 健康节点数
  degraded_nodes    "0"            -- 降级节点数
  offline_nodes     "0"            -- 离线节点数
  collector_type    "vmagent"      -- 主采集器类型
  last_updated      "1727000100"   -- 最后更新时间戳
```

**TTL：** 120s
**刷新频率：** 每 30s

#### 4.1.3 节点状态：`node:{node_id}:status`

单个采集器节点的实时状态。

```
HSET node:{node_id}:status
  node_id           "node-vmagent-east-01"
  zone_id           "zone-east-1"
  name              "华东区 vmagent-01"
  node_type         "vmagent"
  status            "healthy"        -- healthy/degraded/offline
  address           "http://10.0.1.100:8429"
  task_count        "120"            -- 当前分配任务数
  max_tasks         "500"            -- 最大容量
  load_ratio        "0.24"           -- 负载比
  uptime_seconds    "864000"         -- 运行时长
  last_heartbeat    "1727000090"     -- 最后心跳时间戳
  last_scrape       "1727000085"     -- 最后采集时间戳
  error_count       "0"              -- 近期错误数
  last_updated      "1727000100"
```

**TTL：** 90s
**刷新频率：** 每 30s（与心跳周期对齐）

#### 4.1.4 实例状态：`instance:{instance_id}:status`

单个被监控实例的状态摘要。

```
HSET instance:{instance_id}:status
  instance_id       "inst-001"
  zone_id           "zone-east-1"
  name              "MySQL 生产实例"
  resource_type     "mysql"
  status            "active"         -- pending/active/suspended/decommissioned
  host              "10.0.1.5"
  port              "3306"
  task_count        "3"              -- 关联采集任务数
  task_running      "3"              -- 运行中任务数
  task_error        "0"              -- 异常任务数
  last_scrape       "1727000085"     -- 最后一次成功采集时间
  error_count_1h    "0"              -- 近1小时错误次数
  active_alerts     "1"              -- 该实例活跃告警数
  node_id           "node-vmagent-east-01"  -- 分配的采集节点
  last_updated      "1727000100"
```

**TTL：** 90s
**刷新频率：** 每 30s

#### 4.1.5 告警概览：`alert:summary`

当前活跃告警的快速统计。

```
HSET alert:summary
  total_active      "12"
  by_severity_critical  "3"
  by_severity_warning   "7"
  by_severity_info      "2"
  by_zone:zone-east-1   "5"
  by_zone:zone-bj-1     "4"
  by_zone:zone-dmz-1    "3"
  last_updated      "1727000100"
```

**TTL：** 60s
**刷新频率：** 每 15s（告警需要更高实时性）

### 4.2 聚合 SQL 查询

#### 全局概览聚合查询

```sql
-- global:summary 聚合
SELECT
  (SELECT COUNT(*) FROM zone WHERE status = 'active') AS total_zones,
  (SELECT COUNT(*) FROM instance WHERE status != 'decommissioned') AS total_instances,
  (SELECT COUNT(*) FROM collection_task) AS total_tasks,
  (SELECT COUNT(*) FROM collection_task WHERE status = 'running') AS tasks_running,
  (SELECT COUNT(*) FROM collection_task WHERE status = 'error') AS tasks_error,
  (SELECT COUNT(*) FROM alert_event WHERE status = 'firing') AS active_alerts,
  (SELECT COUNT(*) FROM alert_event WHERE status = 'firing' AND severity = 'critical') AS critical_alerts,
  (SELECT COUNT(*) FROM collector_node WHERE status = 'healthy') AS healthy_nodes,
  (SELECT COUNT(*) FROM collector_node WHERE status = 'degraded') AS degraded_nodes,
  (SELECT COUNT(*) FROM collector_node WHERE status = 'offline') AS offline_nodes;
```

#### 网区概览聚合查询

```sql
-- zone:{zone_id}:overview 聚合
SELECT
  z.name AS zone_name,
  z.collector_type,
  COUNT(DISTINCT i.instance_id) AS instance_count,
  COUNT(DISTINCT CASE WHEN i.status = 'active' THEN i.instance_id END) AS instance_active,
  COUNT(DISTINCT CASE WHEN i.status = 'suspended' THEN i.instance_id END) AS instance_suspended,
  COUNT(DISTINCT ct.task_id) AS task_running,
  COUNT(DISTINCT CASE WHEN ct.status = 'error' THEN ct.task_id END) AS task_error,
  COUNT(DISTINCT CASE WHEN ct.status = 'pending' THEN ct.task_id END) AS task_pending,
  COUNT(DISTINCT CASE WHEN ae.status = 'firing' THEN ae.event_id END) AS active_alerts,
  COUNT(DISTINCT CASE WHEN cn.status = 'healthy' THEN cn.node_id END) AS healthy_nodes,
  COUNT(DISTINCT CASE WHEN cn.status = 'degraded' THEN cn.node_id END) AS degraded_nodes,
  COUNT(DISTINCT CASE WHEN cn.status = 'offline' THEN cn.node_id END) AS offline_nodes
FROM zone z
LEFT JOIN instance i ON i.zone_id = z.zone_id AND i.status != 'decommissioned'
LEFT JOIN collection_task ct ON ct.instance_id = i.instance_id AND ct.enabled = TRUE
LEFT JOIN collector_node cn ON cn.zone_id = z.zone_id
LEFT JOIN alert_event ae ON ae.zone_id = z.zone_id AND ae.status = 'firing'
WHERE z.zone_id = ?
GROUP BY z.zone_id, z.name, z.collector_type;
```

### 4.3 数据库扩展表（聚合执行记录）

```sql
CREATE TABLE status_aggregation_log (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    aggregation_type VARCHAR(32) NOT NULL COMMENT 'global/zone/node/instance/alert',
    target_id       VARCHAR(64) COMMENT '聚合目标ID( zone_id / node_id / instance_id )',
    status          VARCHAR(16) NOT NULL COMMENT 'success/failed',
    duration_ms     INT COMMENT '聚合耗时(毫秒)',
    rows_affected   INT COMMENT '影响Redis key数量',
    error_message   TEXT,
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sagg_type ON status_aggregation_log(aggregation_type);
CREATE INDEX idx_sagg_time ON status_aggregation_log(started_at);
CREATE INDEX idx_sagg_status ON status_aggregation_log(status);
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| T5 定时任务 | 聚合调度触发（每 30s / 15s） | 定时回调 |
| M1 网区管理 | zone 表数据 | 数据库查询 |
| M3 实例管理 | instance 表数据 | 数据库查询 |
| T1 采集任务管理 | collection_task 表数据 | 数据库查询 |
| I1 采集器管理 | collector_node 表数据 | 数据库查询 |
| A1 告警策略 | alert_event 表数据 | 数据库查询 |

### 5.2 下游提供

| 消费方 | 提供内容 | 交互方式 |
|--------|----------|----------|
| 前端仪表盘 | 全局概览、网区概览、节点/实例状态 | REST API（读 Redis） |
| 告警大屏 | 活跃告警统计 | REST API（读 Redis） |
| 数据导出 | 聚合数据批量导出 | REST API |

### 5.3 对外 API

#### 5.3.1 获取全局概览

```
GET /api/v1/dashboard/summary
```

**响应（从 Redis 读取）：**

```json
{
  "code": 0,
  "data": {
    "total_zones": 5,
    "total_instances": 1200,
    "total_tasks": 3600,
    "tasks_running": 3580,
    "tasks_error": 20,
    "active_alerts": 12,
    "critical_alerts": 3,
    "healthy_nodes": 15,
    "degraded_nodes": 1,
    "offline_nodes": 1,
    "last_aggregated": "2026-10-27T10:35:00Z",
    "data_source": "cache"
  }
}
```

> `data_source` 字段标识数据来源：`cache` 表示来自 Redis，`database` 表示降级直查数据库。

#### 5.3.2 获取网区概览列表

```
GET /api/v1/dashboard/zones
```

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_ids | string | 逗号分隔的网区ID，不传则返回全部 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "zone_id": "zone-east-1",
        "zone_name": "华东区",
        "instance_count": 240,
        "instance_active": 235,
        "task_running": 720,
        "task_error": 3,
        "active_alerts": 2,
        "healthy_nodes": 3,
        "degraded_nodes": 0,
        "offline_nodes": 0,
        "collector_type": "vmagent",
        "last_updated": "2026-10-27T10:35:00Z"
      },
      {
        "zone_id": "zone-bj-1",
        "zone_name": "北京区",
        "instance_count": 180,
        "instance_active": 178,
        "task_running": 540,
        "task_error": 1,
        "active_alerts": 4,
        "healthy_nodes": 2,
        "degraded_nodes": 1,
        "offline_nodes": 0,
        "collector_type": "prometheus",
        "last_updated": "2026-10-27T10:35:00Z"
      }
    ],
    "data_source": "cache"
  }
}
```

#### 5.3.3 获取单个网区概览

```
GET /api/v1/dashboard/zones/{zone_id}
```

#### 5.3.4 获取节点状态

```
GET /api/v1/dashboard/nodes/{node_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "node_id": "node-vmagent-east-01",
    "zone_id": "zone-east-1",
    "name": "华东区 vmagent-01",
    "node_type": "vmagent",
    "status": "healthy",
    "task_count": 120,
    "max_tasks": 500,
    "load_ratio": 0.24,
    "uptime_seconds": 864000,
    "last_heartbeat": "2026-10-27T10:34:50Z",
    "error_count": 0,
    "last_updated": "2026-10-27T10:35:00Z",
    "data_source": "cache"
  }
}
```

#### 5.3.5 获取节点列表

```
GET /api/v1/dashboard/nodes
```

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| status | string | 按状态过滤 |
| node_type | string | 按采集器类型过滤 |

#### 5.3.6 获取实例状态

```
GET /api/v1/dashboard/instances/{instance_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "instance_id": "inst-001",
    "zone_id": "zone-east-1",
    "name": "MySQL 生产实例",
    "resource_type": "mysql",
    "status": "active",
    "host": "10.0.1.5",
    "port": 3306,
    "task_count": 3,
    "task_running": 3,
    "task_error": 0,
    "last_scrape": "2026-10-27T10:34:45Z",
    "active_alerts": 1,
    "node_id": "node-vmagent-east-01",
    "last_updated": "2026-10-27T10:35:00Z",
    "data_source": "cache"
  }
}
```

#### 5.3.7 获取告警概览

```
GET /api/v1/dashboard/alerts
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "total_active": 12,
    "by_severity": {
      "critical": 3,
      "warning": 7,
      "info": 2
    },
    "by_zone": {
      "zone-east-1": 5,
      "zone-bj-1": 4,
      "zone-dmz-1": 3
    },
    "last_updated": "2026-10-27T10:35:00Z",
    "data_source": "cache"
  }
}
```

#### 5.3.8 手动触发聚合

```
POST /api/v1/dashboard/aggregate
```

**权限：** admin

**请求体（可选）：**

```json
{
  "scope": "zone",
  "target_id": "zone-east-1"
}
```

- `scope` 可选值：`global` / `zone` / `node` / `instance` / `all`
- 不传则聚合全部

**响应：**

```json
{
  "code": 0,
  "data": {
    "aggregation_id": "agg-20261027-103500",
    "scope": "all",
    "duration_ms": 245,
    "keys_updated": 28,
    "status": "success"
  }
}
```

#### 5.3.9 获取聚合执行日志

```
GET /api/v1/dashboard/aggregation-logs
```

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| aggregation_type | string | 聚合类型 |
| status | string | 执行状态 |
| start_time | string | 起始时间 |
| end_time | string | 结束时间 |
| page | int | 页码 |
| page_size | int | 每页数量 |

---

## 六、设计决策

### 6.1 Redis Key 命名规范

**决策：** 采用冒号分隔的层级命名，与 Redis 社区最佳实践一致。

| Key 模式 | 类型 | 说明 | TTL |
|----------|------|------|-----|
| `global:summary` | Hash | 全局概览 | 120s |
| `zone:{zone_id}:overview` | Hash | 网区概览 | 120s |
| `node:{node_id}:status` | Hash | 节点状态 | 90s |
| `instance:{instance_id}:status` | Hash | 实例状态 | 90s |
| `alert:summary` | Hash | 告警统计 | 60s |
| `agg:lock:{type}` | String | 聚合执行分布式锁 | 30s |

**命名理由：**
- 冒号分隔便于 Redis 可视化工具（如 RedisInsight）按层级浏览
- 前缀区分不同维度的聚合数据，避免 key 冲突
- TTL 按数据实时性需求分级：告警(60s) < 节点/实例(90s) < 概览(120s)

### 6.2 聚合机制设计

**决策：** 由 T5 定时任务模块调度触发，本模块负责执行聚合逻辑。

**聚合流程：**

```
T5 定时器触发 (每30s)
    │
    ▼
获取分布式锁 agg:lock:global (防止多实例重复聚合)
    │
    ├── 获取锁失败 → 跳过本次聚合
    │
    ▼
获取锁成功 (TTL=30s)
    │
    ▼
执行全局概览聚合 (global:summary)
    │
    ▼
遍历所有 active 网区，执行网区概览聚合 (zone:{id}:overview)
    │
    ▼
遍历所有非 offline 节点，执行节点状态缓存 (node:{id}:status)
    │
    ▼
遍历所有 active 实例，执行实例状态缓存 (instance:{id}:status)
    │
    ▼
执行告警概览聚合 (alert:summary)
    │
    ▼
释放分布式锁 + 记录聚合日志 (status_aggregation_log)
```

**聚合优先级：**

| 优先级 | 聚合类型 | 频率 | 说明 |
|--------|----------|------|------|
| P0 | alert:summary | 15s | 告警数据实时性要求最高 |
| P0 | global:summary | 30s | 仪表盘首屏数据 |
| P1 | zone:{id}:overview | 30s | 网区详情页 |
| P2 | node:{id}:status | 30s | 节点监控 |
| P2 | instance:{id}:status | 30s | 实例监控 |

### 6.3 缓存更新策略

**双保险机制：TTL 过期 + 定时刷新**

```
                    ┌──────────────────────────┐
                    │      Redis 缓存层         │
                    │                          │
                    │  key TTL: 60s ~ 120s      │
                    │  (兜底过期，防止脏数据残留) │
                    └────────────┬─────────────┘
                                 │
                    ┌────────────┴─────────────┐
                    │     T5 定时刷新           │
                    │                          │
                    │  告警: 每15s              │
                    │  概览: 每30s              │
                    │  详情: 每30s              │
                    └──────────────────────────┘
```

**策略说明：**

| 策略 | 说明 |
|------|------|
| TTL 过期 | 即使定时任务异常停止，缓存数据也会在 TTL 后自动失效，避免前端读到过期数据 |
| 定时刷新 | 正常情况下 TTL 远未到期就已刷新，前端始终读到新鲜数据 |
| 主动失效 | 当底层数据发生重大变更（如网区删除），主动 DEL 对应缓存 key |
| 写入保护 | 聚合时先写入新值再刷新 TTL，避免写入过程中 TTL 到期导致 key 丢失 |

### 6.4 降级策略

**决策：** Redis 不可用时自动降级为数据库直查，保证仪表盘基本可用。

**降级触发条件：**

| 条件 | 说明 |
|------|------|
| Redis 连接失败 | 连续 3 次连接超时/拒绝 |
| Redis 命令执行失败 | 关键命令（HGETALL）连续失败 |
| 手动降级 | 管理员通过配置关闭 Redis |

**降级流程：**

```
API 请求到达
    │
    ▼
尝试从 Redis 读取
    │
    ├── 读取成功 → 返回缓存数据 (data_source: "cache")
    │
    ▼
    读取失败 (Redis 不可用)
    │
    ▼
标记降级状态 (持续60s内不再尝试Redis)
    │
    ▼
直接查询数据库执行聚合SQL
    │
    ├── 查询成功 → 返回数据库数据 (data_source: "database")
    │
    └── 查询失败 → 返回 503 错误
```

**降级期间的行为变化：**

| 方面 | 正常模式 | 降级模式 |
|------|----------|----------|
| 数据来源 | Redis 缓存 | 数据库直查 |
| 响应延迟 | < 10ms | 100ms ~ 1s |
| 数据新鲜度 | 最多延迟 30s | 实时 |
| 数据库压力 | 无额外压力 | 每次请求都查库 |
| data_source 标识 | "cache" | "database" |

**降级恢复：**

```
降级状态持续 60s 后
    │
    ▼
尝试 Redis 心跳检测 (PING)
    │
    ├── 成功 → 退出降级模式，恢复缓存读取
    │
    └── 失败 → 继续降级，再等 60s 重试
```

### 6.5 缓存预热

**决策：** 服务启动时主动预加载关键缓存，避免冷启动时前端空数据。

**预热顺序：**

```
服务启动
    │
    ▼
1. 检查 Redis 连接可用性
    │
    ├── 不可用 → 跳过预热，进入降级模式
    │
    ▼
2. 预热 global:summary (全局概览)
    │
    ▼
3. 预热所有 active zone 的 zone:{id}:overview
    │
    ▼
4. 预热所有非 offline 节点的 node:{id}:status
    │
    ▼
5. 预热 alert:summary
    │
    ▼
6. 预热完成，开始接受请求 + 启动定时聚合
```

**预热耗时预估：**

| 数据规模 | 预热耗时 | 说明 |
|----------|----------|------|
| 5 网区 / 15 节点 / 1200 实例 | < 2s | 小规模，单 SQL 批量查询 |
| 50 网区 / 150 节点 / 12000 实例 | < 10s | 中规模，分批查询 |
| 200+ 网区 / 500+ 节点 / 50000+ 实例 | < 30s | 大规模，并行查询 + 流式写入 |

### 6.6 前端轮询策略

**推荐的仪表盘轮询配置：**

| 页面 | 轮询间隔 | 请求 API | 说明 |
|------|----------|----------|------|
| 全局概览页 | 30s | GET /dashboard/summary | 首屏加载全局数据 |
| 网区列表页 | 30s | GET /dashboard/zones | 网区卡片列表 |
| 网区详情页 | 30s | GET /dashboard/zones/{id} | 单网区详情 |
| 节点监控页 | 15s | GET /dashboard/nodes | 节点状态需更高实时性 |
| 实例详情页 | 30s | GET /dashboard/instances/{id} | 实例状态 |
| 告警大屏 | 15s | GET /dashboard/alerts | 告警需最高实时性 |

**前端轮询优化：**

```
页面可见时 (document.visibilityState === 'visible')
  → 正常轮询

页面不可见时 (切换标签页/最小化)
  → 暂停轮询，减少无效请求

页面重新可见时
  → 立即请求一次最新数据
  → 恢复定时轮询
```

### 6.7 聚合性能优化

**批量查询优化：**

| 优化手段 | 说明 |
|----------|------|
| 合并 SQL | 网区概览使用单条 JOIN 查询，避免 N+1 |
| 增量更新 | 仅重新聚合距上次变更的数据（通过 updated_at 判断） |
| Pipeline 写入 | Redis 写入使用 Pipeline 批量提交，减少网络往返 |
| 并行聚合 | 不同维度的聚合可并行执行（global / zone / node / instance） |
| 结果缓存 | 聚合 SQL 查询结果在内存中短暂缓存（5s），避免同一周期内重复查询 |

**监控指标：**

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| aggregation_duration_ms | 单次聚合耗时 | > 5000ms |
| aggregation_error_count | 聚合失败次数 | > 0 |
| redis_latency_ms | Redis 操作延迟 | > 50ms |
| cache_hit_ratio | 缓存命中率 | < 0.95 |
| stale_data_count | 数据过期未刷新的 key 数 | > 0 |
