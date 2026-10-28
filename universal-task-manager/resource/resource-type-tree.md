# M2 资源类型树

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：M2 | 职责：树形资源类型体系管理，支持父子继承、类型专属属性 Schema 与公共指标配置

---

## 一、概述

资源类型（ResourceType）是以树形结构组织的被监控资源分类体系。每个节点代表一种资源类型，子类型自动继承父类型的公共指标与属性定义。通过资源类型树，平台实现了「定义一次、处处继承」的类型管理体系。

本模块承担以下核心职责：

- 树形资源类型的创建、查询、修改与删除
- 父子继承机制：子类型自动继承父类型的公共指标
- 类型专属属性 Schema（attributes JSON 字段）
- 预置类型管理：内置基础设施、数据库、中间件等常用类型
- 类型与指标的关联管理

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 类型树 CRUD | 创建、查询、更新、删除资源类型节点 |
| 父子继承 | 子类型自动继承父类型的公共指标与属性 |
| 属性 Schema | 定义每种类型的专属属性结构（JSON Schema） |
| 公共指标管理 | 为类型关联 MetricDefinition，形成公共指标集 |
| 预置类型 | 系统内置常用资源类型模板 |
| 类型树查询 | 返回完整的树形结构，支持懒加载 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 实例的具体管理 | M3 实例管理 |
| 指标定义（MetricDefinition） | T3 指标管理 |
| 采集任务类型（TaskType） | T1 采集任务管理 |
| 采集器配置模板 | 采集器适配器设计 |

---

## 三、功能清单

| # | 功能 | 说明 | 优先级 |
|---|------|------|--------|
| F1 | 创建资源类型 | 指定名称、父类型、属性 Schema | P1 |
| F2 | 查询资源类型树 | 返回完整树形结构或按父节点懒加载 | P1 |
| F3 | 查询类型详情 | 返回类型信息、继承的公共指标、属性 Schema | P1 |
| F4 | 更新资源类型 | 修改名称、描述、属性 Schema、排序 | P1 |
| F5 | 删除资源类型 | 需确保无子类型且无实例引用 | P1 |
| F6 | 关联公共指标 | 为类型添加/移除关联的 MetricDefinition | P1 |
| F7 | 查询公共指标集 | 返回类型自身 + 祖先链的全部公共指标（继承合并） | P1 |
| F8 | 属性 Schema 管理 | 定义和查询类型专属的属性结构 | P2 |
| F9 | 预置类型初始化 | 系统首次启动时导入预置类型树 | P1 |
| F10 | 类型搜索 | 按名称、描述模糊搜索类型节点 | P2 |
| F11 | 类型移动 | 修改节点的父类型（变更树结构） | P3 |
| F12 | 类型导出/导入 | JSON 格式批量导入导出类型定义 | P3 |

---

## 四、核心数据模型

### 4.1 resource_type 表

```sql
CREATE TABLE resource_type (
    type_id         VARCHAR(64) PRIMARY KEY,
    parent_id       VARCHAR(64) COMMENT '父类型ID, NULL=根节点',
    name            VARCHAR(128) NOT NULL COMMENT '类型名称',
    description     TEXT COMMENT '类型描述',
    icon            VARCHAR(64) COMMENT '图标标识',
    sort_order      INT DEFAULT 0 COMMENT '同级排序权重',
    attributes      JSON COMMENT '类型专属属性Schema',
    is_builtin      BOOLEAN DEFAULT FALSE COMMENT '是否预置类型',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES resource_type(type_id) ON DELETE SET NULL
);

CREATE INDEX idx_rt_parent ON resource_type(parent_id);
CREATE INDEX idx_rt_sort ON resource_type(parent_id, sort_order);
```

### 4.2 resource_type_metric 表

```sql
CREATE TABLE resource_type_metric (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    type_id         VARCHAR(64) NOT NULL COMMENT '资源类型ID',
    metric_id       VARCHAR(64) NOT NULL COMMENT '引用指标定义ID',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (type_id) REFERENCES resource_type(type_id) ON DELETE CASCADE,
    FOREIGN KEY (metric_id) REFERENCES metric_definition(metric_id),
    UNIQUE KEY uk_type_metric (type_id, metric_id)
);

CREATE INDEX idx_rtm_type ON resource_type_metric(type_id);
CREATE INDEX idx_rtm_metric ON resource_type_metric(metric_id);
```

### 4.3 继承关系示意

```
基础设施 (infra)
├── 操作系统 (infra-os)
│   ├── Linux (infra-os-linux)
│   │   ├── CentOS (infra-os-linux-centos)     ← 继承 Linux 全部指标
│   │   └── Ubuntu (infra-os-linux-ubuntu)     ← 继承 Linux 全部指标
│   └── Windows (infra-os-windows)
├── 数据库 (infra-db)
│   ├── MySQL (infra-db-mysql)                 ← 继承 数据库 公共指标
│   ├── Oracle (infra-db-oracle)
│   └── PostgreSQL (infra-db-pg)
└── 网络设备 (infra-network)
    ├── 交换机 (infra-network-switch)
    └── 路由器 (infra-network-router)

中间件 (middleware)
├── 消息队列 (middleware-mq)
│   ├── Kafka (middleware-mq-kafka)
│   └── RabbitMQ (middleware-mq-rabbitmq)
└── 缓存 (middleware-cache)
    ├── Redis (middleware-cache-redis)
    └── Memcached (middleware-cache-memcached)

应用 (application)
├── Web应用 (application-web)
│   ├── Nginx (application-web-nginx)
│   └── Tomcat (application-web-tomcat)
└── 自定义 (application-custom)
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| T3 指标管理 | MetricDefinition 列表（关联公共指标时查询） | 内部 API 调用 |

### 5.2 下游提供

| 消费模块 | 提供内容 | 交互方式 |
|----------|----------|----------|
| M3 实例管理 | 资源类型树、类型详情、公共指标集 | 查询 API |
| T1 采集任务 | 根据资源类型推荐适用的任务类型 | 查询 API |
| T3 指标管理 | 类型与指标的关联关系 | 数据关联 |
| A1 告警策略 | 根据资源类型推荐适用的告警策略 | 查询 API |

### 5.3 对外 API

#### 5.3.1 创建资源类型

```
POST /api/v1/resource-types
```

**请求体：**

```json
{
  "type_id": "infra-os-linux-centos",
  "parent_id": "infra-os-linux",
  "name": "CentOS",
  "description": "CentOS Linux 操作系统",
  "icon": "centos",
  "sort_order": 1,
  "attributes": {
    "type": "object",
    "properties": {
      "kernel_version": {
        "type": "string",
        "label": "内核版本",
        "description": "Linux 内核版本号"
      },
      "package_manager": {
        "type": "string",
        "enum": ["yum", "dnf"],
        "default": "yum",
        "label": "包管理器"
      }
    }
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "type_id": "infra-os-linux-centos",
    "parent_id": "infra-os-linux",
    "name": "CentOS",
    "description": "CentOS Linux 操作系统",
    "icon": "centos",
    "sort_order": 1,
    "is_builtin": false,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.2 查询资源类型树

```
GET /api/v1/resource-types/tree
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| root_id | string | 否 | 从指定节点开始返回子树，不传则返回完整树 |
| depth | int | 否 | 返回层级深度，默认 -1（完整树） |
| include_metrics | boolean | 否 | 是否在类型节点中包含公共指标信息 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "tree": [
      {
        "type_id": "infra",
        "name": "基础设施",
        "icon": "server",
        "sort_order": 1,
        "instance_count": 500,
        "children": [
          {
            "type_id": "infra-os",
            "name": "操作系统",
            "icon": "desktop",
            "sort_order": 1,
            "instance_count": 300,
            "children": [
              {
                "type_id": "infra-os-linux",
                "name": "Linux",
                "icon": "linux",
                "sort_order": 1,
                "instance_count": 250,
                "children": [
                  {
                    "type_id": "infra-os-linux-centos",
                    "name": "CentOS",
                    "icon": "centos",
                    "sort_order": 1,
                    "instance_count": 150,
                    "children": []
                  },
                  {
                    "type_id": "infra-os-linux-ubuntu",
                    "name": "Ubuntu",
                    "icon": "ubuntu",
                    "sort_order": 2,
                    "instance_count": 100,
                    "children": []
                  }
                ]
              }
            ]
          },
          {
            "type_id": "infra-db",
            "name": "数据库",
            "icon": "database",
            "sort_order": 2,
            "instance_count": 200,
            "children": [
              {
                "type_id": "infra-db-mysql",
                "name": "MySQL",
                "icon": "mysql",
                "sort_order": 1,
                "instance_count": 120,
                "children": []
              }
            ]
          }
        ]
      }
    ]
  }
}
```

#### 5.3.3 查询类型详情

```
GET /api/v1/resource-types/{type_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "type_id": "infra-os-linux-centos",
    "parent_id": "infra-os-linux",
    "name": "CentOS",
    "description": "CentOS Linux 操作系统",
    "icon": "centos",
    "sort_order": 1,
    "is_builtin": false,
    "attributes": {
      "type": "object",
      "properties": {
        "kernel_version": { "type": "string", "label": "内核版本" },
        "package_manager": { "type": "string", "label": "包管理器" }
      }
    },
    "inherited_metrics": [
      {
        "metric_id": "node_cpu_seconds_total",
        "name": "node_cpu_seconds_total",
        "type": "counter",
        "help": "CPU 时间统计",
        "source": "infra-os-linux"
      },
      {
        "metric_id": "node_memory_MemAvailable_bytes",
        "name": "node_memory_MemAvailable_bytes",
        "type": "gauge",
        "help": "可用内存",
        "source": "infra-os"
      }
    ],
    "own_metrics": [
      {
        "metric_id": "node_centos_kernel_version",
        "name": "node_centos_kernel_version",
        "type": "gauge",
        "help": "CentOS 内核版本信息"
      }
    ],
    "instance_count": 150,
    "child_count": 0,
    "path": ["infra", "infra-os", "infra-os-linux", "infra-os-linux-centos"],
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T10:00:00Z"
  }
}
```

**说明：** `inherited_metrics` 包含从祖先链继承的全部公共指标，`source` 字段标识指标来源的类型节点。`own_metrics` 仅包含当前类型自身关联的指标。

#### 5.3.4 更新资源类型

```
PUT /api/v1/resource-types/{type_id}
```

**请求体：**

```json
{
  "name": "CentOS 7/8",
  "description": "CentOS Linux 操作系统（含 7.x 和 8.x）",
  "icon": "centos",
  "sort_order": 1,
  "attributes": { "..." }
}
```

#### 5.3.5 删除资源类型

```
DELETE /api/v1/resource-types/{type_id}
```

**前置校验：**
- 无子类型引用（child_count = 0）
- 无实例引用（instance_count = 0）

#### 5.3.6 关联公共指标

```
POST /api/v1/resource-types/{type_id}/metrics
```

**请求体：**

```json
{
  "metric_ids": ["node_cpu_seconds_total", "node_memory_MemAvailable_bytes"]
}
```

#### 5.3.7 移除公共指标关联

```
DELETE /api/v1/resource-types/{type_id}/metrics/{metric_id}
```

#### 5.3.8 查询类型的公共指标集（含继承）

```
GET /api/v1/resource-types/{type_id}/metrics?include_inherited=true
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "type_id": "infra-os-linux-centos",
    "metrics": [
      {
        "metric_id": "node_cpu_seconds_total",
        "name": "node_cpu_seconds_total",
        "type": "counter",
        "help": "CPU 时间统计",
        "inherited_from": "infra-os-linux",
        "inherited": true
      },
      {
        "metric_id": "node_centos_pkg_count",
        "name": "node_centos_pkg_count",
        "type": "gauge",
        "help": "已安装包数量",
        "inherited_from": null,
        "inherited": false
      }
    ],
    "total": 2
  }
}
```

#### 5.3.9 搜索资源类型

```
GET /api/v1/resource-types/search
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 是 | 搜索关键词（匹配名称、描述） |
| parent_id | string | 否 | 限定搜索范围（某子树下） |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

---

## 六、设计决策

### 6.1 树形结构 vs 扁平分类

**决策：** 采用树形结构（parent_id 自引用）组织资源类型，而非扁平的标签分类。

**理由：**
- 资源类型天然具有层级关系（基础设施 > 操作系统 > Linux > CentOS）
- 树形结构支持父子继承，子类型自动获得父类型的公共指标
- 前端渲染树形选择器直观，用户易于理解
- 通过 `parent_id` 单字段实现，无需额外的关联表

### 6.2 继承策略：深度优先、就近覆盖

**决策：** 子类型继承祖先链上所有公共指标，当同一指标在多个层级被关联时，以距子类型最近的祖先为准。

**理由：**
- 继承链：CentOS → Linux → 操作系统 → 基础设施
- 如果 `node_cpu_seconds_total` 在「Linux」和「操作系统」两层都关联了，CentOS 继承时以「Linux」层的定义为准
- 就近原则符合直觉：越具体的层级，定义越精确

### 6.3 attributes 字段采用 JSON Schema 格式

**决策：** `attributes` 字段使用 JSON Schema 格式定义类型专属属性，而非自定义 DSL。

**理由：**
- JSON Schema 是业界标准，前端可直接用于表单渲染与数据校验
- 无需引入额外的 Schema 解析库
- 支持嵌套对象、枚举、默认值等丰富的类型约束
- 社区工具链完善（编辑器支持、代码生成）

### 6.4 预置类型的管理策略

**决策：** 预置类型通过 `is_builtin` 标识，系统初始化时自动导入，用户可修改但不可删除。

**理由：**
- 预置类型覆盖 90% 的常见监控场景，降低用户初始配置成本
- 允许用户修改预置类型的属性与指标关联，适应个性化需求
- 禁止删除预置类型，避免破坏已创建的实例关联

### 6.5 ON DELETE SET NULL 的外键策略

**决策：** `parent_id` 外键采用 `ON DELETE SET NULL`，而非 `CASCADE`。

**理由：**
- 删除父类型时，子类型不应被级联删除（子类型可能仍有实例引用）
- 子类型的 `parent_id` 置为 NULL，成为新的根节点，保留数据完整性
- 删除操作是低频高危操作，平台应优先保证数据安全

### 6.6 预置类型清单

**决策：** 系统内置以下预置类型树：

```
基础设施 (infra)
├── 操作系统 (infra-os)
│   ├── Linux (infra-os-linux)
│   │   ├── CentOS (infra-os-linux-centos)
│   │   ├── Ubuntu (infra-os-linux-ubuntu)
│   │   └── Debian (infra-os-linux-debian)
│   └── Windows (infra-os-windows)
│       └── Windows Server (infra-os-windows-server)
├── 数据库 (infra-db)
│   ├── MySQL (infra-db-mysql)
│   ├── Oracle (infra-db-oracle)
│   ├── PostgreSQL (infra-db-pg)
│   └── SQL Server (infra-db-mssql)
└── 网络设备 (infra-network)
    ├── 交换机 (infra-network-switch)
    └── 路由器 (infra-network-router)

中间件 (middleware)
├── 消息队列 (middleware-mq)
│   ├── Kafka (middleware-mq-kafka)
│   ├── RabbitMQ (middleware-mq-rabbitmq)
│   └── RocketMQ (middleware-mq-rocketmq)
├── 缓存 (middleware-cache)
│   ├── Redis (middleware-cache-redis)
│   └── Memcached (middleware-cache-memcached)
└── Web服务器 (middleware-web)
    ├── Nginx (middleware-web-nginx)
    ├── Apache (middleware-web-apache)
    └── Tomcat (middleware-web-tomcat)

应用 (application)
└── 自定义 (application-custom)
```
