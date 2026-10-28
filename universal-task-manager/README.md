# 多网区监控采集管理平台

> 版本：v3.0 | 日期：2026-10-27
> 定位：面向社区开源的通用多网区监控采集管理平台，覆盖多业务系统、多协议多存储、资源管理、任务调度、告警通知全链路

---

## 一、设计目标

| # | 目标 | 说明 |
|---|------|------|
| G1 | 多网区统一管理 | 网区 + 网段，自动推荐与冲突检测 |
| G2 | 资源类型可扩展 | 树形类型体系，父子继承公共指标 |
| G3 | 采集任务动态扩展 | 注册新类型即新增表单/详情页/指标 |
| G4 | 告警模块可替换 | 告警层独立，通过接口对接，默认 alertmanager -> kafka |
| G5 | 采集器高可用 | 多节点健康检测 + 动态任务分配 |
| G6 | 用户多源认证 | 本地 / LDAP / SSO，关联多渠道通知 |
| G7 | 概览实时 | 定时状态聚合到 Redis，支撑仪表盘 |
| G8 | 多协议多存储 | 支持 metric/log/trace 多协议，网区内可对接多种存储后端 |
| G9 | 模块化可拆分 | 服务发现、负载均衡、定时任务独立模块，按需组合 |
| G10 | 多业务系统 | 资源按业务系统隔离，支持全览视图与权限控制 |

---

## 二、系统架构

```
+------------------------------------------------------------------------------+
|                          管理平台 (Management Plane)                          |
|                                                                              |
|  +----------------------------------------------------------------------+   |
|  |                       业务系统层 (Business System)                     |   |
|  |  +------------+  +------------+                                      |   |
|  |  | 业务系统   |  | 成员与权限 |                                       |   |
|  |  |   B1       |  |   B1.2     |                                      |   |
|  |  +------------+  +------------+                                      |   |
|  +----------------------------------------------------------------------+   |
|                                                                              |
|  +----------------------------------------------------------------------+   |
|  |                         核心资源层                                    |   |
|  |  +----------+ +----------+ +----------+ +----------+ +----------+   |   |
|  |  |网区与网段 | |资源类型树 | | 实例管理  | | 凭据管理  | |数据存储   |   |   |
|  |  |  M1      | |  M2      | |  M3      | |  M4      | |  M5      |   |   |
|  |  +----------+ +----------+ +----------+ +----------+ +----------+   |   |
|  +----------------------------------------------------------------------+   |
|                                                                              |
|  +----------------------------------------------------------------------+   |
|  |                         任务管理层                                    |   |
|  |  +----------+ +----------+ +----------+ +----------+ +----------+   |   |
|  |  | 采集任务  | | RC任务   | | 指标管理  | |服务发现   | |定时任务   |   |   |
|  |  |  T1      | | T2       | | T3       | | T4       | | T5       |   |   |
|  |  +----------+ +----------+ +----------+ +----------+ +----------+   |   |
|  +----------------------------------------------------------------------+   |
|                                                                              |
|  +----------------------------------------------------------------------+   |
|  |                    告警层 (可替换 / Pluggable)                        |   |
|  |  +----------+ +----------+ +----------+ +----------+                 |   |
|  |  | 告警策略  | | 告警组   | | 通知渠道  | | 告警收敛  |                 |   |
|  |  |  A1      | | A2       | | A3       | | A4       |                |   |
|  |  +----------+ +----------+ +----------+ +----------+                 |   |
|  +----------------------------------------------------------------------+   |
|                                                                              |
|  +----------------------------------------------------------------------+   |
|  |                         基础设施层                                    |   |
|  |  +----------+ +----------+ +----------+                               |   |
|  |  |采集器管理 | | 用户管理  | | 状态缓存  |                               |   |
|  |  | I1       | | I2       | | I3       |                              |   |
|  |  +----------+ +----------+ +----------+                               |   |
|  +----------------------------------------------------------------------+   |
|                                                                              |
+------------------------------------------------------------------------------+
                              |
                +-------------+-------------+
                |             |             |
                v             v             v
        +--------------+ +--------------+ +--------------+
        |   Zone-A     | |   Zone-B     | |   Zone-C     |
        |              | |              | |              |
        | 区协调器     | | 区协调器     | | 区协调器     |
        | (gossip)     | | (gossip)     | | (gossip)     |
        | 协调:        | | 协调:        | | 协调:        |
        |  vmalert × N | |  vmalert × N | |  vmalert     |
        |  vmagent × N | |  vmagent × N | | prometheus ×N|
        | 区AM(去重)   | | 区AM(去重)   | | 区AM(去重)   |
        | DataSink:    | | DataSink:    | | DataSink:    |
        |  VM + ES     | |  VM + Loki   | |  VM + S3     |
        +------+-------+ +------+-------+ +------+-------+
               |                |                |
               +----------------+----------------+
                                |
                                v
                    +-----------------------+
                    |   核心AM (策略路由)    |
                    | 分组/抑制/静默/升级    |
                    +-----------+-----------+
                                |
                                v
                          +----------+
                          |  Kafka   |
                          +----+-----+
                               |
                               v
                          +----------+
                          | 消费端   |
                          |(平台/IM) |
                          +----------+
```

---

## 三、核心概念

### 3.1 概念全景

```
BusinessSystem (业务系统)
  |-- Zone (网区) x N
  |     |-- NetworkSegment (网段) x N
  |     |-- ZoneCoordinator (区协调器, gossip) x 1
  |     |     └── 协调: vmalert × N, vmagent × N, prometheus × N 的 job 分配
  |     |-- CollectorNode (采集器/vmalert 节点) x N
  |     |-- DataSink (数据存储) x N
  |     |
  |     +-- Instance (实例)
  |           |-- ResourceType (资源类型，来自类型树)
  |           |-- CollectionTask (采集任务) x N
  |           |     |-- TaskType (任务类型定义)
  |           |     +-- Metric (产出指标) x N
  |           |
  |           |-- RCTask (RC 任务) x N
  |           |     +-- AlertStrategy (告警策略) x N
  |           |           |-- MetricRef (关联指标)
  |           |           |-- AlertGroup (告警组)
  |           |           +-- NotificationChannel (通知渠道)
  |           |
  |           +-- InstanceAssociation (实例关联，通过 IP)
  |
  |-- ServiceDiscovery (服务发现) x N
  |-- ScheduledTask (定时任务) x N
  +-- BusinessSystemMember (成员) x N
```

### 3.2 概念定义

| 概念 | 说明 | 示例 |
|------|------|------|
| **BusinessSystem (业务系统)** | 顶层组织单元，隔离资源与权限 | 电商平台、支付系统、内部OA |
| **Zone (网区)** | 物理或逻辑的监控部署域 | 华东区、北京区、DMZ区 |
| **NetworkSegment (网段)** | 分配给网区的 IP 范围 | 10.0.1.0/24 |
| **DataSink (数据存储)** | 网区内的数据接收端，支持多类型 | VictoriaMetrics、Elasticsearch、OpenObserve、S3 |
| **ResourceType (资源类型)** | 树形结构的被监控资源分类 | 基础设施 > Linux > CentOS |
| **Instance (实例)** | 被监控的具体实体 | 10.0.1.5 的 MySQL 实例 |
| **CollectionTask (采集任务)** | 绑定到实例的采集配置 | "每15s采集MySQL指标" |
| **RCTask (RC任务)** | 按网区拆分的规则检查任务 | "Zone-A 的 MySQL 延迟检查" |
| **Metric (指标)** | PromQL 模板 + 变量，分事件/指标两类 | `mysql_up{instance="$host"}` |
| **ServiceDiscovery (服务发现)** | 动态目标发现机制 | http_sd、file_sd、dns_sd、consul_sd |
| **ScheduledTask (定时任务)** | 周期性执行的后台任务 | 状态聚合、数据清理、报表生成 |
| **AlertStrategy (告警策略)** | 实例 x 策略 x 指标 -> 网区告警规则 | "MySQL 连接数 > 100 持续 5m" |
| **AlertGroup (告警组)** | 告警分组与路由目标 | "DBA 值班组" |
| **NotificationChannel (通知渠道)** | 通知投递方式 | 邮件、短信、飞书、钉钉、微信 |
| **ZoneCoordinator (区协调器)** | 每区一个，运行 gossip 协议，协调区内所有 job 分配 | vmalert/vmagent/prometheus 的负载分配与故障接管 |
| **CollectorNode (采集/规则节点)** | 实际执行采集或规则评估的进程 | vmagent-1, vmalert-1, prometheus-2 |

---

## 四、模块清单

### 4.1 业务系统层

| # | 模块 | 文档 | 核心职责 |
|---|------|------|----------|
| B1 | 业务系统管理 | [business/business-system.md](business/business-system.md) | 业务系统 CRUD、成员管理、资源隔离、全览视图 |

### 4.2 核心资源层

| # | 模块 | 文档 | 核心职责 |
|---|------|------|----------|
| M1 | 网区与网段管理 | [resource/zone-segment.md](resource/zone-segment.md) | 网区 CRUD、网段分配、自动推荐、冲突检测 |
| M2 | 资源类型树 | [resource/resource-type-tree.md](resource/resource-type-tree.md) | 树形类型定义、父子继承、公共指标配置 |
| M3 | 实例管理 | [resource/instance-management.md](resource/instance-management.md) | 实例台账、IP 关联、类型关联、多 Sheet 详情页 |
| M4 | 凭据管理 | [resource/credential.md](resource/credential.md) | 凭据加密存储、引用分发 |
| M5 | 数据存储管理 | [resource/data-sink.md](resource/data-sink.md) | 多类型存储注册、协议适配、写入路由 |

### 4.3 任务管理层

| # | 模块 | 文档 | 核心职责 |
|---|------|------|----------|
| T1 | 采集任务管理 | [task/collection-task.md](task/collection-task.md) | 动态任务类型、表单渲染、配置生成 |
| T2 | RC 任务管理 | [task/rc-task.md](task/rc-task.md) | 规则包生成发布、按网区拆分、下发到区协调器分配 |
| T3 | 指标管理 | [task/metric-management.md](task/metric-management.md) | PromQL 模板、变量替换、事件/指标分类 |
| T4 | 服务发现 | [task/service-discovery.md](task/service-discovery.md) | http_sd/file_sd/dns_sd/consul_sd 多协议支持 |
| T5 | 定时任务 | [task/scheduled-task.md](task/scheduled-task.md) | 周期性后台任务调度、执行记录、失败重试 |

### 4.4 告警层（可替换）

| # | 模块 | 文档 | 核心职责 |
|---|------|------|----------|
| A1 | 告警策略 | [alert/alert-strategy.md](alert/alert-strategy.md) | 实例 x 策略 x 指标 -> 按网区下发规则 |
| A2 | 告警组 | [alert/alert-group.md](alert/alert-group.md) | 告警分组、路由、认领 |
| A3 | 通知渠道 | [alert/notification-channel.md](alert/notification-channel.md) | 邮件/短信/飞书/钉钉/微信投递 |
| A4 | 告警收敛 | [alert/alert-convergence.md](alert/alert-convergence.md) | 两级 AM 管理(区AM去重/核心AM策略)、AlertEngine 可替换接口 |

### 4.5 基础设施层

| # | 模块 | 文档 | 核心职责 |
|---|------|------|----------|
| I1 | 采集器与协调器管理 | [infra/collector-management.md](infra/collector-management.md) | 区协调器(gossip)管理、节点注册、健康检测、job 分配协调 |
| I2 | 用户管理 | [infra/user-management.md](infra/user-management.md) | 本地/LDAP/SSO、关联通知目标 |
| I3 | 状态缓存 | [infra/status-cache.md](infra/status-cache.md) | 定时聚合状态到 Redis、概览仪表盘 |

---

## 五、模块间关系

### 5.1 数据流

```
用户在管理平台操作
    |
    +-- 创建业务系统 -> 分配网区 -> 添加成员 -> 设置权限
    |
    +-- 注册网区 -> 分配网段 -> 自动推荐 / 冲突检测
    |
    +-- 注册数据存储 -> 选择类型(VM/ES/Loki/S3) -> 配置连接参数
    |
    +-- 定义资源类型树 -> 父子继承公共指标
    |
    +-- 注册实例 -> 关联资源类型 -> 按 IP 匹配网区
    |
    +-- 创建采集任务 -> 选择任务类型 -> 填充表单 -> 保存到管理平台
    |                                    └── 区协调器监听变更, gossip 分配到 vmagent/prometheus
    |
    +-- 配置服务发现 -> 选择协议(http_sd/file_sd/dns_sd) -> 动态注册目标
    |
    +-- 创建 RC 任务 -> 按网区拆分 -> 保存到管理平台
    |                    └── 区协调器监听变更, gossip 一致性哈希分配到 vmalert 节点
    |
    +-- 定义告警策略 -> 实例 x 策略 x 指标 -> 按网区下发到 vmalert
    |
    +-- 告警触发 -> 区AM去重 -> 核心AM策略路由 -> Kafka -> 通知渠道 -> 用户
    |
    +-- 定时任务 -> 状态聚合 / 数据清理 / 报表生成 -> 执行记录
```

### 5.2 模块依赖

```
                    +----------------------------------+
                    |         业务系统层               |
                    |                                  |
                    |  B1 业务系统管理                  |
                    |  (资源隔离 / 成员权限 / 全览)     |
                    +----------------+-----------------+
                                     |
                    +----------------+-----------------+
                    |          核心资源层               |
                    |                                  |
                    |  M1 网区网段 <--> M2 资源类型树  |
                    |       |                |         |
                    |       v                v         |
                    |  M3 实例管理 --> M4 凭据管理     |
                    |       |                          |
                    |       +--> M5 数据存储管理        |
                    +----------------+-----------------+
                                     |
            +------------------------+------------------------+
            |                        |                        |
            v                        v                        v
  +------------------+   +------------------+   +------------------+
  |   任务管理层      |   |   告警层         |   |  基础设施层       |
  |                  |   |                  |   |                  |
  | T1 采集任务      |   | A1 告警策略      |   | I1 采集器与      |
  | T2 RC 任务       |   | A2 告警组        |   |    协调器管理    |
  | T3 指标管理      |   | A3 通知渠道      |   |    (gossip)     |
  | T4 服务发现      |   | A4 告警收敛      |   | I2 用户          |
  | T5 定时任务      |   |   (两级AM)      |   | I3 状态缓存      |
  +------------------+   +------------------+   +------------------+

  T1/T2 发布 job → 区协调器(I1) gossip 分配 → vmalert/vmagent/prometheus 执行
```

---

## 六、多协议多存储设计

### 6.1 数据存储抽象

网区内可注册多种数据存储后端，每种存储通过 DataSinkAdapter 接口适配。采集任务可指定目标存储，未指定时使用网区默认存储。

```go
// DataSinkAdapter 数据存储适配器接口（可替换）
type DataSinkAdapter interface {
    // Type 返回存储类型标识
    Type() string

    // Validate 校验连接配置
    Validate(config DataSinkConfig) error

    // Write 写入数据（metric/log/trace）
    Write(data []DataPoint) error

    // Query 查询数据（用于告警规则评估）
    Query(query string, start, end time.Time) ([]DataPoint, error)

    // HealthCheck 检查存储健康状态
    HealthCheck() error
}
```

### 6.2 内置存储适配器

| 存储类型 | 协议 | 数据类别 | 说明 |
|----------|------|----------|------|
| **VictoriaMetrics** | Prometheus remote_write | metric | 高性能时序数据库，兼容 PromQL |
| **Elasticsearch** | HTTP bulk API | log/trace | 全文检索，适合日志与链路追踪 |
| **Loki** | HTTP push API | log | 轻量级日志聚合，标签索引 |
| **OpenObserve** | HTTP API | metric/log/trace | 统一可观测平台，多协议支持 |
| **S3/MinIO** | S3 API | metric/log/archive | 对象存储，适合冷数据归档 |
| **Custom** | 自定义 | 任意 | 用户实现 DataSinkAdapter 接口 |

### 6.3 数据路由

```
采集任务 (CollectionTask)
    │
    ├── 指定 data_sink_id → 写入对应存储
    │
    └── 未指定 → 使用网区默认存储 (zone.default_sink_id)

多存储场景:
    Zone-A
    ├── DataSink: vm-prod (VictoriaMetrics, 默认)
    ├── DataSink: es-logs (Elasticsearch)
    └── DataSink: s3-archive (S3)

    采集任务1 (MySQL指标) → vm-prod
    采集任务2 (应用日志) → es-logs
    采集任务3 (历史归档) → s3-archive
```

---

## 七、多业务系统设计

### 7.1 业务系统作为顶层组织

业务系统（BusinessSystem）是资源的顶层隔离单元。每个网区、实例、任务都归属于某个业务系统。

```
BusinessSystem: 电商平台
  ├── Zone: 华东区
  ├── Zone: 华南区
  ├── 成员: alice (admin), bob (operator)
  └── 资源: 200 实例, 600 采集任务

BusinessSystem: 支付系统
  ├── Zone: 北京区
  ├── 成员: charlie (admin), dave (viewer)
  └── 资源: 50 实例, 150 采集任务
```

### 7.2 权限模型

| 角色 | 权限范围 | 说明 |
|------|----------|------|
| **admin** | 业务系统内全部资源 | 可管理成员、配置、告警 |
| **operator** | 业务系统内操作权限 | 可创建/修改任务，不可管理成员 |
| **viewer** | 业务系统内只读权限 | 可查看资源与告警，不可修改 |

### 7.3 全览视图

用户可同时属于多个业务系统。平台提供「全览」模式，聚合展示用户有权限的所有业务系统资源：

```
全览视图 (用户 alice 属于 电商平台 + 支付系统)
  ├── 电商平台
  │     ├── 华东区: 120 实例, 360 任务, 2 告警
  │     └── 华南区: 80 实例, 240 任务, 0 告警
  │
  └── 支付系统
        └── 北京区: 50 实例, 150 任务, 1 告警

  汇总: 250 实例, 750 任务, 3 告警
```

---

## 八、服务发现与定时任务

### 8.1 服务发现 (T4)

服务发现模块提供动态目标注册能力，支持多种协议：

| 协议 | 说明 | 适用场景 |
|------|------|----------|
| **http_sd** | HTTP 接口暴露目标列表 | 平台内置，采集器定期拉取 |
| **file_sd** | 文件系统目标列表 | 本地配置，适合静态环境 |
| **dns_sd** | DNS SRV/A 记录发现 | 微服务环境 |
| **consul_sd** | Consul 服务目录 | Consul 注册中心 |
| **kubernetes_sd** | K8s API 发现 | 容器化部署 |

```
服务发现配置:
  discovery_id: "sd-web-cluster"
  protocol: "http_sd"
  refresh_interval: "30s"
  targets:
    - labels: {cluster: "web", env: "prod"}
      endpoint: "http://consul:8500/v1/catalog/service/web"
```

### 8.2 定时任务 (T5)

定时任务模块管理周期性后台作业，与采集任务解耦：

| 任务类型 | 说明 | 默认周期 |
|----------|------|----------|
| **status_aggregate** | 聚合实例/节点状态到 Redis | 30s |
| **data_cleanup** | 清理过期执行日志与事件 | 24h |
| **report_generate** | 生成日报/周报 | 1h |
| **health_check** | 检查存储/采集器健康 | 60s |
| **custom** | 用户自定义定时任务 | 自定义 |

```
定时任务配置:
  task_id: "cleanup-logs"
  task_type: "data_cleanup"
  cron: "0 2 * * *"  # 每天凌晨2点
  config:
    retention_days: 30
    target_tables: ["task_execution_log", "alert_event"]
  enabled: true
```

---

## 九、告警层可替换设计

告警层通过明确的接口边界与管理平台解耦，支持后期替换。

### 9.1 默认实现（两级 Alertmanager）

```
vmalert (各区) → 区AM (去重) → 核心AM (策略路由) → Kafka → 通知调度 → IM/邮件/短信
```

| 层级 | 组件 | 职责 | 部署 |
|------|------|------|------|
| 区级 | vmalert | 规则评估，产生原始告警 | 每区独立部署 |
| 区级 | alertmanager | 区内去重、fingerprint 消重 | 每区单实例，不集群 |
| 核心 | alertmanager | 全局策略路由：分组/抑制/静默/升级/分发 | 中心部署，可集群 |

**区 AM 只做去重**：同一区内多个 vmalert 可能产生重复告警（相同 fingerprint），区 AM 负责消重后转发到核心 AM。区 AM 不配置 group/inhibit/silence 策略，保持轻量。

**核心 AM 承载策略**：接收所有区 AM 转发的告警，执行分组（group_by）、抑制（inhibit）、静默（silence）、路由分发到通知渠道。核心 AM 拥有全局视图，可跨区关联告警。

### 9.2 替换接口

```go
// AlertEngine 告警引擎接口（可替换）
type AlertEngine interface {
    // PushRules 下发告警规则到指定区的 vmalert
    PushRules(zoneID string, rules []AlertRule) error
    // RemoveRules 移除指定区的告警规则
    RemoveRules(zoneID string, ruleIDs []string) error
    // QueryAlerts 查询当前活跃告警（从核心AM）
    QueryAlerts(filter AlertFilter) ([]Alert, error)
    // Acknowledge 告警认领
    Acknowledge(alertID string, userID string) error
    // Silence 告警静默（核心AM级别）
    Silence(filter AlertFilter, duration time.Duration, reason string) error
}

// ZoneAlertmanager 区级去重接口
type ZoneAlertmanager interface {
    // Forward 将去重后的告警转发到核心AM
    Forward(alerts []Alert) error
    // Status 查询区AM状态
    Status() AlertmanagerStatus
}

// NotificationDispatcher 通知投递接口（可替换）
type NotificationDispatcher interface {
    // Send 发送通知
    Send(channel NotificationChannel, payload NotificationPayload) error
    // Validate 校验渠道配置
    Validate(channel NotificationChannel) error
}
```

### 9.3 替换路径

| 组件 | 默认实现 | 可替换为 |
|------|----------|----------|
| 规则评估 | vmalert | Prometheus rules / Thanos Ruler / 自定义 |
| 区级去重 | alertmanager (单实例) | Cortex alertmanager / 自定义去重服务 |
| 核心策略路由 | alertmanager (集群) | Cortex alertmanager / PagerDuty AIOps / 自定义 |
| 消息队列 | Kafka | RabbitMQ / Pulsar / Redis Streams |
| 通知投递 | 平台内置 | Grafana OnCall / PagerDuty / OpsGenie |

---

## 十、文件结构

```
universal-task-manager/
|-- README.md                              # 本文档
|-- design/
|   |-- data-model.md                      # 数据库表结构
|   |-- api-spec.md                        # API 规范
|   |-- collector-adapter.md               # 采集器适配器
|   +-- extension-guide.md                 # 扩展开发指南
|-- business/
|   +-- business-system.md                 # B1 业务系统管理
|-- resource/
|   |-- zone-segment.md                    # M1 网区与网段
|   |-- resource-type-tree.md              # M2 资源类型树
|   |-- instance-management.md             # M3 实例管理
|   |-- credential.md                      # M4 凭据管理
|   +-- data-sink.md                       # M5 数据存储管理
|-- task/
|   |-- collection-task.md                 # T1 采集任务
|   |-- rc-task.md                         # T2 RC 任务
|   |-- metric-management.md               # T3 指标管理
|   |-- service-discovery.md               # T4 服务发现
|   +-- scheduled-task.md                  # T5 定时任务
|-- alert/
|   |-- alert-strategy.md                  # A1 告警策略
|   |-- alert-group.md                     # A2 告警组
|   |-- notification-channel.md            # A3 通知渠道
|   +-- alert-convergence.md               # A4 告警收敛
|-- infra/
|   |-- collector-management.md            # I1 采集器与协调器管理
|   |-- user-management.md                 # I2 用户管理
|   +-- status-cache.md                    # I3 状态缓存
+-- examples/
    |-- mysql-task-type.json               # MySQL 任务类型示例
    |-- linux-task-type.json               # Linux 任务类型示例
    |-- custom-sql-type.json               # 自定义SQL示例
    |-- resource-type-tree-example.json    # 资源类型树示例
    +-- data-sink-config.json              # 数据存储配置示例
```

---

## 十一、分期策略

### Phase 1 -- 核心骨架

- B1 业务系统管理（基础 CRUD + 成员）
- M1 网区与网段管理
- M2 资源类型树（基础版，无继承）
- M3 实例管理（基础 CRUD）
- M5 数据存储管理（单存储，VictoriaMetrics）
- T1 采集任务管理（动态类型注册）
- I1 采集器管理（单节点注册 + 健康检测）
- I3 状态缓存（Redis 基础聚合）

### Phase 2 -- 任务与指标

- T2 RC 任务管理（按网区拆分）
- T3 指标管理（PromQL 模板 + 变量替换）
- T4 服务发现（http_sd + file_sd）
- T5 定时任务（状态聚合 + 数据清理）
- M2 资源类型树（父子继承）
- M3 实例管理（IP 关联 + 多 Sheet）
- M5 数据存储管理（多存储支持）
- I1 采集器管理（多节点任务分配）

### Phase 3 -- 告警与通知

- A1 告警策略（实例 x 策略 x 指标 -> 网区下发）
- A2 告警组
- A3 通知渠道（邮件/飞书/钉钉/微信）
- A4 告警收敛（alertmanager -> Kafka）
- T4 服务发现（dns_sd + consul_sd + kubernetes_sd）

### Phase 4 -- 用户与增强

- I2 用户管理（本地 -> LDAP -> SSO）
- M4 凭据管理（加密存储 + 引用分发）
- B1 业务系统管理（全览视图 + 细粒度权限）
- I1 采集器管理（gossip 协议扩展）
- 告警层替换接口完善

---

## 十二、设计原则

| # | 原则 | 说明 |
|---|------|------|
| P1 | 通用优先 | 核心模型不绑定特定采集器、存储或告警引擎 |
| P2 | 接口隔离 | 告警层、存储层、通知层通过接口解耦，可独立替换 |
| P3 | 声明式扩展 | 新任务类型通过 JSON 注册，无需改代码 |
| P4 | 网区自治 | 各网区独立运行，管理平台不可用时不中断采集 |
| P5 | 渐进增强 | 分期交付，Phase 1 即可独立使用 |
| P6 | 社区友好 | 标准协议（Prometheus/OTel）、标准接口、标准存储 |
| P7 | 业务隔离 | 多业务系统间资源与权限完全隔离，全览视图按需聚合 |
| P8 | 多协议统一 | 同一平台管理 metric/log/trace，存储层统一抽象 |
