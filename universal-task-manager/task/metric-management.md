# T3 指标管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：T3 | 职责：管理平台所有指标定义（MetricDefinition），包括 event/metric 分类、PromQL 模板与变量替换、常用查询集（CommonQuery）管理，以及指标来源的追踪与继承。

---

## 一、概述

指标管理模块是平台的元数据中枢。它定义了平台能够采集、存储、查询和告警的所有指标的元信息，包括指标名称、类型、PromQL 模板、变量定义等。上层的采集任务通过 task_type.declared_metrics 声明产出哪些指标，告警策略通过 MetricDefinition 的 PromQL 模板构建告警表达式，前端的查询面板通过 CommonQuery 提供预置查询。

本模块的核心价值在于：

- **统一指标目录**：所有指标集中定义，避免命名冲突与语义歧义。
- **分类管理**：区分 event（离散事件）与 metric（连续时序），适配不同处理逻辑。
- **PromQL 模板化**：指标定义中携带带变量占位符的 PromQL 模板，运行时替换生成具体查询。
- **常用查询集**：预置高频查询场景，降低用户上手门槛。
- **来源追溯**：指标可来源于 task_type.declared_metrics（直接声明）或 resource_type_metric（类型继承）。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 指标定义 CRUD | 创建、查询、更新、删除 MetricDefinition |
| 指标分类管理 | event / metric 两大类别，以及 gauge / counter / histogram / summary 子类型 |
| PromQL 模板管理 | 维护含 $variable 占位符的 PromQL 模板 |
| 变量定义管理 | 维护模板变量的 name / type / default 定义 |
| 常用查询集管理 | CommonQuery 的 CRUD，预置与用户自定义 |
| 指标来源追踪 | 记录指标来源于 task_type.declared_metrics 或 resource_type_metric |
| 指标标签管理 | 维护指标的 labels 定义 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 指标的实际采集 | T1 采集任务模块（通过 declared_metrics 声明） |
| 指标数据的存储 | 数据存储层（DataSink: VictoriaMetrics / Elasticsearch / Loki 等） |
| 指标数据的查询执行 | 查询引擎 / Grafana |
| 基于指标的告警规则 | T2 RC 任务模块（使用 PromQL 模板构建 expr） |
| 任务类型注册 | T1 采集任务模块 |
| 资源类型管理 | 资源管理模块 |

---

## 三、功能清单

| 编号 | 功能 | 优先级 | 说明 |
|------|------|--------|------|
| F3.1 | 指标定义创建 | P0 | 定义指标名称、分类、类型、单位、PromQL 模板、变量、标签 |
| F3.2 | 指标定义查询 | P0 | 按名称 / 分类 / 类型 / 标签等多维度查询 |
| F3.3 | 指标定义更新 | P0 | 更新模板、变量、标签等元信息 |
| F3.4 | 指标定义删除 | P1 | 删除前校验是否有告警策略或查询引用 |
| F3.5 | 指标分类筛选 | P0 | 按 event / metric 分类过滤 |
| F3.6 | PromQL 模板渲染 | P0 | 给定变量值，将模板中的 $variable 替换为具体值，输出可执行 PromQL |
| F3.7 | 变量定义管理 | P0 | 维护 [{name, type, default}] 格式的变量列表 |
| F3.8 | 变量校验 | P1 | 校验模板中所有 $variable 均有对应变量定义 |
| F3.9 | 常用查询创建 | P0 | 基于指标定义创建常用查询，含完整 PromQL 与变量 |
| F3.10 | 常用查询查询 | P0 | 按分类 / 指标 / 关键词搜索常用查询 |
| F3.11 | 常用查询更新 | P1 | 更新查询内容 |
| F3.12 | 常用查询删除 | P1 | 删除查询 |
| F3.13 | 预置查询管理 | P1 | 系统内置的常用查询（is_builtin=true），用户不可删除 |
| F3.14 | 指标来源查询 | P1 | 查询某指标被哪些 task_type / resource_type 引用 |
| F3.15 | 指标继承关系 | P2 | 展示 resource_type 树上的指标继承链 |
| F3.16 | 指标批量导入 | P2 | 从 JSON / CSV 批量导入指标定义 |
| F3.17 | 指标导出 | P2 | 导出指标定义为 JSON / Markdown |
| F3.18 | PromQL 语法校验 | P1 | 校验模板的 PromQL 语法合法性 |

---

## 四、核心数据模型

### 4.1 metric_definition（指标定义）

```sql
CREATE TABLE metric_definition (
    metric_id       VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL COMMENT '指标名，遵循 Prometheus 命名规范',
    category        VARCHAR(32) NOT NULL COMMENT 'event: 离散事件 / metric: 连续时序',
    metric_type     VARCHAR(32) COMMENT 'gauge/counter/histogram/summary，category=metric时必填',
    help            TEXT COMMENT '指标说明',
    unit            VARCHAR(64) COMMENT '单位: bytes, seconds, count, percent...',
    promql_template TEXT COMMENT 'PromQL模板, 含$variable占位符',
    variables       JSON COMMENT '变量定义: [{name, type, default, description}]',
    labels          JSON COMMENT '标签定义: [{name, description, required}]',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_metric_category ON metric_definition(category);
CREATE INDEX idx_metric_name ON metric_definition(name);
CREATE INDEX idx_metric_type ON metric_definition(metric_type);
```

**category 分类说明：**

| 分类 | 含义 | 特征 | 示例 |
|------|------|------|------|
| `event` | 离散事件 | 无连续时序，表示某个时间点发生的事 | `mysql_slow_query`（慢查询事件）、`host_reboot`（主机重启事件） |
| `metric` | 连续时序 | 随时间持续变化的数值型数据 | `mysql_connections`（连接数）、`cpu_usage_percent`（CPU 使用率） |

**variables JSON 结构：**

```json
[
  {
    "name": "instance",
    "type": "string",
    "default": "",
    "description": "实例标识，对应 instance label"
  },
  {
    "name": "window",
    "type": "duration",
    "default": "5m",
    "description": "时间窗口"
  },
  {
    "name": "threshold",
    "type": "number",
    "default": "80",
    "description": "告警阈值"
  }
]
```

**labels JSON 结构：**

```json
[
  {
    "name": "instance",
    "description": "实例标识",
    "required": true
  },
  {
    "name": "zone",
    "description": "所属网区",
    "required": true
  },
  {
    "name": "job",
    "description": "采集任务标识",
    "required": false
  }
]
```

### 4.2 common_query（常用查询集）

```sql
CREATE TABLE common_query (
    query_id        VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL COMMENT '查询名称',
    description     TEXT COMMENT '查询说明',
    metric_id       VARCHAR(64) NOT NULL COMMENT '关联指标定义',
    promql          TEXT NOT NULL COMMENT '完整PromQL(可含变量占位符)',
    variables       JSON COMMENT '查询级变量覆盖 [{name, value}]',
    category        VARCHAR(64) COMMENT '分类标签: overview/performance/error/capacity...',
    is_builtin      BOOLEAN DEFAULT FALSE COMMENT '是否系统预置',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (metric_id) REFERENCES metric_definition(metric_id)
);

CREATE INDEX idx_cq_metric ON common_query(metric_id);
CREATE INDEX idx_cq_category ON common_query(category);
CREATE INDEX idx_cq_builtin ON common_query(is_builtin);
```

### 4.3 指标来源关系（视图，非独立表）

指标的来源通过以下两条路径追溯：

**路径一：task_type.declared_metrics（直接声明）**

```
task_type (type_id = "mysql-standard")
  └── declared_metrics: [
        { "name": "mysql_up", "type": "gauge", "help": "..." },
        { "name": "mysql_connections", "type": "gauge", "help": "..." }
      ]
           │
           ▼  引用
metric_definition (metric_id = "mysql_up")
```

**路径二：resource_type_metric（类型继承）**

```
resource_type (type_id = "database")
  └── resource_type (type_id = "mysql", parent_id = "database")
        └── resource_type_metric: [
              { metric_id: "mysql_up" },
              { metric_id: "mysql_connections" }
            ]
           │
           ▼  继承
metric_definition (metric_id = "mysql_up")
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| T1 采集任务（TaskType） | task_type.declared_metrics 作为指标来源之一 | 内部服务调用 |
| 资源管理（ResourceType） | resource_type_metric 作为指标来源之二 | 内部服务调用 |

### 5.2 下游提供

| 消费方 | 提供内容 | 交互方式 |
|--------|----------|----------|
| T2 RC 任务 | PromQL 模板用于构建告警规则 expr | 内部服务调用 |
| 告警层（AlertStrategy） | expr_template 引用指标 PromQL 模板 | 内部服务调用 |
| 前端查询面板 | 常用查询集 + 指标目录 | REST API |
| Grafana | 指标元信息用于 Dashboard 模板 | REST API |

### 5.3 对外 API

#### 5.3.1 指标定义管理

**创建指标定义**

```
POST /api/v1/metrics
```

请求体：

```json
{
  "metric_id": "mysql_up",
  "name": "mysql_up",
  "category": "metric",
  "metric_type": "gauge",
  "help": "MySQL实例是否可达，1=可达，0=不可达",
  "unit": "boolean",
  "promql_template": "mysql_up{instance=\"$instance\"}",
  "variables": [
    {
      "name": "instance",
      "type": "string",
      "default": "",
      "description": "实例标识"
    }
  ],
  "labels": [
    { "name": "instance", "description": "实例标识", "required": true },
    { "name": "zone", "description": "所属网区", "required": true }
  ]
}
```

**查询指标列表**

```
GET /api/v1/metrics?category=metric&metric_type=gauge&keyword=mysql&page=1&page_size=20
```

**查询单个指标**

```
GET /api/v1/metrics/{metric_id}
```

**更新指标**

```
PUT /api/v1/metrics/{metric_id}
```

**删除指标**

```
DELETE /api/v1/metrics/{metric_id}
```

> 前置校验：无告警策略或常用查询引用该指标时方可删除。

**渲染 PromQL（变量替换）**

```
POST /api/v1/metrics/{metric_id}/render
```

请求体：

```json
{
  "variables": {
    "instance": "inst-mysql-001",
    "window": "5m"
  }
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "metric_id": "mysql_up",
    "promql_template": "mysql_up{instance=\"$instance\"}",
    "promql_rendered": "mysql_up{instance=\"=\"inst-mysql-001\"}",
    "variables_used": {
      "instance": "inst-mysql-001"
    }
  }
}
```

**查询指标来源**

```
GET /api/v1/metrics/{metric_id}/sources
```

响应：

```json
{
  "code": 0,
  "data": {
    "metric_id": "mysql_up",
    "sources": [
      {
        "source_type": "task_type_declared",
        "source_id": "mysql-standard",
        "source_name": "MySQL 标准采集"
      },
      {
        "source_type": "resource_type_metric",
        "source_id": "mysql",
        "source_name": " "
      }
    ]
  }
}
```

**查询指标继承链**

```
GET /api/v1/metrics/{metric_id}/inheritance
```

#### 5.3.2 常用查询管理

**创建常用查询**

```
POST /api/v1/common-queries
```

请求体：

```json
{
  "query_id": "cq-mysql-connections-trend",
  "name": "MySQL 连接数趋势",
  "description": "查看MySQL实例近1小时的连接数变化趋势",
  "metric_id": "mysql_global_status_threads_connected",
  "promql": "mysql_global_status_threads_connected{instance=\"$instance\"}",
  "variables": [
    { "name": "instance", "value": "inst-mysql-001" }
  ],
  "category": "performance",
  "is_builtin": false
}
```

**查询常用查询列表**

```
GET /api/v1/common-queries?metric_id=mysql_up&category=performance&is_builtin=true&page=1&page_size=20
```

**查询单个常用查询**

```
GET /api/v1/common-queries/{query_id}
```

**更新常用查询**

```
PUT /api/v1/common-queries/{query_id}
```

**删除常用查询**

```
DELETE /api/v1/common-queries/{query_id}
```

> is_builtin=true 的预置查询不可删除。

**批量导入指标定义**

```
POST /api/v1/metrics/import
```

请求体：

```json
{
  "format": "json",
  "metrics": [
    {
      "metric_id": "redis_up",
      "name": "redis_up",
      "category": "metric",
      "metric_type": "gauge",
      "help": "Redis实例是否可达",
      "promql_template": "redis_up{instance=\"$instance\"}",
      "variables": [{ "name": "instance", "type": "string", "default": "" }]
    }
  ]
}
```

**导出指标定义**

```
GET /api/v1/metrics/export?format=json&category=metric&metric_type=gauge
```

---

## 六、设计决策

### 6.1 event 与 metric 双分类

**决策**：将指标分为 event（离散事件）和 metric（连续时序）两大类。

**理由**：
- event 类型指标表示离散事件（如慢查询发生、主机重启），不适合做 rate/sum_over_time 等时序运算。
- metric 类型指标表示连续时序数据，适合 PromQL 的各种聚合与函数运算。
- 分类后，前端查询面板可针对不同类型提供不同的可视化方式（event → 事件列表，metric → 时序图）。
- 告警规则构建时，event 类型通常使用 count 或 exists 语义，metric 类型使用阈值比较语义。

### 6.2 PromQL 模板 + 变量替换

**决策**：MetricDefinition 中存储带 $variable 占位符的 PromQL 模板，运行时通过变量替换生成可执行 PromQL。

**模板示例**：

```
# 模板
rate(mysql_global_status_queries_total{instance="$instance"}[$window]) > $threshold

# 变量
{ "instance": "inst-mysql-001", "window": "5m", "threshold": "1000" }

# 渲染结果
rate(mysql_global_status_queries_total{instance="inst-mysql-001"}[5m]) > 1000
```

**理由**：
- 模板化避免为每个实例重复定义指标查询。
- 变量定义（含 type / default）提供类型安全与默认值。
- 告警策略的 expr_template 可直接引用 MetricDefinition 的 promql_template。
- 前端查询面板只需用户提供变量值即可获得完整 PromQL。

### 6.3 双来源指标追溯

**决策**：指标可来源于两条路径——task_type.declared_metrics（直接声明）和 resource_type_metric（类型继承）。

**理由**：
- declared_metrics 是任务类型自描述的产出指标，语义精确。
- resource_type_metric 是资源类型级别的公共指标定义，支持继承复用。
- 双来源使得指标目录既可以从下往上（采集任务声明）构建，也可以从上往下（资源类型规划）规划。
- 来源追踪帮助用户理解「这个指标由哪些采集任务产出」和「这个资源类型应该采集哪些指标」。

### 6.4 常用查询集设计

**决策**：CommonQuery 独立于 MetricDefinition，提供预置的查询场景。

**理由**：
- 一个指标可能有多种查询场景（当前值、趋势、同环比、Top N）。
- 常用查询封装了完整的 PromQL + 变量值，用户可直接使用或微调。
- is_builtin 标记区分系统预置（不可删除）与用户自定义。
- category 字段支持按场景分组（overview / performance / error / capacity）。

### 6.5 指标命名规范

**决策**：强制遵循 Prometheus 命名规范 `<namespace>_<subsystem>_<name>_<unit>`。

**示例**：

| 指标名 | 解析 |
|--------|------|
| `mysql_global_status_threads_connected` | namespace=mysql, subsystem=global_status, name=threads_connected |
| `redis_used_memory_bytes` | namespace=redis, name=used_memory, unit=bytes |
| `node_cpu_seconds_total` | namespace=node, name=cpu, unit=seconds_total |

**理由**：
- 统一命名规范降低认知负担，便于搜索与过滤。
- 与 Prometheus 生态工具（Grafana / Alertmanager / vmalert）无缝对接。
- 创建时进行命名规范校验，不符合规范的指标拒绝创建。

### 6.6 变量类型系统

**决策**：变量定义支持 string / number / duration / label_selector 四种类型。

| 类型 | 说明 | 示例 |
|------|------|------|
| `string` | 字符串，替换后加引号 | `$instance` → `"inst-001"` |
| `number` | 数值，替换后不加引号 | `$threshold` → `80` |
| `duration` | 时间长度，替换后保留单位 | `$window` → `5m` |
| `label_selector` | 标签选择器，支持通配符 | `$instance` → `"inst-mysql-.*"` |

**理由**：
- 类型化的变量使 PromQL 渲染结果始终语法正确。
- 前端可根据变量类型渲染不同的输入控件。
- 校验阶段可检查变量类型与值的匹配性。
