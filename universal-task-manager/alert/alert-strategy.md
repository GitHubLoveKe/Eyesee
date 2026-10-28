# A1 告警策略模块

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：A1 | 职责：定义告警规则的策略模板，管理策略与实例的关联关系，驱动告警规则向各网区的下发与生效。

---

## 一、概述

告警策略模块是告警层的规则定义中心。它将"指标 + 表达式模板 + 持续时间 + 严重级别"封装为可复用的策略单元，通过作用域机制（instance / type / zone）灵活控制规则的生效范围，并借助 `alert_strategy_instance` 关联表实现按网区拆分的规则下发。

策略模块不直接执行告警评估，而是将策略翻译为具体的告警规则（RCRule），推送至底层告警引擎（默认实现为 vmalert）执行。这种"策略定义"与"规则执行"的分离，使得同一策略可以适配不同的告警引擎后端。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 策略 CRUD | 创建、查询、更新、删除告警策略 |
| 策略模板管理 | 维护 expr_template 及其变量定义 |
| 作用域管理 | 管理 instance / type / zone 三级作用域 |
| 策略-实例关联 | 维护 alert_strategy_instance 关联表 |
| 规则生成与下发 | 将策略翻译为 RCRule 并推送至告警引擎 |
| 策略状态管理 | 管理策略生命周期（draft → active → paused → archived） |
| 策略批量操作 | 按网区、按类型批量启用/禁用/下发策略 |

### 本模块不负责

| 职责 | 归属 |
|------|------|
| 告警规则的实际执行与评估 | A4 告警收敛模块（AlertEngine） |
| 告警事件的通知与投递 | A3 通知渠道模块（NotificationDispatcher） |
| 告警组的维护与路由 | A2 告警组模块 |
| 指标定义的维护 | 任务管理层（MetricDefinition） |
| 采集任务的配置与调度 | 任务管理层（CollectionTask / RCTask） |

---

## 三、功能清单

| 功能编号 | 功能名称 | 优先级 | 说明 |
|----------|----------|--------|------|
| A1-F01 | 策略创建 | P0 | 创建告警策略，关联指标与表达式模板 |
| A1-F02 | 策略查询 | P0 | 支持按名称、指标、严重级别、状态、作用域等多维查询 |
| A1-F03 | 策略更新 | P0 | 修改策略定义，触发规则重新下发 |
| A1-F04 | 策略删除 | P0 | 删除策略并级联清理关联关系与下发规则 |
| A1-F05 | 策略启用/禁用 | P0 | 切换策略状态，联动规则的加载/卸载 |
| A1-F06 | 作用域绑定 | P0 | 将策略绑定到具体实例、资源类型或网区 |
| A1-F07 | 策略-实例关联管理 | P0 | 维护 alert_strategy_instance 关联，支持批量绑定/解绑 |
| A1-F08 | 表达式模板渲染 | P0 | 运行时将模板变量替换为实际值，生成可执行的 PromQL |
| A1-F09 | 规则生成 | P0 | 将渲染后的策略翻译为 RCRule 写入 rc_rule 表 |
| A1-F10 | 规则下发 | P0 | 调用 AlertEngine.PushRules 将规则推送至指定网区 |
| A1-F11 | 规则撤回 | P1 | 策略禁用/删除时，调用 AlertEngine.RemoveRules 撤回已下发规则 |
| A1-F12 | 策略预览 | P1 | 预览策略渲染后的完整规则内容，不实际下发 |
| A1-F13 | 策略导入/导出 | P2 | 支持 JSON/YAML 格式的策略批量导入导出 |
| A1-F14 | 策略版本历史 | P2 | 记录策略变更历史，支持回溯 |
| A1-F15 | 策略模板库 | P2 | 内置常用告警策略模板（如 CPU > 90%、磁盘 > 85% 等） |

---

## 四、核心数据模型

### 4.1 alert_strategy（告警策略表）

> 引用 data-model.md 4.1 节定义，此处列出关键字段说明。

```sql
CREATE TABLE alert_strategy (
    strategy_id     VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL,
    description     TEXT,
    metric_id       VARCHAR(64) NOT NULL COMMENT '关联指标定义',
    expr_template   TEXT NOT NULL COMMENT '告警表达式模板，含 $variable 占位符',
    duration        VARCHAR(32) DEFAULT '5m' COMMENT '持续时间，如 5m/10s/1h',
    severity        VARCHAR(32) NOT NULL DEFAULT 'warning' COMMENT 'critical/warning/info',
    scope_type      VARCHAR(32) NOT NULL COMMENT 'instance/type/zone',
    scope_value     TEXT COMMENT '作用域值，JSON数组格式',
    alert_group_id  VARCHAR(64) COMMENT '关联告警组',
    labels          JSON COMMENT '附加标签，透传至告警事件',
    annotations     JSON COMMENT '注释模板，支持 Go template 语法',
    enabled         BOOLEAN DEFAULT TRUE,
    status          VARCHAR(32) DEFAULT 'draft' COMMENT 'draft/active/paused/archived',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (metric_id) REFERENCES metric_definition(metric_id)
);

CREATE INDEX idx_strategy_metric ON alert_strategy(metric_id);
CREATE INDEX idx_strategy_group ON alert_strategy(alert_group_id);
CREATE INDEX idx_strategy_status ON alert_strategy(status);
```

**关键字段说明：**

| 字段 | 说明 |
|------|------|
| `expr_template` | PromQL 表达式模板，支持 `$instance`、`$type`、`$zone` 等变量占位符。运行时根据作用域替换为实际值 |
| `scope_type` | `instance` = 逐实例下发独立规则；`type` = 按资源类型统一下发；`zone` = 网区级别统一下发 |
| `scope_value` | JSON 数组。instance 类型为实例 ID 列表，type 类型为资源类型 ID 列表，zone 类型为网区 ID 列表 |
| `labels` | 附加标签，会合并到告警事件的 labels 中，用于路由和过滤 |
| `annotations` | 注释模板，支持 `{{ $labels.instance }}` 等 Go template 变量，用于生成告警描述 |

### 4.2 alert_strategy_instance（策略-实例关联表）

> 引用 data-model.md 4.2 节定义。

```sql
CREATE TABLE alert_strategy_instance (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    strategy_id     VARCHAR(64) NOT NULL,
    instance_id     VARCHAR(64) NOT NULL,
    zone_id         VARCHAR(64) NOT NULL COMMENT '冗余存储，用于按网区拆分下发',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (strategy_id) REFERENCES alert_strategy(strategy_id) ON DELETE CASCADE,
    FOREIGN KEY (instance_id) REFERENCES instance(instance_id) ON DELETE CASCADE,
    UNIQUE KEY uk_strategy_instance (strategy_id, instance_id)
);

CREATE INDEX idx_asi_zone ON alert_strategy_instance(zone_id);
```

**设计说明：**

- `zone_id` 冗余存储是为了在规则下发时，能够高效地按网区查询关联的实例列表，避免跨表 JOIN。
- 唯一约束 `(strategy_id, instance_id)` 防止重复关联。
- 级联删除：策略删除时自动清理关联；实例删除时自动清理关联。

### 4.3 关联模型引用

策略模块涉及的其他表定义引用 data-model.md：

| 表名 | 引用章节 | 用途 |
|------|----------|------|
| `metric_definition` | 3.5 节 | 策略关联的指标定义，提供 promql_template 和变量定义 |
| `rc_task` | 3.3 节 | 策略下发的目标 RC 任务，strategy_id 字段关联来源策略 |
| `rc_rule` | 3.4 节 | 策略生成的具体告警规则，存储渲染后的 PromQL 表达式 |
| `instance` | 2.5 节 | 策略作用的实例实体 |
| `alert_group` | 4.3 节 | 策略关联的告警组，决定告警路由目标 |

---

## 五、接口与交互

### 5.1 上游依赖

| 上游模块 | 依赖接口 | 用途 |
|----------|----------|------|
| 指标管理层 | MetricDefinition 查询 | 获取指标定义、promql_template、变量声明 |
| 核心资源层 | Instance 查询 | 获取实例信息，用于作用域解析和规则渲染 |
| 核心资源层 | Zone 查询 | 获取网区信息，用于 zone 级作用域 |
| 核心资源层 | ResourceType 查询 | 获取资源类型信息，用于 type 级作用域 |
| A2 告警组模块 | AlertGroup 查询 | 获取告警组信息，建立策略→组的路由关系 |

### 5.2 下游提供

| 下游模块 | 提供接口 | 用途 |
|----------|----------|------|
| A4 告警收敛模块 | AlertEngine.PushRules / RemoveRules | 推送/撤回告警规则至告警引擎 |
| 任务管理层 | RCTask / RCRule 写入 | 创建或更新 RC 任务与规则记录 |
| A2 告警组模块 | alert_strategy.alert_group_id | 告警组可通过策略 ID 反查关联的策略列表 |

### 5.3 对外 API

#### 5.3.1 创建告警策略

```
POST /api/v1/alert-strategies
```

**请求体：**

```json
{
  "name": "MySQL连接数过高",
  "description": "当MySQL连接数超过阈值的90%时触发告警",
  "metric_id": "mysql_threads_connected",
  "expr_template": "mysql_threads_connected{instance=\"$instance\"} > $threshold * 0.9",
  "duration": "5m",
  "severity": "warning",
  "scope_type": "instance",
  "scope_value": ["inst-mysql-001", "inst-mysql-002"],
  "alert_group_id": "group-dba",
  "labels": {
    "team": "dba",
    "component": "mysql"
  },
  "annotations": {
    "summary": "MySQL {{ $labels.instance }} 连接数过高",
    "description": "当前连接数 {{ $value }}，超过阈值的 90%"
  }
}
```

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "strategy_id": "strategy-001",
    "name": "MySQL连接数过高",
    "status": "draft",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.2 查询告警策略列表

```
GET /api/v1/alert-strategies
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| metric_id | string | 否 | 按关联指标过滤 |
| severity | string | 否 | 按严重级别过滤：critical/warning/info |
| scope_type | string | 否 | 按作用域类型过滤：instance/type/zone |
| status | string | 否 | 按状态过滤：draft/active/paused/archived |
| alert_group_id | string | 否 | 按告警组过滤 |
| zone_id | string | 否 | 按网区过滤（查询该网区关联的策略） |
| keyword | string | 否 | 按名称/描述模糊搜索 |
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "strategy_id": "strategy-001",
        "name": "MySQL连接数过高",
        "metric_id": "mysql_threads_connected",
        "severity": "warning",
        "scope_type": "instance",
        "scope_value": ["inst-mysql-001", "inst-mysql-002"],
        "alert_group_id": "group-dba",
        "enabled": true,
        "status": "active",
        "instance_count": 2,
        "created_at": "2026-10-27T10:00:00Z"
      }
    ],
    "total": 15,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询单个告警策略

```
GET /api/v1/alert-strategies/{strategy_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "strategy_id": "strategy-001",
    "name": "MySQL连接数过高",
    "description": "当MySQL连接数超过阈值的90%时触发告警",
    "metric_id": "mysql_threads_connected",
    "expr_template": "mysql_threads_connected{instance=\"$instance\"} > $threshold * 0.9",
    "duration": "5m",
    "severity": "warning",
    "scope_type": "instance",
    "scope_value": ["inst-mysql-001", "inst-mysql-002"],
    "alert_group_id": "group-dba",
    "labels": { "team": "dba", "component": "mysql" },
    "annotations": {
      "summary": "MySQL {{ $labels.instance }} 连接数过高",
      "description": "当前连接数 {{ $value }}，超过阈值的 90%"
    },
    "enabled": true,
    "status": "active",
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T10:30:00Z"
  }
}
```

#### 5.3.4 更新告警策略

```
PUT /api/v1/alert-strategies/{strategy_id}
```

**请求体：** 同创建，字段可选（部分更新）。

**注意：** 策略更新后，若状态为 `active`，自动触发规则重新渲染与下发。

#### 5.3.5 删除告警策略

```
DELETE /api/v1/alert-strategies/{strategy_id}
```

**行为：**

1. 标记策略状态为 `archived`（软删除）或直接删除（硬删除，可配置）。
2. 级联删除 `alert_strategy_instance` 关联记录。
3. 调用 AlertEngine.RemoveRules 撤回已下发的规则。
4. 清理关联的 RCRule 记录。

#### 5.3.6 启用/禁用策略

```
POST /api/v1/alert-strategies/{strategy_id}/enable
POST /api/v1/alert-strategies/{strategy_id}/disable
```

**行为：**

- 启用：设置 `enabled=true`，触发规则下发。
- 禁用：设置 `enabled=false`，撤回已下发规则。

#### 5.3.7 管理策略-实例关联

```
# 批量绑定实例
POST /api/v1/alert-strategies/{strategy_id}/instances/bind
{
  "instance_ids": ["inst-001", "inst-002", "inst-003"]
}

# 批量解绑实例
POST /api/v1/alert-strategies/{strategy_id}/instances/unbind
{
  "instance_ids": ["inst-003"]
}

# 查询策略关联的实例列表
GET /api/v1/alert-strategies/{strategy_id}/instances

# 查询实例关联的策略列表
GET /api/v1/instances/{instance_id}/alert-strategies
```

#### 5.3.8 预览策略规则

```
POST /api/v1/alert-strategies/{strategy_id}/preview
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "strategy_id": "strategy-001",
    "rendered_rules": [
      {
        "instance_id": "inst-mysql-001",
        "zone_id": "zone-east-1",
        "alert_name": "MySQL连接数过高",
        "expr": "mysql_threads_connected{instance=\"inst-mysql-001\"} > 500 * 0.9",
        "duration": "5m",
        "severity": "warning",
        "labels": { "team": "dba", "instance": "inst-mysql-001" },
        "annotations": {
          "summary": "MySQL inst-mysql-001 连接数过高",
          "description": "当前连接数 {{ $value }}，超过阈值的 90%"
        }
      },
      {
        "instance_id": "inst-mysql-002",
        "zone_id": "zone-east-1",
        "alert_name": "MySQL连接数过高",
        "expr": "mysql_threads_connected{instance=\"inst-mysql-002\"} > 500 * 0.9",
        "duration": "5m",
        "severity": "warning"
      }
    ]
  }
}
```

#### 5.3.9 手动触发规则下发

```
POST /api/v1/alert-strategies/{strategy_id}/deploy
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "strategy_id": "strategy-001",
    "deploy_status": "success",
    "zones_affected": ["zone-east-1"],
    "rules_pushed": 2,
    "deployed_at": "2026-10-27T11:00:00Z"
  }
}
```

#### 5.3.10 策略导入/导出

```
# 导出策略（支持 JSON/YAML）
GET /api/v1/alert-strategies/export?format=json&strategy_ids=strategy-001,strategy-002

# 导入策略
POST /api/v1/alert-strategies/import
Content-Type: multipart/form-data
{
  "file": <策略定义文件>,
  "format": "json"
}
```

---

## 六、设计决策

### 6.1 策略模板与变量替换机制

**决策：** 使用 `$variable` 占位符语法（如 `$instance`、`$threshold`），运行时根据作用域类型替换为实际值。

**理由：**

- 与 Prometheus/vmalert 的 PromQL 语法兼容，避免引入额外模板引擎。
- 变量定义可复用 metric_definition 中的 `variables` 字段，保持元数据一致性。
- 支持用户在创建策略时覆盖变量的默认值。

**变量替换规则：**

| 变量 | 来源 | 说明 |
|------|------|------|
| `$instance` | scope_value / alert_strategy_instance | 实例 ID 或实例标签 |
| `$type` | scope_value | 资源类型 ID |
| `$zone` | scope_value / instance.zone_id | 网区 ID |
| `$threshold` | 策略自定义参数或指标默认值 | 告警阈值 |
| 自定义变量 | metric_definition.variables | 指标定义的扩展变量 |

### 6.2 三级作用域设计

**决策：** 支持 instance / type / zone 三级作用域，而非仅支持实例级绑定。

**理由：**

| 作用域 | 适用场景 | 规则生成行为 |
|--------|----------|-------------|
| `instance` | 精细化告警，如实例级阈值 | 为每个实例生成独立的 RCRule，expr 中包含 instance 标签过滤 |
| `type` | 同类型实例统一告警 | 按资源类型生成一条 RCRule，expr 中通过 resource_type 标签匹配 |
| `zone` | 网区级聚合告警 | 按网区生成一条 RCRule，expr 中使用 zone 标签进行聚合 |

三级作用域覆盖了从粗到细的告警需求，同时避免了为每个实例手动配置策略的运维负担。

### 6.3 规则下发流程

**决策：** 策略 → RCRule → AlertEngine.PushRules 的三步下发流程。

**流程：**

```
策略变更（创建/更新/启用）
    │
    ▼
1. 解析作用域，确定目标实例/类型/网区
    │
    ▼
2. 渲染表达式：expr_template + 变量替换 → 具体 PromQL
    │
    ▼
3. 生成 RCRule 记录，写入 rc_rule 表
    │
    ▼
4. 按 zone_id 分组，组装 vmalert rule package
    │
    ▼
5. 调用 AlertEngine.PushRules(zoneID, rules)
    │
    ▼
6. 更新 rc_task 状态，记录下发结果
```

**理由：**

- RCRule 作为中间层，记录了策略到规则的映射关系，便于审计和排障。
- 按 zone_id 分组下发，与多网区架构对齐，每个网区的告警引擎独立接收规则。
- AlertEngine 接口抽象，使得底层引擎可替换（vmalert / Prometheus rules / Thanos Ruler）。

### 6.4 zone_id 冗余存储

**决策：** 在 alert_strategy_instance 表中冗余存储 zone_id，而非通过 instance 表 JOIN 获取。

**理由：**

- 规则下发是高频操作，按 zone_id 查询关联实例的性能至关重要。
- 避免在规则下发流程中进行多表 JOIN，降低数据库压力。
- zone_id 与 instance 的映射关系稳定（实例创建后不会跨区迁移），冗余存储的一致性风险极低。

### 6.5 策略状态机

**决策：** 策略生命周期为 `draft → active → paused → archived`。

```
         创建
          │
          ▼
       [draft] ──── 发布 ────→ [active]
          │                       │    │
          │                     暂停   删除
          │                       │    │
          │                       ▼    ▼
          │                   [paused] [archived]
          │                       │
          └──── 直接发布 ──────────┘
```

| 状态 | 说明 | 规则状态 |
|------|------|----------|
| `draft` | 草稿，尚未发布 | 未下发 |
| `active` | 已启用，规则生效中 | 已下发至告警引擎 |
| `paused` | 已暂停，临时禁用 | 已撤回 |
| `archived` | 已归档/删除 | 已撤回并清理 |
