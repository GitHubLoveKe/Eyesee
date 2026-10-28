# I1 采集器与协调器管理

> 版本：v1.1 | 日期：2026-10-27
> 模块编号：I1 | 职责：区协调器(ZoneCoordinator)管理、采集器/规则节点注册、健康检测、状态管理。区协调器运行 gossip 协议，统一协调本区内 vmalert/vmagent/prometheus 的 job 分配。

---

## 一、概述

采集器与协调器管理模块是基础设施层的核心组件，负责两件事：

1. **区协调器管理**：每个网区部署一个区协调器进程（可主备），运行 gossip 协议，统一协调本区内所有 job 的分配——包括 vmalert 规则分片、vmagent 采集任务分配、prometheus scrape job 分配。
2. **节点管理**：管理所有采集器/规则节点（CollectorNode）的注册、健康检测、状态流转。节点向区协调器注册，协调器通过一致性哈希将 job 分配到各节点。

```
管理平台 (控制面)
  │
  │  I1 管理: 区协调器注册/状态, 节点注册/状态
  │
  ▼
区协调器 × N (每区一个, gossip)
  │
  │  协调本区所有 job 分配:
  │  ├── vmalert × N: 规则一致性哈希分片
  │  ├── vmagent × N: 采集任务一致性哈希分配
  │  └── prometheus × N: scrape job 一致性哈希分配
  │
  ▼
执行节点 (vmalert, vmagent, prometheus)
```

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 区协调器管理 | 区协调器的注册、主备切换、状态监控 |
| Gossip 协议管理 | 区协调器运行 gossip，管理成员列表、一致性哈希分配 |
| 节点注册与发现 | 采集器/规则节点的 CRUD，关联所属网区，向区协调器注册 |
| 健康检测 | 定时心跳检测 + health_endpoint 探活 |
| 状态管理 | healthy / degraded / offline 三态流转与事件通知 |
| Job 分配策略 | 区协调器通过一致性哈希将 job 分配到各节点 |
| 容量管理 | owned_rules/task_count 统计、超载检测与告警 |
| 多类型节点支持 | vmagent、prometheus、vmalert 的差异化管理 |

### 本模块不负责

| 不负责项 | 归属模块 |
|----------|----------|
| 规则包生成与发布 | T2 RC 任务管理 |
| 采集任务本身的 CRUD | T1 采集任务管理 |
| 网区管理 | M1 网区与网段管理 |
| 服务发现目标列表生成 | T4 服务发现 |
| 状态数据缓存到 Redis | I3 状态缓存 |
| 告警策略定义 | A1 告警策略 |

---

## 三、功能清单

| # | 功能 | 优先级 | 阶段 | 说明 |
|---|------|--------|------|------|
| F1 | 区协调器注册 | P0 | Phase 1 | 注册区协调器到指定网区，记录地址、gossip端口 |
| F2 | 区协调器主备管理 | P1 | Phase 2 | 主备切换、standby 状态监控 |
| F3 | Gossip 成员管理 | P0 | Phase 1 | 节点向协调器注册，协调器维护 gossip 成员列表 |
| F4 | 一致性哈希分配 | P0 | Phase 1 | 协调器按 hash(job_id) mod N 分配 job 到节点 |
| F5 | 节点注册 | P0 | Phase 1 | 注册采集器/规则节点到指定网区，记录类型、地址 |
| F6 | 节点信息维护 | P0 | Phase 1 | 修改节点名称、地址、标签等属性 |
| F7 | 节点注销 | P0 | Phase 1 | 注销节点前自动触发 job 迁移到其他节点 |
| F8 | 心跳检测 | P0 | Phase 1 | 定期接收节点心跳，更新 last_heartbeat |
| F9 | 健康端点检查 | P0 | Phase 1 | 主动探测 health_endpoint，判定节点健康状态 |
| F10 | 状态流转 | P0 | Phase 1 | healthy -> degraded -> offline 自动流转与事件触发 |
| F11 | Job 再平衡 | P1 | Phase 2 | 节点上线/下线时触发一致性哈希重算，迁移 job |
| F12 | 容量监控 | P1 | Phase 2 | owned_rules/task_count 统计，超载阈值告警 |
| F13 | 节点列表查询 | P0 | Phase 1 | 按网区、状态、类型筛选节点列表 |
| F14 | 节点详情查询 | P0 | Phase 1 | 查看单节点完整信息、关联 job 列表 |

---

## 四、核心数据模型

### 4.1 collector_node（采集器节点表）

> 引用 data-model.md 5.1 节定义

```sql
CREATE TABLE collector_node (
    node_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    node_type       VARCHAR(32) NOT NULL COMMENT 'vmagent/prometheus/alloy',
    address         VARCHAR(256) NOT NULL COMMENT 'http://ip:port',
    health_endpoint VARCHAR(256) DEFAULT '/-/healthy',
    labels          JSON,
    status          VARCHAR(32) DEFAULT 'unknown' COMMENT 'healthy/degraded/offline',
    task_count      INT DEFAULT 0,
    last_heartbeat  TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_cn_zone ON collector_node(zone_id);
CREATE INDEX idx_cn_status ON collector_node(status);
```

### 4.2 关联表：collection_task（采集任务表，引用 node_id）

> 引用 data-model.md 3.2 节定义，仅列出关联字段

```sql
-- collection_task.node_id 指向 collector_node.node_id
-- 表示该采集任务当前分配到的采集器节点
ALTER TABLE collection_task ADD CONSTRAINT fk_ct_node
    FOREIGN KEY (node_id) REFERENCES collector_node(node_id) ON DELETE SET NULL;
```

### 4.3 节点容量配置表（扩展）

```sql
CREATE TABLE collector_node_capacity (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    node_type       VARCHAR(32) NOT NULL COMMENT 'vmagent/prometheus/alloy',
    max_tasks       INT NOT NULL DEFAULT 500 COMMENT '单节点最大任务数',
    warn_threshold  FLOAT NOT NULL DEFAULT 0.8 COMMENT '超载告警阈值(0-1)',
    critical_threshold FLOAT NOT NULL DEFAULT 0.95 COMMENT '严重超载阈值(0-1)',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_node_type (node_type)
);
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| M1 网区管理 | zone_id 有效性校验、网区信息 | 内部函数调用 |
| T1 采集任务管理 | 任务创建/删除时通知节点分配/释放 | 事件驱动 |
| T5 定时任务 | 健康检查定时调度、心跳超时检测调度 | 定时触发 |

### 5.2 下游提供

| 消费方 | 提供内容 | 交互方式 |
|--------|----------|----------|
| T1 采集任务管理 | 可用节点列表、节点当前负载 | 内部函数调用 |
| 采集器适配器 | 节点地址、类型、健康端点 | 内部函数调用 |
| I3 状态缓存 | 节点状态变更事件 | 事件驱动 |
| A1 告警策略 | 节点超载/离线告警事件 | 事件驱动 |
| 前端仪表盘 | 节点列表、状态、负载信息 | REST API |

### 5.3 对外 API

#### 5.3.1 注册节点

```
POST /api/v1/zones/{zone_id}/collector-nodes
```

**请求体：**

```json
{
  "node_id": "node-vmagent-east-01",
  "name": "华东区 vmagent-01",
  "node_type": "vmagent",
  "address": "http://10.0.1.100:8429",
  "health_endpoint": "/-/healthy",
  "labels": {
    "env": "prod",
    "dc": "east-1"
  }
}
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
    "status": "unknown",
    "task_count": 0,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.2 查询节点列表

```
GET /api/v1/zones/{zone_id}/collector-nodes
```

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| node_type | string | 按采集器类型过滤：vmagent/prometheus/alloy |
| status | string | 按状态过滤：healthy/degraded/offline |
| page | int | 页码 |
| page_size | int | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "node_id": "node-vmagent-east-01",
        "name": "华东区 vmagent-01",
        "node_type": "vmagent",
        "status": "healthy",
        "task_count": 120,
        "max_tasks": 500,
        "load_ratio": 0.24,
        "last_heartbeat": "2026-10-27T10:30:00Z",
        "address": "http://10.0.1.100:8429"
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询节点详情

```
GET /api/v1/collector-nodes/{node_id}
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
    "address": "http://10.0.1.100:8429",
    "health_endpoint": "/-/healthy",
    "labels": {"env": "prod", "dc": "east-1"},
    "status": "healthy",
    "task_count": 120,
    "last_heartbeat": "2026-10-27T10:30:00Z",
    "assigned_tasks": [
      {"task_id": "task-001", "instance_id": "inst-001", "status": "running"},
      {"task_id": "task-002", "instance_id": "inst-002", "status": "running"}
    ],
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T10:30:00Z"
  }
}
```

#### 5.3.4 更新节点信息

```
PUT /api/v1/collector-nodes/{node_id}
```

**请求体：**

```json
{
  "name": "华东区 vmagent-01 (updated)",
  "address": "http://10.0.1.101:8429",
  "health_endpoint": "/-/healthy",
  "labels": {"env": "prod", "dc": "east-1", "version": "v1.93"}
}
```

#### 5.3.5 注销节点

```
DELETE /api/v1/collector-nodes/{node_id}
```

**行为：** 注销前自动将该节点上的所有采集任务迁移到其他可用节点（若无可迁移节点则任务状态标记为 `unassigned`）。

#### 5.3.6 上报心跳

```
POST /api/v1/collector-nodes/{node_id}/heartbeat
```

**请求体：**

```json
{
  "timestamp": "2026-10-27T10:30:00Z",
  "status": "healthy",
  "task_count": 120,
  "uptime_seconds": 864000,
  "metrics": {
    "scrape_duration_avg_ms": 125,
    "scrape_error_count": 0
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "node_status": "healthy",
    "pending_reassignments": []
  }
}
```

#### 5.3.7 手动触发健康检查

```
POST /api/v1/collector-nodes/{node_id}/health-check
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "node_id": "node-vmagent-east-01",
    "health_status": "healthy",
    "response_time_ms": 15,
    "checked_at": "2026-10-27T10:30:15Z",
    "details": {
      "build_info": {"version": "v1.93.0"},
      "active_targets": 120,
      "errors": []
    }
  }
}
```

#### 5.3.8 查询节点任务分配

```
GET /api/v1/collector-nodes/{node_id}/tasks
```

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| status | string | 按任务状态过滤 |
| page | int | 页码 |
| page_size | int | 每页数量 |

#### 5.3.9 触发任务再平衡

```
POST /api/v1/zones/{zone_id}/collector-nodes/rebalance
```

**行为：** 对该网区内所有 healthy 节点重新分配采集任务，使各节点负载均衡。

**响应：**

```json
{
  "code": 0,
  "data": {
    "zone_id": "zone-east-1",
    "rebalanced_at": "2026-10-27T10:35:00Z",
    "movements": [
      {"task_id": "task-010", "from_node": "node-01", "to_node": "node-03"},
      {"task_id": "task-011", "from_node": "node-01", "to_node": "node-03"}
    ],
    "node_loads_after": [
      {"node_id": "node-01", "task_count": 100},
      {"node_id": "node-02", "task_count": 105},
      {"node_id": "node-03", "task_count": 103}
    ]
  }
}
```

---

## 六、设计决策

### 6.1 健康检测双机制

**决策：** 同时使用「被动心跳」和「主动探测」两种机制。

| 机制 | 说明 | 周期 |
|------|------|------|
| 被动心跳 | 采集器节点主动上报心跳，平台更新 last_heartbeat | 节点每 30s 上报 |
| 主动探测 | 平台定时请求节点的 health_endpoint | 平台每 60s 探测 |

**理由：**
- 心跳机制可感知节点进程存活，但无法判断采集器内部功能是否正常
- health_endpoint 探测可验证采集器实际工作能力（如 vmagent 的 /-/healthy 返回版本与运行状态）
- 两者互补，避免单点判断失误

### 6.2 状态流转规则

```
                    心跳超时(3次)
unknown ──────────────────────────────> offline
   │                                       ↑
   │ 健康检查通过                            │ 连续探测失败(3次)
   ↓                                       │
healthy ←────────────────────────────> degraded
        健康检查通过(从degraded恢复)     心跳超时或探测异常
```

| 当前状态 | 触发条件 | 目标状态 | 动作 |
|----------|----------|----------|------|
| unknown | 首次注册 | - | 等待首次心跳 |
| unknown | 健康检查通过 | healthy | 可接受任务分配 |
| unknown | 心跳超时(90s) | offline | 标记不可用 |
| healthy | 心跳超时(90s) | degraded | 触发告警，暂停新任务分配 |
| healthy | 主动探测失败(连续3次) | degraded | 触发告警 |
| degraded | 心跳恢复 + 探测通过 | healthy | 解除告警，恢复任务分配 |
| degraded | 持续不可达(300s) | offline | 触发任务迁移 |
| offline | 心跳恢复 + 探测通过 | healthy | 重新纳入负载均衡 |

**超时阈值配置：**

| 参数 | 默认值 | 说明 |
|------|--------|------|
| heartbeat_timeout | 90s | 心跳超时阈值（3个心跳周期） |
| health_check_interval | 60s | 主动探测周期 |
| health_check_fail_threshold | 3 | 连续失败次数触发降级 |
| offline_timeout | 300s | degraded 持续超时转 offline |

### 6.3 任务分配策略

**决策：** 采用加权最少任务分配（Weighted Least-Task）算法。

**分配流程：**

```
新任务需要分配节点
    │
    ▼
筛选候选节点（同网区 + status=healthy）
    │
    ▼
过滤超载节点（task_count / max_tasks >= warn_threshold）
    │
    ├── 无可用节点 → 任务标记为 unassigned，触发告警
    │
    ▼
计算各节点负载比（task_count / max_tasks）
    │
    ▼
选择负载比最低的节点
    │
    ▼
更新 collection_task.node_id + collector_node.task_count++
```

**负载均衡因子计算：**

```
load_ratio = node.task_count / capacity.max_tasks

选择 min(load_ratio) 的节点分配任务
```

**再平衡触发条件：**

| 事件 | 行为 |
|------|------|
| 新节点上线（healthy） | 检查是否需要从过载节点迁移任务 |
| 节点降级/离线 | 将该节点任务迁移到其他可用节点 |
| 手动触发 | 管理员通过 API 手动触发再平衡 |

### 6.4 多采集器类型差异处理

| 采集器类型 | health_endpoint | 配置热更新 | 服务发现 | 特殊处理 |
|-----------|-----------------|-----------|---------|---------|
| vmagent | /-/healthy | POST /-/reload | http_sd, file_sd | 兼容 Prometheus 协议 |
| prometheus | /-/healthy | 写文件 + POST /-/reload | file_sd, dns_sd, consul_sd, kubernetes_sd | 需写入配置文件 |
| alloy | /-/healthy | PUT /-/config | http_sd | River 配置语言 |

**设计原则：** 采集器管理模块只负责节点生命周期与任务分配，具体的配置生成与下发由采集器适配器（collector-adapter.md）负责。两模块通过 node_type 字段协作。

### 6.5 容量管理

**默认容量配置：**

| node_type | max_tasks | 说明 |
|-----------|-----------|------|
| vmagent | 500 | 单 vmagent 实例推荐上限 |
| prometheus | 300 | Prometheus 受限于内存，建议更低 |
| alloy | 400 | Alloy 性能介于两者之间 |

**超载告警级别：**

| 负载比 | 级别 | 动作 |
|--------|------|------|
| >= 0.8 | warning | 记录日志，新任务优先分配到其他节点 |
| >= 0.95 | critical | 触发告警通知，自动触发再平衡 |
| == 1.0 | overload | 拒绝新任务分配，紧急迁移 |

### 6.6 区协调器 Gossip 协议

**当前实现：** Gossip 协议运行在区协调器（Zone Coordinator）内部，而非采集器节点之间。区协调器通过 gossip 管理区内所有执行节点（vmalert、vmagent、prometheus）的成员关系与 job 分配。

**工作机制：**

| 能力 | 说明 |
|------|------|
| 成员管理 | 执行节点启动时注册到区协调器，协调器通过 gossip 维护成员列表 |
| 故障检测 | 区协调器通过 gossip 心跳检测节点离开，触发 job 再平衡 |
| 一致性哈希 | hash(job_id) mod N → owner 节点，最小化迁移 |
| 统一协调 | 同一区协调器管理 vmalert / vmagent / prometheus 三类 job |

**架构示意：**

```
区协调器 (Zone Coordinator)
  ├── Gossip 成员管理
  │     ├── vmalert-1 ←── gossip ──→ vmalert-2
  │     ├── vmagent-1 ←── gossip ──→ vmagent-2
  │     └── prometheus-1 (单节点时退化为直接管理)
  │
  └── 一致性哈希分配
        ├── hash(rule-001) mod 3 → vmalert-2
        ├── hash(scrape-001) mod 2 → vmagent-1
        └── 节点离开 → 自动 rebalance
```

区协调器预留 `collector_node.labels` 中的扩展字段，支持后续 gossip 集群标识与分组策略。
