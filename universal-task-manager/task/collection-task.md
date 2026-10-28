# T1 采集任务管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：T1 | 职责：管理采集任务的全生命周期，包括动态任务类型注册、表单驱动的任务配置、采集器原生配置生成、任务状态流转与节点分配。

---

## 一、概述

采集任务管理模块是平台的核心执行单元。它基于「任务类型（TaskType）」的声明式注册机制，将用户配置转化为各采集器（vmagent / Prometheus / Alloy）可执行的原生配置，并驱动采集任务从创建到运行的完整生命周期。

本模块的核心价值在于：

- **声明式扩展**：新增采集目标无需编码，只需注册 JSON 格式的 TaskType 即可。
- **配置解耦**：用户填写表单配置，平台自动转换为采集器原生格式，屏蔽底层差异。
- **多存储路由**：任务可指定数据输出目标（data_sink_id），实现采集数据分流。
- **节点感知**：任务由区协调器(I1)通过 gossip 一致性哈希分配到采集器节点，支持分布式部署。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| TaskType 注册与管理 | 动态注册、更新、版本化任务类型定义 |
| 表单渲染数据提供 | 输出 form_schema 供前端动态生成配置表单 |
| 采集任务 CRUD | 创建、查询、更新、删除采集任务 |
| 配置生成 | task.config + instance 信息 + config_template → 采集器原生配置 |
| 任务生命周期管理 | pending → running → error → stopped 状态流转 |
| 多存储路由 | 任务级 data_sink_id 指定数据输出目标 |
| 任务启用/禁用控制 | 运行时启停任务，不删除配置 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 采集器节点的注册与健康检查 | 基础设施层（I1 采集器与协调器管理） |
| 节点任务分配与负载均衡 | 区协调器（I1）通过 gossip 一致性哈希自动分配 |
| 实际的数据采集执行 | 采集器节点自身 |
| 数据的存储与查询 | 数据存储层（DataSink: VictoriaMetrics / Elasticsearch / Loki 等） |
| 服务发现与目标自动发现 | T4 服务发现模块 |
| 告警规则生成与下发 | T2 RC 任务模块 |
| 指标定义与管理 | T3 指标管理模块 |

---

## 三、功能清单

| 编号 | 功能 | 优先级 | 说明 |
|------|------|--------|------|
| F1.1 | 任务类型注册 | P0 | 通过 JSON 声明式注册新的 TaskType，含 form_schema / config_template / declared_metrics |
| F1.2 | 任务类型版本管理 | P1 | 更新 TaskType 时自动递增 version，已创建的任务保持旧版本快照 |
| F1.3 | 任务类型批量刷新 | P2 | 将引用某 TaskType 的所有任务批量升级到最新版本配置 |
| F1.4 | 表单 Schema 查询 | P0 | 返回 form_schema JSON 供前端动态渲染配置表单 |
| F1.5 | 采集任务创建 | P0 | 绑定 instance + type_id，填写表单配置，创建采集任务 |
| F1.6 | 采集任务配置更新 | P0 | 更新 task.config，触发配置重新生成与热加载 |
| F1.7 | 采集任务删除 | P0 | 删除任务并清理采集器上的对应配置 |
| F1.8 | 配置生成引擎 | P0 | task.config + instance 信息 + config_template.{collector} → 采集器原生配置 |
| F1.9 | 配置预览 | P1 | 在创建/更新前预览生成的采集器配置（支持 vmagent / alloy / prometheus） |
| F1.10 | 任务状态流转 | P0 | pending → running → error → stopped，含自动重试与错误记录 |
| F1.11 | 任务启用/禁用 | P0 | 运行时启停任务，禁用后从采集器移除对应 scrape config |
| F1.12 | 多存储路由 | P1 | 任务可指定 data_sink_id，配置生成时注入对应的 remote_write / remote_read 配置 |
| F1.13 | 配置热加载 | P0 | 配置变更后自动触发采集器 reload（vmagent POST /-/reload 等） |
| F1.14 | 批量操作 | P1 | 批量启用 / 禁用 / 删除任务 |
| F1.15 | 任务状态统计 | P1 | 按 zone / type 维度统计任务数量与状态分布 |
| F1.16 | 任务执行日志 | P2 | 记录 last_scrape_at / last_error，支持查看近期执行历史 |

---

## 四、核心数据模型

### 4.1 task_type（任务类型）

```sql
CREATE TABLE task_type (
    type_id         VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    source_type     VARCHAR(32) NOT NULL COMMENT '适用资源类型: mysql/redis/linux/snmp...',
    collector       VARCHAR(32) NOT NULL COMMENT '默认采集器: vmagent/prometheus/alloy/custom',
    form_schema     JSON NOT NULL COMMENT '表单Schema，驱动前端动态渲染',
    detail_panels   JSON COMMENT '详情页面板布局定义',
    declared_metrics JSON COMMENT '声明的指标列表 [{name, type, help, labels}]',
    config_template JSON NOT NULL COMMENT '各采集器配置模板 {vmagent: {...}, alloy: {...}}',
    is_builtin      BOOLEAN DEFAULT FALSE COMMENT '是否系统内置',
    version         INT DEFAULT 1 COMMENT '版本号，每次更新递增',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**form_schema 结构规范：**

```json
{
  "fields": [
    {
      "key": "scrape_interval",
      "label": "采集间隔",
      "type": "select",
      "options": ["10s", "15s", "30s", "60s"],
      "default": "15s",
      "required": true,
      "group": "基础配置",
      "help": "两次采集之间的时间间隔"
    },
    {
      "key": "databases",
      "label": "监控数据库",
      "type": "tags",
      "default": ["0"],
      "group": "采集项"
    }
  ]
}
```

**config_template 结构规范：**

```json
{
  "vmagent": {
    "job_name": "{{type_id}}-{{instance_id}}",
    "scrape_interval": "{{config.scrape_interval}}",
    "scrape_timeout": "10s",
    "metrics_path": "/metrics",
    "static_configs": [{
      "targets": ["{{instance.host}}:{{instance.port}}"],
      "labels": {
        "instance": "{{instance_id}}",
        "zone": "{{instance.zone_id}}"
      }
    }]
  },
  "alloy": {
    "prometheus.scrape": {
      "job_name": "{{type_id}}-{{instance_id}}",
      "scrape_interval": "{{config.scrape_interval}}",
      "targets": [{
        "__address__": "{{instance.host}}:{{instance.port}}"
      }],
      "forward_to": ["prometheus.remote_write.default.receiver"]
    }
  }
}
```

### 4.2 collection_task（采集任务）

```sql
CREATE TABLE collection_task (
    task_id         VARCHAR(64) PRIMARY KEY,
    instance_id     VARCHAR(64) NOT NULL COMMENT '绑定的实例',
    type_id         VARCHAR(64) NOT NULL COMMENT '任务类型',
    node_id         VARCHAR(64) COMMENT '区协调器分配的采集器节点(只读)',
    name            VARCHAR(256) NOT NULL,
    config          JSON NOT NULL COMMENT '用户填写的表单配置值',
    data_sink_id    VARCHAR(64) COMMENT '数据存储目标，NULL使用zone默认',
    enabled         BOOLEAN DEFAULT TRUE,
    status          VARCHAR(32) DEFAULT 'pending' COMMENT 'pending/running/error/stopped',
    last_scrape_at  TIMESTAMP NULL COMMENT '最近一次成功采集时间',
    last_error      TEXT COMMENT '最近一次错误信息',
    error_count     INT DEFAULT 0 COMMENT '累计错误次数',
    type_version    INT COMMENT '创建时的TaskType版本快照',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (instance_id) REFERENCES instance(instance_id) ON DELETE CASCADE,
    FOREIGN KEY (type_id) REFERENCES task_type(type_id),
    FOREIGN KEY (node_id) REFERENCES collector_node(node_id) ON DELETE SET NULL  -- 协调器分配后回写
);

CREATE INDEX idx_ct_instance ON collection_task(instance_id);
CREATE INDEX idx_ct_type ON collection_task(type_id);
CREATE INDEX idx_ct_node ON collection_task(node_id);
CREATE INDEX idx_ct_status ON collection_task(status);
CREATE INDEX idx_ct_enabled ON collection_task(enabled);
CREATE INDEX idx_ct_data_sink ON collection_task(data_sink_id);
```

### 4.3 data_sink（数据存储目标，扩展）

```sql
CREATE TABLE data_sink (
    sink_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    sink_type       VARCHAR(32) NOT NULL COMMENT 'victoriametrics/elasticsearch/loki/openobserve/s3',
    config          JSON NOT NULL COMMENT '连接配置 {url, auth, ...}',
    enabled         BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);
```

### 4.4 状态枚举定义

| 状态 | 含义 | 允许转换 |
|------|------|----------|
| `pending` | 已创建，等待首次配置下发 | → running, stopped |
| `running` | 正常运行中 | → error, stopped |
| `error` | 采集异常（含 last_error 详情） | → running, stopped |
| `stopped` | 手动停止 | → pending, running |

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| 资源管理（Instance） | 获取实例的 host / port / credential / zone_id | 内部服务调用 |
| 资源管理（ResourceType） | 校验 source_type 与 task_type.source_type 匹配 | 内部服务调用 |
| 基础设施（CollectorNode） | 获取可用节点列表、节点负载信息 | 内部服务调用 |
| 基础设施（DataSink） | 获取存储目标的连接配置 | 内部服务调用 |
| 凭据管理（Credential） | 获取认证信息注入采集配置 | 内部服务调用 |

### 5.2 下游提供

| 消费方 | 提供内容 | 交互方式 |
|--------|----------|----------|
| 采集器适配层 | 采集器原生配置字节流 | 内部服务调用 |
| T4 服务发现 | 已启用任务的 target 列表（http_sd 端点） | HTTP API |
| T3 指标管理 | task_type.declared_metrics 作为指标来源之一 | 内部服务调用 |
| T5 定时任务 | 采集任务状态作为健康检查的输入 | 内部服务调用 |
| 前端 UI | form_schema / detail_panels 驱动表单渲染 | REST API |

### 5.3 对外 API

#### 5.3.1 任务类型管理

**注册任务类型**

```
POST /api/v1/task-types
```

请求体：

```json
{
  "type_id": "mysql-standard",
  "name": "MySQL 标准采集",
  "description": "采集MySQL运行状态、连接数、慢查询等指标",
  "source_type": "mysql",
  "collector": "vmagent",
  "form_schema": { "fields": [...] },
  "detail_panels": [...],
  "declared_metrics": [
    { "name": "mysql_up", "type": "gauge", "help": "MySQL实例是否可达" }
  ],
  "config_template": {
    "vmagent": { "job_name": "mysql-{{instance_id}}", "..." : "..." },
    "alloy": { "..." : "..." }
  }
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "type_id": "mysql-standard",
    "version": 1,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**查询任务类型列表**

```
GET /api/v1/task-types?source_type=mysql&collector=vmagent&is_builtin=true&page=1&page_size=20
```

**查询单个任务类型**

```
GET /api/v1/task-types/{type_id}
```

**更新任务类型**

```
PUT /api/v1/task-types/{type_id}
```

> 更新时自动递增 version。已创建的采集任务保持旧版本快照，需调用批量刷新接口升级。

**删除任务类型**

```
DELETE /api/v1/task-types/{type_id}
```

> 前置校验：无采集任务引用该类型时方可删除。

**获取表单 Schema**

```
GET /api/v1/task-types/{type_id}/form
```

响应：

```json
{
  "code": 0,
  "data": {
    "type_id": "mysql-standard",
    "form_schema": {
      "fields": [
        { "key": "scrape_interval", "label": "采集间隔", "type": "select", "..." : "..." }
      ]
    }
  }
}
```

**获取声明的指标列表**

```
GET /api/v1/task-types/{type_id}/metrics
```

**批量刷新任务版本**

```
POST /api/v1/task-types/{type_id}/refresh
```

请求体：

```json
{
  "dry_run": false,
  "apply_config": true
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "affected_tasks": 15,
    "old_version": 2,
    "new_version": 3,
    "dry_run": false
  }
}
```

#### 5.3.2 采集任务管理

**创建采集任务**

```
POST /api/v1/instances/{instance_id}/tasks
```

请求体：

```json
{
  "type_id": "mysql-standard",
  "name": "MySQL 生产采集",
  "config": {
    "scrape_interval": "15s",
    "enable_slow_query": false,
    "databases": ["mysql", "performance_schema"]
  },
  "data_sink_id": "sink-vm-001"
}
```

> node_id 由区协调器(I1)自动分配，创建任务时无需指定。data_sink_id 可选，不指定时使用 zone 默认存储。

**查询实例的任务列表**

```
GET /api/v1/instances/{instance_id}/tasks?status=running&enabled=true&page=1&page_size=20
```

**查询单个采集任务**

```
GET /api/v1/collection-tasks/{task_id}
```

响应：

```json
{
  "code": 0,
  "data": {
    "task_id": "ct-001",
    "instance_id": "inst-mysql-001",
    "type_id": "mysql-standard",
    "node_id": "node-east-1",
    "name": "MySQL 生产采集",
    "config": { "..." : "..." },
    "data_sink_id": "sink-vm-001",
    "enabled": true,
    "status": "running",
    "type_version": 3,
    "last_scrape_at": "2026-10-27T10:30:15Z",
    "last_error": null,
    "error_count": 0,
    "created_at": "2026-10-27T09:00:00Z"
  }
}
```

**更新采集任务**

```
PUT /api/v1/collection-tasks/{task_id}
```

> 更新 config 后自动触发配置重新生成与热加载。

**删除采集任务**

```
DELETE /api/v1/collection-tasks/{task_id}
```

**启用任务**

```
POST /api/v1/collection-tasks/{task_id}/enable
```

**禁用任务**

```
POST /api/v1/collection-tasks/{task_id}/disable
```

**获取采集器配置（预览）**

```
GET /api/v1/collection-tasks/{task_id}/config/{collector_type}
```

> collector_type: vmagent | alloy | prometheus | custom

**查询任务执行状态**

```
GET /api/v1/collection-tasks/{task_id}/status
```

响应：

```json
{
  "code": 0,
  "data": {
    "task_id": "ct-001",
    "status": "running",
    "collector": "vmagent",
    "zone_id": "zone-east-1",
    "node_id": "node-east-1",
    "target": "10.0.1.5:9104",
    "scrape_count_24h": 5760,
    "error_count_24h": 0,
    "last_success_at": "2026-10-27T10:30:15Z",
    "avg_scrape_duration_ms": 125
  }
}
```

**批量操作**

```
POST /api/v1/collection-tasks/batch
```

请求体：

```json
{
  "action": "enable",
  "task_ids": ["ct-001", "ct-002", "ct-003"]
}
```

> action 可选：enable / disable / delete / rebalance

**触发节点重新分配（由区协调器处理）**

```
POST /api/v1/collection-tasks/{task_id}/rebalance
```

> 该接口通知区协调器(I1)对指定任务或整个 zone 的任务重新执行一致性哈希分配，不支持手动指定 node_id。

**按维度统计**

```
GET /api/v1/collection-tasks/statistics?group_by=node&type_id=mysql-standard
```

---

## 六、设计决策

### 6.1 声明式 TaskType 注册

**决策**：TaskType 以 JSON 文档形式注册，包含 form_schema / config_template / declared_metrics，而非硬编码。

**理由**：
- 新增监控目标（如 Redis → Kafka → 自定义设备）无需修改平台代码。
- form_schema 驱动前端表单渲染，实现「配置即代码」。
- config_template 将采集器配置生成逻辑从代码中解耦，以模板变量方式声明。
- 社区贡献者只需提交 JSON 定义即可扩展平台能力。

### 6.2 任务版本快照

**决策**：创建采集任务时记录 type_version（当时的 TaskType 版本号），TaskType 更新不影响已有任务。

**理由**：
- 避免批量任务配置被意外覆盖，保障生产稳定性。
- 用户可主动选择是否批量刷新到新版本。
- 支持版本回滚：保留历史 config_template 快照。

### 6.3 配置生成流水线

**决策**：采用 task.config + instance 信息 + config_template.{collector} 三方合并的方式生成采集器配置。

**流程**：

```
用户填写表单 → task.config (JSON)
                    │
                    ▼
         ┌─── 模板渲染引擎 ───┐
         │                     │
instance 信息 ──────────→  变量替换  ──→ 采集器原生配置
         │                     │
config_template ─────────┘
```

**理由**：
- 用户只关心业务配置（表单），不感知采集器差异。
- 同一任务可生成 vmagent / alloy / prometheus 多种格式配置。
- 模板引擎可复用，新增采集器只需添加 config_template 条目。

### 6.4 节点分配策略

**决策**：节点分配由区协调器(I1)通过 gossip 一致性哈希自动完成，不在 T1 中处理。collection_task.node_id 字段记录协调器的分配结果，仅供查询展示。

**理由**：
- 区协调器统一管理 vmalert / vmagent / prometheus 三类节点，分配策略集中。
- gossip 一致性哈希保证节点变动时最小迁移。
- T1 只关心任务配置本身，不感知节点拓扑。

### 6.5 多存储路由

**决策**：在 collection_task 上增加 data_sink_id 字段，任务级指定数据存储目标。

**理由**：
- 不同业务可能有不同的存储需求（如敏感数据存私有 VM，普通数据存共享集群）。
- 配置生成时根据 data_sink_id 注入对应的 remote_write 端点。
- data_sink_id 为 NULL 时使用 zone 的默认 remote_write_url，向后兼容。

### 6.6 配置热加载机制

**决策**：配置变更后通过采集器适配器的 ApplyConfig 方法触发热加载，不中断采集。

**理由**：
- vmagent 支持 POST /-/reload 热加载。
- Alloy 支持 PUT /-/config 热更新。
- Prometheus 支持写入文件 + POST /-/reload。
- 多个任务同时变更时，合并为一次 reload（batch apply），减少采集器压力。
