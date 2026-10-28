# Zone Coordinator 详细设计

> 版本：v1.0 | 日期：2026-10-28
> 定位：每个网区部署一个区协调器进程（可多副本 HA），负责拉取本区资源、健康检测、采集任务与告警策略的协调分发。

---

## 一、概述

Zone Coordinator 是协调面的核心进程，运行在每个网区内。它从管理平台拉取本区的资源清单（vmagent、vmalert、exporter、存储后端等），持续检测各组件健康状态，然后通过一致性哈希将采集任务（scrape job）和告警规则（alert rule）分发到对应的 worker 节点。

多个 Coordinator 副本在同一网区内通过 gossip 协议保持数据和任务派发的一致性，worker 从任意副本拉取到的任务分配结果相同。

```
管理平台 (控制面)
  │
  │  推送: 采集任务 + 告警规则 (POST /api/v1/push/tasks, /api/v1/push/rules)
  │
  ▼
Zone Coordinator × N (每区多副本, gossip 同步)
  │
  ├── 资源拉取: 按 zone_id 拉取本区资源列表
  │     ├── vmagent × N (采集 worker)
  │     ├── vmalert × N (规则评估 worker)
  │     ├── VictoriaMetrics / Prometheus (存储/查询)
  │     ├── OpenSearch / Elasticsearch (日志存储)
  │     ├── OpenObserve (可观测平台)
  │     ├── OTel Collector (数据管道)
  │     ├── S3 (对象存储)
  │     └── Exporters (node/mysql/redis/snmp...)
  │
  ├── 健康检测
  │     ├── 主动探测: GET /-/healthy (各组件 health_endpoint)
  │     └── 被动心跳: worker 定期 POST heartbeat
  │
  ├── 任务协调分发
  │     ├── 一致性哈希: hash(job_id) → ring → owner worker
  │     ├── 推送: SSE 通知 worker 任务有变更
  │     └── 拉取: worker 定期 GET 获取完整分配
  │
  └── 任务接口 (原生协议)
        ├── vmagent: Prometheus scrape_config (YAML) + http_sd
        └── vmalert: Prometheus rule_groups (YAML)
```

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 资源拉取 | 按 zone_id 从管理平台拉取本区所有资源，按类型分组 |
| 健康检测 | 主动探测 + 被动心跳，三态流转 (healthy/degraded/offline) |
| 任务分配 | 一致性哈希将 collection_task / rc_rule 分配到 worker |
| 配置生成 | 为 vmagent 生成 scrape_config YAML，为 vmalert 生成 rule_groups YAML |
| 任务推送 | SSE 主动推送变更通知到 worker |
| Gossip 同步 | 多副本间同步资源列表、分配表、健康状态 |
| HTTP API | 供 worker 拉取任务、供管理平台推送更新 |

### 本模块不负责

| 不负责项 | 归属模块 |
|----------|----------|
| 采集任务/告警规则的 CRUD | 管理平台 T1/T2 |
| 实际的采集执行 | vmagent / Prometheus |
| 实际的规则评估 | vmalert |
| 数据存储 | DataSink (VictoriaMetrics / ES / S3...) |
| 告警通知路由 | 区 AM → 核心 AM → Kafka |

---

## 三、核心功能

### 3.1 资源拉取

Coordinator 启动时及周期性地向管理平台拉取本区资源。资源按类型分组管理。

**资源类型分类：**

| 类别 | 资源类型 | 说明 |
|------|----------|------|
| Worker | vmagent | 采集 worker，接收 scrape_config |
| Worker | vmalert | 规则评估 worker，接收 rule_groups |
| 存储 | victoriametrics | VictoriaMetrics 单节点或集群 |
| 存储 | prometheus | Prometheus 存储 |
| 存储 | opensearch | OpenSearch 日志存储 |
| 存储 | elasticsearch | Elasticsearch 日志存储 |
| 存储 | openobserve | OpenObserve 可观测平台 |
| 存储 | s3 | S3/MinIO 对象存储 |
| 管道 | otel_collector | OTel Collector 数据管道 |
| 目标 | exporter | 各类 exporter (node/mysql/redis/snmp...) |

**拉取模式：**

```
Coordinator 启动
  │
  ├── 从管理平台拉取全量资源: GET /api/v1/zones/{zone_id}/resources
  │
  ├── 按类型分组存入本地 store
  │
  └── 启动周期刷新 (默认 30s)
        │
        ├── 增量拉取: If-None-Match / ETag
        │
        └── 差异更新: 新增/更新/删除资源
```

**推送模式（管理平台主动通知）：**

```
管理平台资源变更
  │
  └── POST /api/v1/push/resources
        │
        └── Coordinator 接收并更新本地 store
              │
              └── 触发 gossip 广播 + 重新计算分配
```

### 3.2 健康检测

采用「主动探测 + 被动心跳」双机制，综合判定组件健康状态。

**主动探测：**

| 资源类型 | 探测端点 | 判定标准 | 周期 |
|----------|----------|----------|------|
| vmagent | `/-/healthy` | HTTP 200 | 15s |
| vmalert | `/-/healthy` | HTTP 200 | 15s |
| VictoriaMetrics | `/-/healthy` | HTTP 200 | 30s |
| Prometheus | `/-/healthy` | HTTP 200 | 30s |
| OTel Collector | `/health` | HTTP 200 | 30s |
| Exporter | `/metrics` | HTTP 200 | 30s |
| OpenSearch | `/_cluster/health` | HTTP 200 + status != red | 60s |
| S3 | HEAD bucket | HTTP 200 | 60s |

**被动心跳：**

Worker (vmagent/vmalert) 定期向 Coordinator 上报心跳：

```
POST /api/v1/workers/{worker_id}/heartbeat

{
  "timestamp": "2026-10-28T10:30:00Z",
  "status": "healthy",
  "metrics": {
    "task_count": 120,
    "uptime_seconds": 864000
  }
}
```

**状态流转：**

```
                  心跳超时(3次) 或 探测连续失败(3次)
unknown ──────────────────────────────> offline
   │                                       ↑
   │ 首次探测通过                            │ 持续不可达(300s)
   ↓                                       │
healthy <────────────────────────────> degraded
        探测恢复 + 心跳正常              探测失败或心跳超时
```

**健康检测结果上报：**

Coordinator 将健康状态汇总后上报管理平台：

```
POST {management_url}/api/v1/zones/{zone_id}/health-report

{
  "zone_id": "zone-east-1",
  "timestamp": "2026-10-28T10:30:00Z",
  "resources": [
    {"resource_id": "vm-001", "type": "victoriametrics", "status": "healthy", "latency_ms": 12},
    {"resource_id": "vmagent-001", "type": "vmagent", "status": "healthy", "task_count": 120}
  ]
}
```

### 3.3 任务协调分发

#### 3.3.1 一致性哈希分配

所有 worker (vmagent/vmalert) 注册到一致性哈希环上。每个 job (collection_task / rc_rule) 通过 `hash(job_id)` 定位到环上的 owner worker。

```
一致性哈希环 (160 虚拟节点/worker)

         hash(task-001) → slot 42 → vmagent-01
         hash(task-002) → slot 87 → vmagent-02
         hash(task-003) → slot 15 → vmagent-01
         hash(rule-001) → slot 103 → vmalert-01
         hash(rule-002) → slot 61 → vmalert-02
```

**分配特性：**

| 特性 | 说明 |
|------|------|
| 确定性 | 相同 job_id + 相同 worker 列表 → 相同分配结果 |
| 最小迁移 | worker 增减时，仅 1/N 的 job 需要迁移 |
| 负载均衡 | 160 虚拟节点保证均匀分布 |
| 去重 | vmalert 多副本跑相同规则，存储层 dedup |

#### 3.3.2 推送 + 拉取双模式

**推送（SSE）：** 任务分配变更时，Coordinator 主动通知受影响的 worker。

```
GET /api/v1/stream/assignments?worker_id=vmagent-01

event: assignment_changed
data: {"worker_id": "vmagent-01", "action": "full_sync"}

event: task_added
data: {"task_id": "task-003", "type": "collection"}
```

**拉取（HTTP）：** Worker 定期或按需拉取完整分配。

```
GET /api/v1/workers/{worker_id}/scrape-config
Accept: application/x-yaml

返回 Prometheus scrape_config YAML
```

```
GET /api/v1/workers/{worker_id}/rules
Accept: application/x-yaml

返回 Prometheus rule_groups YAML
```

#### 3.3.3 多副本一致性

多个 Coordinator 副本通过 gossip 同步以下状态：

| 同步项 | 同步方式 | 一致性保证 |
|--------|----------|------------|
| Worker 成员列表 | gossip 传播 | 最终一致 (O(log N) 收敛) |
| 资源列表 | 管理平台推送 + gossip 广播 | 所有副本收到相同推送 |
| 分配表 | 各副本独立计算 (确定性哈希) | 天然一致 (相同输入 → 相同输出) |
| 健康状态 | gossip 心跳 + 本地探测合并 | 最终一致 |

**关键设计：** 分配表不需要 gossip 同步——因为一致性哈希是确定性算法，只要所有副本的 worker 成员列表一致，各自独立计算出的分配表必然一致。Gossip 只负责保证成员列表的最终一致性。

#### 3.3.4 任务更新推送流程

```
管理平台推送任务更新到 Coordinator-A (leader)
  │
  ├── Coordinator-A 更新本地 store
  │
  ├── Coordinator-A 重新计算分配表
  │
  ├── Coordinator-A 通过 gossip 广播变更通知
  │     └── Coordinator-B/C 收到通知 → 重新拉取资源 → 重算分配
  │
  └── Coordinator-A 通过 SSE 推送变更到受影响的 worker
        └── worker 拉取新配置 → 热加载
```

### 3.4 任务接口（原生协议）

#### 3.4.1 vmagent: Prometheus scrape_config

Coordinator 为每个 vmagent worker 生成其负责采集的 scrape_config YAML：

```yaml
global:
  scrape_interval: 15s
  scrape_timeout: 10s

scrape_configs:
  - job_name: mysql-standard-inst-001
    scrape_interval: 15s
    metrics_path: /metrics
    static_configs:
      - targets: ["10.0.1.5:9104"]
        labels:
          instance: inst-001
          zone: zone-east-1
          __meta_task_id: task-001
    basic_auth:
      username: monitor
      password_file: /etc/vmagent/creds/task-001.password

  - job_name: linux-node-inst-002
    scrape_interval: 30s
    metrics_path: /metrics
    static_configs:
      - targets: ["10.0.1.6:9100"]
        labels:
          instance: inst-002
          zone: zone-east-1
          __meta_task_id: task-002

remote_write:
  - url: http://vmstorage-east:8428/api/v1/write
    queue_config:
      max_samples_per_send: 10000
```

vmagent 通过 `POST /-/reload` 热加载新配置。

#### 3.4.2 vmalert: Prometheus rule_groups

Coordinator 为每个 vmalert worker 生成其负责评估的 rule_groups YAML：

```yaml
groups:
  - name: zone-east-1-mysql-alerts
    interval: 1m
    rules:
      - alert: MySQLDown
        expr: mysql_up{instance="inst-001"} == 0
        for: 5m
        labels:
          severity: critical
          zone: zone-east-1
        annotations:
          summary: "MySQL 实例 inst-001 不可达"

      - alert: MySQLHighConnections
        expr: mysql_global_status_threads_connected > 100
        for: 10m
        labels:
          severity: warning
          zone: zone-east-1
        annotations:
          summary: "MySQL 连接数过高"

  - name: zone-east-1-linux-alerts
    interval: 30s
    rules:
      - alert: InstanceDown
        expr: up{job="linux-node"} == 0
        for: 3m
        labels:
          severity: warning
        annotations:
          summary: "实例 {{ $labels.instance }} 不可达"
```

vmalert 通过 `GET /-/reload` 热加载新规则。

---

## 四、HTTP API

### 4.1 Worker 端 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/workers/register` | Worker 注册到 Coordinator |
| POST | `/api/v1/workers/{id}/heartbeat` | Worker 心跳上报 |
| GET | `/api/v1/workers/{id}/assignments` | 拉取任务分配 (JSON) |
| GET | `/api/v1/workers/{id}/scrape-config` | vmagent 拉取 scrape_config (YAML) |
| GET | `/api/v1/workers/{id}/rules` | vmalert 拉取 rule_groups (YAML) |
| GET | `/api/v1/stream/assignments` | SSE 订阅变更推送 |

### 4.2 管理平台端 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/push/resources` | 推送资源列表更新 |
| POST | `/api/v1/push/tasks` | 推送采集任务更新 |
| POST | `/api/v1/push/rules` | 推送告警规则更新 |
| GET | `/api/v1/status` | Coordinator 状态查询 |
| GET | `/api/v1/assignments` | 查询全局分配视图 |

---

## 五、配置文件

```yaml
zone_id: "zone-east-1"

server:
  http_addr: ":8080"

gossip:
  bind_port: 7946
  seeds:
    - "10.0.1.1:7946"
    - "10.0.1.2:7946"
  node_name: "coordinator-east-1"

management:
  url: "http://management:8080"
  auth_token: ""
  sync_interval: "30s"

health:
  probe_interval: "15s"
  heartbeat_timeout: "90s"
  probe_timeout: "5s"
  fail_threshold: 3

distribution:
  hash_partitions: 160
  rebalance_debounce: "5s"
```

---

## 六、数据流全景

```
┌─────────────────────────────────────────────────────────────────────┐
│                        管理平台 (控制面)                              │
│                                                                     │
│  资源管理 ─── 采集任务(T1) ─── RC任务(T2)                            │
│       │              │               │                              │
│       └──────────────┼───────────────┘                              │
│                      │ 推送资源/任务/规则                              │
└──────────────────────┼──────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                  Zone Coordinator (区协调器)                         │
│                                                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │ 资源拉取  │→│ 健康检测  │→│ 分配计算  │→│ 配置生成  │          │
│  │ (pull)   │  │ (probe)  │  │ (hash)   │  │ (render) │          │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘          │
│       ↑                            │              │                │
│       │  gossip 同步               │              │                │
│       │                            │   SSE push   │                │
│  ┌──────────┐                      │              │                │
│  │ Gossip   │←─────────────────────┘              │                │
│  │ Protocol │                                     │                │
│  └──────────┘                                     │                │
│       │                                           │                │
└───────┼───────────────────────────────────────────┼────────────────┘
        │                                           │
        ▼                                           ▼
┌───────────────────┐                    ┌───────────────────┐
│   Coordinator Peer │                    │    Worker Nodes    │
│   (gossip 同步)    │                    │                    │
└───────────────────┘                    │  vmagent × N       │
                                         │    ← scrape_config │
                                         │                    │
                                         │  vmalert × N       │
                                         │    ← rule_groups   │
                                         └───────────────────┘
```

---

## 七、设计决策

### DEC-ZC-01: 分配表不 gossip 同步

**决策：** 各 Coordinator 副本独立计算分配表，不通过 gossip 传播分配结果。

**理由：** 一致性哈希是确定性算法。只要 gossip 保证了 worker 成员列表的最终一致性，各副本独立计算的分配表必然收敛到相同结果。避免了 gossip 传播分配表的一致性问题。

### DEC-ZC-02: 推送 + 拉取双模式

**决策：** 任务变更时 SSE 推送通知，worker 收到通知后主动拉取完整配置。

**理由：**
- 纯推送模式下 worker 断连期间会丢失变更。
- 纯拉取模式下变更感知延迟 = 拉取间隔。
- 推送通知 + 拉取配置的组合兼顾实时性和可靠性。

### DEC-ZC-03: 原生协议下发

**决策：** vmagent 使用 Prometheus scrape_config YAML，vmalert 使用 Prometheus rule_groups YAML。

**理由：**
- 两种格式都是 Prometheus 生态标准。
- vmagent / vmalert 原生支持，无需额外适配层。
- 配置可人工审查和版本管理。

### DEC-ZC-04: 资源拉取而非纯推送

**决策：** Coordinator 主动从管理平台拉取资源，同时接受管理平台的推送更新。

**理由：**
- 纯推送模式下 Coordinator 重启后需要管理平台重新推送全量。
- 拉取模式保证 Coordinator 启动时能自行恢复完整状态。
- 推送作为加速通道，减少拉取延迟。
