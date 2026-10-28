# T2 RC 任务管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：T2 | 职责：管理按网区拆分的规则检查（Rule Check）任务，负责 RC 规则的 CRUD、Prometheus rule group YAML 生成、规则包下发到各网区的 vmalert，以及 RC 任务的全生命周期状态管理。

---

## 一、概述

RC 任务管理模块是平台告警链路的执行核心。告警策略（AlertStrategy）定义了「检查什么」，而 RC 任务定义了「在哪里检查」——每个 RC 任务绑定一个网区（zone），包含一组 RC 规则（RCRule），最终生成 Prometheus 兼容的 rule group YAML 并下发到该网区的 vmalert 实例执行。

本模块的核心价值在于：

- **按网区拆分**：每个 zone 独立管理告警规则，避免跨 DC 规则混乱。
- **策略驱动**：RC 任务关联告警策略（strategy_id），规则从策略模板实例化而来。
- **规则包生成**：自动将 rc_rule 集合编译为 Prometheus rule group YAML。
- **多网区下发**：将生成的规则包推送到各网区的 vmalert 实例。
- **状态管控**：draft → active → paused → archived 四态流转，控制规则是否生效。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| RC 任务 CRUD | 创建、查询、更新、删除 RC 任务 |
| RC 规则 CRUD | 在 RC 任务下管理具体的告警规则 |
| RC Worker 管理 | 注册/注销 vmalert worker，维护查询地址与下发地址 |
| 规则包生成 | rc_rule 集合 → Prometheus rule group YAML |
| 规则下发 | 将规则包推送到目标 rc_worker 的 vmalert |
| 大区拆分 | 多 worker 场景下按规则组/查询地址/实例分片分配规则 |
| 状态管理 | draft → active → paused → archived 生命周期 |
| 策略关联 | rc_task.strategy_id 关联告警策略，从策略模板实例化规则 |
| 规则验证 | 下发前校验 PromQL 语法与变量替换完整性 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 告警策略的定义与管理 | 告警层（AlertStrategy） |
| 告警事件的接收与处理 | 告警层（AlertEvent） |
| 告警通知的发送 | 通知层（NotificationChannel） |
| 指标定义与 PromQL 模板 | T3 指标管理模块 |
| vmalert 实例的部署与运维 | 基础设施层 |
| 采集任务的执行 | T1 采集任务模块 |

---

## 三、功能清单

| 编号 | 功能 | 优先级 | 说明 |
|------|------|--------|------|
| F2.1 | RC 任务创建 | P0 | 指定 zone_id + strategy_id，创建 RC 任务 |
| F2.2 | RC 任务查询 | P0 | 按 zone / strategy / status 维度查询 RC 任务列表 |
| F2.3 | RC 任务更新 | P0 | 更新名称、描述、rule_group 等元信息 |
| F2.4 | RC 任务删除 | P0 | 删除前需先归档或确保无 active 状态 |
| F2.5 | RC 规则添加 | P0 | 在 RC 任务下添加告警规则（alert_name + expr + duration + severity） |
| F2.6 | RC 规则批量导入 | P1 | 从告警策略模板批量实例化规则到 RC 任务 |
| F2.7 | RC 规则更新 | P0 | 更新规则的 expr / duration / severity / labels / annotations |
| F2.8 | RC 规则删除 | P0 | 删除单条规则 |
| F2.9 | 规则包生成 | P0 | 将 RC 任务下所有规则编译为 Prometheus rule group YAML |
| F2.10 | 规则包预览 | P1 | 生成前预览 YAML 内容，支持手动调整 |
| F2.11 | 规则包下发 | P0 | 将生成的 YAML 推送到目标网区的 vmalert 实例 |
| F2.12 | 规则包下发确认 | P1 | 下发后查询 vmalert 确认规则加载成功 |
| F2.13 | 状态流转 | P0 | draft → active → paused → archived，含各状态转换的前置校验 |
| F2.14 | 批量激活/暂停 | P1 | 按 zone 批量激活或暂停 RC 任务 |
| F2.15 | PromQL 校验 | P1 | 规则保存前校验 PromQL 语法合法性 |
| F2.16 | 变量替换校验 | P0 | 确认 expr 中所有 $variable 占位符已被替换 |
| F2.17 | 规则版本历史 | P2 | 记录规则的每次变更，支持回滚 |
| F2.18 | 下发失败重试 | P1 | 下发失败后自动重试，重试失败记录错误并告警 |

---

## 四、核心数据模型

### 4.1 rc_task（RC 任务）

```sql
CREATE TABLE rc_task (
    task_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '所属网区',
    name            VARCHAR(256) NOT NULL,
    description     TEXT,
    strategy_id     VARCHAR(64) COMMENT '来源告警策略ID',
    enabled         BOOLEAN DEFAULT TRUE,
    status          VARCHAR(32) DEFAULT 'draft' COMMENT 'draft/active/paused/archived',
    rule_group      VARCHAR(128) COMMENT 'Prometheus rule group 名称',
    deploy_target   VARCHAR(512) COMMENT 'vmalert 下发地址',
    last_deploy_at  TIMESTAMP NULL COMMENT '最近一次下发时间',
    last_deploy_status VARCHAR(32) COMMENT 'recent_deploy_status: success/failed',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_rct_zone ON rc_task(zone_id);
CREATE INDEX idx_rct_strategy ON rc_task(strategy_id);
CREATE INDEX idx_rct_status ON rc_task(status);
```

### 4.2 rc_rule（RC 规则）

```sql
CREATE TABLE rc_rule (
    rule_id         VARCHAR(64) PRIMARY KEY,
    task_id         VARCHAR(64) NOT NULL COMMENT '所属RC任务',
    alert_name      VARCHAR(256) NOT NULL COMMENT '告警规则名称',
    expr            TEXT NOT NULL COMMENT 'PromQL表达式(变量已替换)',
    duration        VARCHAR(32) NOT NULL COMMENT '持续时间: 5m, 10s, 1h',
    severity        VARCHAR(32) NOT NULL COMMENT 'critical/warning/info',
    labels          JSON COMMENT '附加标签 {"team": "infra"}',
    annotations     JSON COMMENT '注释模板 {"summary": "...", "description": "..."}',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES rc_task(task_id) ON DELETE CASCADE
);

CREATE INDEX idx_rule_task ON rc_rule(task_id);
CREATE INDEX idx_rule_severity ON rc_rule(severity);
```

### 4.3 rc_rule_history（规则变更历史，扩展）

```sql
CREATE TABLE rc_rule_history (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    rule_id         VARCHAR(64) NOT NULL,
    action          VARCHAR(16) NOT NULL COMMENT 'create/update/delete',
    old_value       JSON COMMENT '变更前快照',
    new_value       JSON COMMENT '变更后快照',
    operator        VARCHAR(64) COMMENT '操作人',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (rule_id) REFERENCES rc_rule(rule_id) ON DELETE CASCADE
);

CREATE INDEX idx_rule_hist_rule ON rc_rule_history(rule_id);
CREATE INDEX idx_rule_hist_time ON rc_rule_history(created_at);
```

### 4.4 状态枚举定义

| 状态 | 含义 | 允许转换 | 规则下发 |
|------|------|----------|----------|
| `draft` | 草稿，规则编辑中 | → active, archived | 不下发 |
| `active` | 已激活，规则生效中 | → paused, archived, draft | 下发到 vmalert worker |
| `paused` | 已暂停，规则暂时不生效 | → active, archived | 从 vmalert 移除 |
| `archived` | 已归档，不可恢复 | 终态 | 从 vmalert 移除 |

### 4.5 RC Worker（vmalert 工作节点 — 区协调器分配）

> 大区场景下规则量大（1000+），单个 vmalert 无法承载全量评估。区协调器通过 gossip 协议管理同区多个 vmalert 节点，一致性哈希分配规则归属，每个 vmalert 只评估自己分到的规则子集，不跑全量。

```sql
CREATE TABLE rc_worker (
    worker_id       VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '所属网区',
    name            VARCHAR(128) NOT NULL COMMENT '显示名称: vmalert-east-1',
    address         VARCHAR(256) NOT NULL COMMENT 'vmalert HTTP地址(规则拉取/热加载)',
    query_address   VARCHAR(256) NOT NULL COMMENT '查询地址(规则评估用)',
    capacity        INT DEFAULT 500 COMMENT '该实例可承载的最大规则数',
    owned_rules     INT DEFAULT 0 COMMENT '当前实际拥有的规则数(协调器分配后)',
    status          VARCHAR(32) DEFAULT 'unknown' COMMENT 'healthy/degraded/offline',
    last_heartbeat  TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_rcw_zone ON rc_worker(zone_id);
CREATE INDEX idx_rcw_status ON rc_worker(status);
```

**区协调器的 Job 分配机制：**

```
区协调器 (每区一个, 运行 gossip 协议)
│
├── 管理本区所有节点:
│     ├── vmalert × N (规则评估节点)
│     ├── vmagent × N (采集节点)
│     └── prometheus × N (采集节点)
│
├── Gossip 功能:
│     ├── 成员发现: 节点启动时注册到协调器
│     ├── 成员列表收敛: gossip 传播，所有节点最终一致
│     ├── 一致性哈希: hash(job_id) mod N → owner 节点
│     └── 故障接管: 节点离开 → 剩余节点接管其 job
│
├── 规则分配 (vmalert):
│     ├── T2 将全量规则保存到管理平台数据库
│     ├── 区协调器监听规则变更, 按 hash(rule_id) 分配给 vmalert 节点
│     └── 区协调器推送规则到对应 vmalert 的 HTTP API (热加载)
│
└── 采集任务分配 (vmagent/prometheus):
      ├── T1 将全量采集任务发布
      ├── 区协调器按 hash(target_id) 分配给 vmagent/prometheus
      └── 每个采集节点只负责自己 owned 的 target
```

**与静态分配的区别：**

| 维度 | 静态分配(已废弃) | 区协调器 gossip(当前方案) |
|------|-----------------|-------------------------|
| 规则归属 | T2 手动指定 worker_id | 区协调器一致性哈希自动决定 |
| 扩缩容 | 手动调整分配表 | 新节点注册后自动 rebalance |
| 故障恢复 | 手动迁移规则 | 协调器检测离开，剩余节点接管 |
| T2 职责 | 规则生成 + 分配 + 下发 | 只管规则生成, 区协调器负责分配与推送 |
| 统一协调 | 各类型独立管理 | 区协调器统一管理 vmalert/vmagent/prometheus |

**rc_task 表扩展字段：**

```sql
ALTER TABLE rc_task ADD COLUMN query_address VARCHAR(256) COMMENT '规则评估查询地址(覆盖集群默认)';
```

> `query_address` 允许单条 RC 任务覆盖集群默认的查询地址——比如某些跨区规则需要查 global vmselect 而非本地 vmselect。规则归属哪个 vmalert 实例由区协调器自动决定，不需要 worker_id 字段。

**单 worker 区（默认）：** 小区只注册 1 个 rc_worker，区协调器退化为单节点管理，该节点跑全量规则，无需 gossip 通信开销。

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| 资源管理（Zone） | 获取 zone 信息、vmalert 下发地址 | 内部服务调用 |
| 告警层（AlertStrategy） | 获取告警策略模板，实例化规则 | 内部服务调用 |
| T3 指标管理（MetricDefinition） | 获取 PromQL 模板与变量定义 | 内部服务调用 |
| 基础设施（CollectorNode） | 获取 zone 下 vmalert 实例地址 | 内部服务调用 |

### 5.2 下游提供

| 消费方 | 提供内容 | 交互方式 |
|--------|----------|----------|
| 区协调器 (I1) | 规则包推送到 vmalert 节点 | 区协调器推送规则到 vmalert HTTP API |
| rc_worker (vmalert) | 规则元数据 + 查询地址 | vmalert 通过 gossip 自组织，pull 模式获取规则 |
| 区AM | 去重后的告警转发到核心AM | vmalert → 区AM(fingerprint去重) → 核心AM(策略路由) |
| 核心AM | 分组/抑制/静默后输出到 Kafka | 核心AM → Kafka → 消费者 |
| 前端 UI | RC 任务与规则的查询展示 | REST API |

### 5.3 对外 API

#### 5.3.1 RC 任务管理

**创建 RC 任务**

```
POST /api/v1/zones/{zone_id}/rc-tasks
```

请求体：

```json
{
  "name": "华东区 MySQL 告警检查",
  "description": "覆盖华东区所有 MySQL 实例的告警规则",
  "strategy_id": "strategy-mysql-ha",
  "rule_group": "zone-east-mysql-alerts",
  "deploy_target": "http://vmalert-east:8880"
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "task_id": "rct-001",
    "zone_id": "zone-east-1",
    "status": "draft",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**查询 RC 任务列表**

```
GET /api/v1/zones/{zone_id}/rc-tasks?status=active&strategy_id=strategy-mysql-ha&page=1&page_size=20
```

**查询单个 RC 任务**

```
GET /api/v1/rc-tasks/{task_id}
```

**更新 RC 任务**

```
PUT /api/v1/rc-tasks/{task_id}
```

**删除 RC 任务**

```
DELETE /api/v1/rc-tasks/{task_id}
```

> 前置校验：status 必须为 archived 或 draft（无 active 规则）。

**激活 RC 任务**

```
POST /api/v1/rc-tasks/{task_id}/activate
```

> 触发规则包生成并下发到 vmalert。

**暂停 RC 任务**

```
POST /api/v1/rc-tasks/{task_id}/pause
```

> 从 vmalert 移除该 rule group。

**归档 RC 任务**

```
POST /api/v1/rc-tasks/{task_id}/archive
```

> 终态操作，不可恢复。

#### 5.3.2 RC 规则管理

**添加规则**

```
POST /api/v1/rc-tasks/{task_id}/rules
```

请求体：

```json
{
  "alert_name": "MySQLDown",
  "expr": "mysql_up{instance=\"$instance\"} == 0",
  "duration": "5m",
  "severity": "critical",
  "labels": {
    "team": "dba",
    "dc": "zone-east-1"
  },
  "annotations": {
    "summary": "MySQL 实例 {{ $labels.instance }} 不可达",
    "description": "实例 {{ $labels.instance }} 已持续 {{ $value }} 秒无法连接，超过 5 分钟阈值。"
  }
}
```

**从策略批量导入规则**

```
POST /api/v1/rc-tasks/{task_id}/rules/import
```

请求体：

```json
{
  "strategy_id": "strategy-mysql-ha",
  "scope": {
    "type": "zone",
    "value": "zone-east-1"
  },
  "variable_overrides": {
    "instance": "inst-mysql-*"
  }
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "imported_count": 8,
    "rules": [
      { "rule_id": "rule-001", "alert_name": "MySQLDown", "severity": "critical" },
      { "rule_id": "rule-002", "alert_name": "MySQLHighConnections", "severity": "warning" }
    ]
  }
}
```

**查询规则列表**

```
GET /api/v1/rc-tasks/{task_id}/rules?page=1&page_size=50
```

**查询单个规则**

```
GET /api/v1/rc-rules/{rule_id}
```

**更新规则**

```
PUT /api/v1/rc-rules/{rule_id}
```

**删除规则**

```
DELETE /api/v1/rc-rules/{rule_id}
```

**查询规则变更历史**

```
GET /api/v1/rc-rules/{rule_id}/history?page=1&page_size=20
```

#### 5.3.3 规则包操作

**预览规则包**

```
POST /api/v1/rc-tasks/{task_id}/rule-group/preview
```

响应（YAML 格式）：

```json
{
  "code": 0,
  "data": {
    "yaml": "groups:\n  - name: zone-east-mysql-alerts\n    rules:\n      - alert: MySQLDown\n        expr: mysql_up{instance=\"inst-mysql-001\"} == 0\n        for: 5m\n        labels:\n          severity: critical\n          team: dba\n        annotations:\n          summary: \"MySQL 实例不可达\"\n",
    "rule_count": 8,
    "validation": {
      "valid": true,
      "errors": []
    }
  }
}
```

**生成并下发规则包**

```
POST /api/v1/rc-tasks/{task_id}/deploy
```

响应：

```json
{
  "code": 0,
  "data": {
    "task_id": "rct-001",
    "deploy_target": "http://vmalert-east:8880",
    "rule_count": 8,
    "deploy_status": "success",
    "deployed_at": "2026-10-27T10:15:00Z"
  }
}
```

**查询下发状态**

```
GET /api/v1/rc-tasks/{task_id}/deploy-status
```

**批量操作**

```
POST /api/v1/rc-tasks/batch
```

请求体：

```json
{
  "action": "activate",
  "task_ids": ["rct-001", "rct-002"],
  "zone_id": "zone-east-1"
}
```

#### 5.3.3 RC Worker 管理

**注册 Worker**

```
POST /api/v1/zones/{zone_id}/rc-workers
```

请求体：

```json
{
  "name": "vmalert-east-1",
  "address": "http://vmalert-east-1:8880",
  "query_address": "http://vmselect-east-1:8481",
  "gossip_addr": "10.0.1.5:7946",
  "capacity": 800
}
```

**查询 Worker 列表**

```
GET /api/v1/zones/{zone_id}/rc-workers
```

**查询单个 Worker**

```
GET /api/v1/rc-workers/{worker_id}
```

**更新 Worker**

```
PUT /api/v1/rc-workers/{worker_id}
```

**注销 Worker**

```
DELETE /api/v1/rc-workers/{worker_id}
```

> 注销前需确保该 worker 上无 active 状态的 RC 任务，否则拒绝删除。

**查询 Worker 负载**

```
GET /api/v1/rc-workers/{worker_id}/load
```

响应：

```json
{
  "code": 0,
  "data": {
    "worker_id": "rcw-east-1",
    "owned_rules": 342,
    "capacity": 800,
    "utilization": 0.43,
    "gossip_members": 4,
    "status": "healthy"
  }
}
```

---

## 六、设计决策

### 6.1 按网区拆分 RC 任务

**决策**：每个 zone 对应一个或多个独立的 rc_task，规则按 zone 隔离。

**理由**：
- 不同网区（DC）的 vmalert 实例独立运行，规则需分别下发。
- 同一告警策略在不同 zone 可能有不同的变量值（如实例列表不同）。
- 按 zone 管理便于权限隔离：zone 管理员只管理本区规则。
- 故障域隔离：一个 zone 的 vmalert 故障不影响其他 zone。

### 6.2 策略驱动的规则实例化

**决策**：RC 任务通过 strategy_id 关联告警策略，规则从策略模板批量实例化。

**流程**：

```
AlertStrategy (全局模板)
    │
    ├── expr_template: "mysql_up{instance=\"$instance\"} == 0"
    ├── scope_type: "instance"
    └── scope_value: ["inst-001", "inst-002", ...]
    │
    ▼  实例化
rc_task (zone-east-1)
    │
    ├── rc_rule: expr="mysql_up{instance=\"inst-001\"} == 0"
    ├── rc_rule: expr="mysql_up{instance=\"inst-002\"} == 0"
    └── rc_rule: expr="mysql_up{instance=\"inst-003\"} == 0"
```

**理由**：
- 策略定义「检查什么」（模板），RC 任务定义「在哪里检查」（实例化）。
- 策略更新后，可选择性地重新实例化受影响 zone 的规则。
- 支持手动添加策略之外的自定义规则。

### 6.3 规则包生成格式

**决策**：生成标准 Prometheus rule group YAML 格式，兼容 vmalert / Prometheus / Thanos Ruler。

**生成的 YAML 结构**：

```yaml
groups:
  - name: zone-east-mysql-alerts
    interval: 1m
    rules:
      - alert: MySQLDown
        expr: mysql_up{instance="inst-mysql-001"} == 0
        for: 5m
        labels:
          severity: critical
          team: dba
          zone: zone-east-1
        annotations:
          summary: "MySQL 实例 inst-mysql-001 不可达"
          description: "实例已持续 {{ $value }} 秒无法连接"
      - alert: MySQLHighConnections
        expr: mysql_global_status_threads_connected > 100
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "连接数过高"
```

**理由**：
- Prometheus rule group 是业界标准格式，生态兼容性最好。
- vmalert 完全兼容 Prometheus rules 语法。
- YAML 格式便于人工审查和版本管理。

### 6.4 下发机制

**决策**：通过 HTTP API 将规则包推送到 vmalert，支持失败重试与状态确认。

**下发流程**：

```
rc_task.status = active
    │
    ▼
生成 rule group YAML
    │
    ▼
POST {deploy_target}/api/v1/rules/{rule_group}
    │
    ├── 200 → 下发成功，更新 last_deploy_status = success
    │
    └── 失败 → 重试(最多3次, 间隔指数退避)
                │
                ├── 成功 → 更新状态
                └── 仍失败 → last_deploy_status = failed, 触发告警通知
```

**理由**：
- HTTP 推送是 vmalert 支持的标准下发方式。
- 失败重试避免网络抖动导致的下发失败。
- 下发状态记录便于运维排查。

### 6.5 状态机设计

**决策**：RC 任务采用 draft → active → paused → archived 四态模型。

```
         activate
  draft ──────────→ active
    │                  │  │
    │    archive       │  │ pause
    │  ┌───────────────┘  │
    │  │                  ▼
    │  │              paused
    │  │                │
    │  │    activate    │
    │  │  ◄─────────────┘
    │  │
    ▼  ▼
  archived (终态)
```

**理由**：
- draft 状态允许安全编辑规则，不影响线上 vmalert。
- active 状态表示规则已下发并生效。
- paused 状态支持临时停用（如维护窗口），规则从 vmalert 移除但保留配置。
- archived 为终态，防止误操作恢复已废弃的规则集。

### 6.6 规则版本历史

**决策**：记录每条规则的变更快照（old_value / new_value），支持审计与回滚。

**理由**：
- 告警规则是生产安全的关键配置，变更需可追溯。
- 出错时可快速回滚到上一个正确版本。
- 满足合规审计要求。

### 6.7 区协调器统一管理 Job 分配

**决策**：每个网区部署一个区协调器进程，运行 gossip 协议，统一协调本区内 vmalert、vmagent、prometheus 的 job 分配。大区多节点时一致性哈希打散，每个节点只跑自己 owned 的 job 子集。

**为什么用区协调器而不是各组件自治：**

| 方案 | 优点 | 缺点 |
|------|------|------|
| 各组件独立 gossip | 解耦 | 多个 gossip 通道，运维复杂 |
| **区协调器统一管理** | 单一 gossip 通道，统一视图 | 协调器本身需高可用 |
| 中心调度分配 | 简单 | 中心故障则全区停摆 |

**区协调器职责：**

```
区协调器 (每区一个)
│
├── Gossip 成员管理
│     ├── vmalert 节点注册/发现/心跳
│     ├── vmagent 节点注册/发现/心跳
│     └── prometheus 节点注册/发现/心跳
│
├── Job 分配 (一致性哈希)
│     ├── RC 规则 → vmalert 节点
│     ├── 采集任务 → vmagent/prometheus 节点
│     └── 成员变更时自动 rebalance
│
├── 故障接管
│     ├── 节点超时 → 标记 offline
│     └── 其 owned job → 分配给剩余节点
│
└── 与上游交互
      ├── T2 发布规则 → 协调器分配给 vmalert
      ├── T1 发布采集任务 → 协调器分配给 vmagent/prometheus
      └── I1 注册节点 → 协调器维护成员列表
```

**架构示意：**

```
大区 Zone-A (2000 规则, 4 个 vmalert)
│
├── 区协调器-A (gossip seed, 统一管理)
│     ├── 管理 vmalert-A1~A4 成员列表
│     ├── 管理 vmagent-A1~A3 成员列表
│     └── hash 分配规则/任务
│
├── vmalert-A1 owns ~500 条规则
├── vmalert-A2 owns ~500 条规则
├── vmalert-A3 owns ~500 条规则
├── vmalert-A4 owns ~500 条规则
│
└── 区AM-A (fingerprint去重, 单实例)
      └──→ 核心AM (策略路由)
```

**单 worker 区（默认）：** 小区只注册一个 vmalert/vmagent，区协调器退化为单节点管理，无需 gossip 通信开销。

**理由：**
- 统一 gossip 通道减少运维复杂度
- 区协调器故障时，已有分配继续运行（节点本地缓存分配表）
- 一致性哈希保证 rebalance 时最小迁移
- 区协调器本身可通过主备或 raft 实现高可用
