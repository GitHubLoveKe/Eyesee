# M3 实例管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：M3 | 职责：被监控实例的全生命周期管理，包括 IP 关联、类型关联、多 Sheet 详情页与状态生命周期

---

## 一、概述

实例（Instance）是平台中被监控的具体实体，如一台 Linux 服务器、一个 MySQL 数据库、一台网络交换机。每个实例关联一个资源类型（来自类型树）和一个网区，通过 host 字段（IP 或域名）自动匹配网段确定归属。

本模块承担以下核心职责：

- 实例的创建、查询、修改与删除
- IP 关联：通过 host 字段匹配网段，自动确定或推荐网区
- 类型关联：关联 resource_type，继承公共指标与属性 Schema
- 多 Sheet 详情页：基本信息、采集配置、告警策略、指标覆盖
- 实例状态生命周期管理：pending → active → suspended → decommissioned

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 实例 CRUD | 创建、查询、更新、删除实例 |
| IP 关联 | 根据 host 匹配网段，推荐或自动分配网区 |
| 类型关联 | 绑定资源类型，加载继承的公共指标与属性 |
| 多 Sheet 详情页 | 聚合实例各维度信息（基本信息、采集、告警、指标） |
| 状态生命周期 | pending → active → suspended → decommissioned 流转 |
| 批量操作 | 批量导入、批量状态变更 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 网区与网段管理 | M1 网区与网段管理 |
| 资源类型定义 | M2 资源类型树 |
| 凭据存储与管理 | M4 凭据管理 |
| 采集任务配置 | T1 采集任务管理 |
| 告警策略配置 | A1 告警策略 |
| 指标定义 | T3 指标管理 |

---

## 三、功能清单

| # | 功能 | 说明 | 优先级 |
|---|------|------|--------|
| F1 | 创建实例 | 指定名称、host、port、资源类型、网区、凭据 | P1 |
| F2 | 查询实例列表 | 支持按网区、类型、状态、标签筛选 | P1 |
| F3 | 查询实例详情（多 Sheet） | 聚合基本信息、采集配置、告警策略、指标覆盖 | P1 |
| F4 | 更新实例 | 修改实例属性、标签、凭据关联等 | P1 |
| F5 | 删除实例 | 级联删除关联的采集任务与告警策略 | P1 |
| F6 | IP 自动关联 | 创建时根据 host 匹配网段，推荐网区 | P1 |
| F7 | 类型继承加载 | 关联资源类型后，自动加载公共指标与属性 Schema | P1 |
| F8 | 状态流转 | pending → active → suspended → decommissioned | P1 |
| F9 | 连通性测试 | 验证实例的可达性与认证有效性 | P2 |
| F10 | 批量导入 | CSV/JSON 格式批量导入实例 | P2 |
| F11 | 批量导出 | 导出实例列表为 CSV/JSON | P3 |
| F12 | 实例搜索 | 按名称、host、标签全文搜索 | P2 |
| F13 | 指标覆盖视图 | 展示实例实际产出的指标与声明指标的对比 | P2 |

---

## 四、核心数据模型

### 4.1 instance 表

```sql
CREATE TABLE instance (
    instance_id     VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '所属网区ID',
    resource_type_id VARCHAR(64) NOT NULL COMMENT '关联资源类型ID',
    name            VARCHAR(256) NOT NULL COMMENT '实例名称',
    host            VARCHAR(256) NOT NULL COMMENT 'IP地址或域名',
    port            INT COMMENT '服务端口',
    credential_id   VARCHAR(64) COMMENT '关联凭据ID',
    labels          JSON COMMENT '自定义标签',
    extra           JSON COMMENT '类型专属扩展属性',
    status          VARCHAR(32) DEFAULT 'pending' COMMENT 'pending/active/suspended/decommissioned',
    last_test_at    TIMESTAMP NULL COMMENT '最近一次连通性测试时间',
    last_test_result VARCHAR(32) COMMENT '最近测试结果: passed/failed/timeout',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id),
    FOREIGN KEY (resource_type_id) REFERENCES resource_type(type_id),
    FOREIGN KEY (credential_id) REFERENCES credential(credential_id) ON DELETE SET NULL
);

CREATE INDEX idx_inst_zone ON instance(zone_id);
CREATE INDEX idx_inst_type ON instance(resource_type_id);
CREATE INDEX idx_inst_host ON instance(host);
CREATE INDEX idx_inst_status ON instance(status);
CREATE INDEX idx_inst_credential ON instance(credential_id);
```

### 4.2 状态生命周期

```
                  ┌──────────────────────────────────────────┐
                  │                                          │
    创建实例      v       激活           暂停          停用   │
  ┌────────┐  ┌────────┐  ┌────────┐  ┌───────────┐  ┌──────────────┐
  │pending │→│ active │→│suspended│→│decommissioned│
  └────────┘  └────────┘  └────────┘  └──────────────┘
                  │          │
                  │  恢复    │  恢复
                  │←─────────│
                  │
                  │  直接停用
                  │←─────────────────────┘
```

| 状态 | 含义 | 允许操作 |
|------|------|----------|
| `pending` | 已创建但未激活，等待首次连通性测试 | 激活、测试、编辑、删除 |
| `active` | 正常运行中，采集任务可执行 | 暂停、测试、编辑 |
| `suspended` | 暂停状态，采集任务暂停执行 | 恢复、编辑、停用 |
| `decommissioned` | 终态，已停用，数据保留但不可操作 | 查看（只读） |

### 4.3 关联关系

```
resource_type (1) ──< (N) instance >── (1) zone
       │                    │                │
       │                    │                ├──< (N) network_segment (IP 匹配)
       │                    │                │
       ├── 继承公共指标      ├──< (N) collection_task
       ├── 属性 Schema       ├──< (N) alert_strategy (通过关联表)
       └── extra 字段约束    └── (1) credential (引用)
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| M1 网区与网段 | 网区列表、IP 匹配网段推荐 | 查询 API |
| M2 资源类型树 | 类型详情、公共指标集、属性 Schema | 查询 API |
| M4 凭据管理 | 凭据列表、凭据有效性校验 | 查询 API |
| M5 数据存储 | 实例关联的数据存储信息 | 查询 API |

### 5.2 下游提供

| 消费模块 | 提供内容 | 交互方式 |
|----------|----------|----------|
| T1 采集任务 | 实例列表、实例详情（host/port/credential） | 数据关联 |
| A1 告警策略 | 实例列表（用于策略作用域绑定） | 数据关联 |
| I3 状态缓存 | 实例状态变更事件 | 事件通知 |
| T4 服务发现 | 实例列表（生成 target） | 查询 API |

### 5.3 对外 API

#### 5.3.1 创建实例

```
POST /api/v1/zones/{zone_id}/instances
```

**请求体：**

```json
{
  "instance_id": "inst-mysql-prod-001",
  "resource_type_id": "infra-db-mysql",
  "name": "MySQL 生产主库",
  "host": "10.0.1.5",
  "port": 3306,
  "credential_id": "cred-mysql-001",
  "labels": {
    "env": "prod",
    "role": "primary",
    "team": "dba"
  },
  "extra": {
    "version": "8.0.35",
    "database": "mydb",
    "socket": "/var/run/mysqld/mysqld.sock"
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "instance_id": "inst-mysql-prod-001",
    "zone_id": "zone-east-1",
    "resource_type_id": "infra-db-mysql",
    "resource_type_name": "MySQL",
    "name": "MySQL 生产主库",
    "host": "10.0.1.5",
    "port": 3306,
    "credential_id": "cred-mysql-001",
    "labels": { "env": "prod", "role": "primary", "team": "dba" },
    "extra": { "version": "8.0.35", "database": "mydb" },
    "status": "pending",
    "matched_segment": {
      "segment_id": 1,
      "cidr": "10.0.1.0/24",
      "segment_name": "生产网段A"
    },
    "inherited_metrics_count": 12,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**创建时自动执行：**
1. 根据 host 匹配网段，验证 IP 属于该网区的某个网段
2. 加载资源类型的属性 Schema，校验 extra 字段合法性
3. 统计继承的公共指标数量

#### 5.3.2 查询实例列表

```
GET /api/v1/zones/{zone_id}/instances
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| resource_type_id | string | 否 | 按资源类型过滤 |
| status | string | 否 | 按状态过滤 |
| label | string | 否 | 按标签过滤，格式: `key=value` |
| keyword | string | 否 | 搜索名称或 host |
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "instance_id": "inst-mysql-prod-001",
        "name": "MySQL 生产主库",
        "host": "10.0.1.5",
        "port": 3306,
        "resource_type": {
          "type_id": "infra-db-mysql",
          "name": "MySQL",
          "path": ["基础设施", "数据库", "MySQL"]
        },
        "status": "active",
        "task_count": 3,
        "alert_count": 2,
        "last_test_result": "passed",
        "labels": { "env": "prod", "role": "primary" }
      }
    ],
    "total": 120,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询实例详情（多 Sheet）

```
GET /api/v1/instances/{instance_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "instance_id": "inst-mysql-prod-001",
    "zone_id": "zone-east-1",
    "zone_name": "华东区",
    "resource_type_id": "infra-db-mysql",
    "resource_type_name": "MySQL",
    "resource_type_path": ["基础设施", "数据库", "MySQL"],
    "name": "MySQL 生产主库",
    "host": "10.0.1.5",
    "port": 3306,
    "credential_id": "cred-mysql-001",
    "credential_name": "MySQL 生产凭据",
    "labels": { "env": "prod", "role": "primary", "team": "dba" },
    "extra": { "version": "8.0.35", "database": "mydb" },
    "status": "active",
    "last_test_at": "2026-10-27T10:05:25Z",
    "last_test_result": "passed",
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T12:00:00Z",

    "sheets": {
      "basic_info": {
        "zone": { "zone_id": "zone-east-1", "name": "华东区" },
        "segment": { "segment_id": 1, "cidr": "10.0.1.0/24", "name": "生产网段A" },
        "resource_type": {
          "type_id": "infra-db-mysql",
          "name": "MySQL",
          "path": ["基础设施", "数据库", "MySQL"],
          "attributes_schema": { "..." }
        }
      },

      "collection": {
        "tasks": [
          {
            "task_id": "task-001",
            "type_id": "mysql-standard",
            "type_name": "MySQL 标准采集",
            "name": "MySQL 生产采集",
            "enabled": true,
            "status": "running",
            "scrape_interval": "15s",
            "last_scrape_at": "2026-10-27T10:30:15Z",
            "last_error": null
          }
        ],
        "total_tasks": 3,
        "running_tasks": 3
      },

      "alert": {
        "strategies": [
          {
            "strategy_id": "strategy-001",
            "name": "MySQL 连接数过高",
            "metric_name": "mysql_global_status_threads_connected",
            "severity": "warning",
            "enabled": true,
            "status": "active"
          }
        ],
        "total_strategies": 2,
        "active_alerts": 0
      },

      "metrics": {
        "inherited": [
          {
            "metric_id": "mysql_up",
            "name": "mysql_up",
            "type": "gauge",
            "help": "MySQL 实例是否可达",
            "source": "infra-db"
          }
        ],
        "actual": [
          {
            "name": "mysql_up",
            "value": 1,
            "timestamp": "2026-10-27T10:30:15Z"
          },
          {
            "name": "mysql_global_status_connections",
            "value": 152,
            "timestamp": "2026-10-27T10:30:15Z"
          }
        ],
        "coverage": {
          "declared": 50,
          "actual": 42,
          "rate": "84%"
        }
      }
    }
  }
}
```

#### 5.3.4 更新实例

```
PUT /api/v1/instances/{instance_id}
```

**请求体：**

```json
{
  "name": "MySQL 生产主库（更新）",
  "host": "10.0.1.5",
  "port": 3306,
  "credential_id": "cred-mysql-002",
  "labels": { "env": "prod", "role": "primary", "team": "dba-new" },
  "extra": { "version": "8.0.36", "database": "mydb" }
}
```

#### 5.3.5 删除实例

```
DELETE /api/v1/instances/{instance_id}
```

**级联删除：**
- 关联的采集任务（collection_task）
- 关联的告警策略绑定（alert_strategy_instance）
- 关联的 RC 任务引用

#### 5.3.6 状态流转

```
POST /api/v1/instances/{instance_id}/status
```

**请求体：**

```json
{
  "action": "activate"
}
```

**支持的 action：**

| action | 源状态 | 目标状态 | 说明 |
|--------|--------|----------|------|
| `activate` | pending | active | 激活实例，开始采集 |
| `suspend` | active | suspended | 暂停实例，暂停采集任务 |
| `resume` | suspended | active | 恢复实例，重启采集任务 |
| `decommission` | active / suspended | decommissioned | 停用实例（终态） |

#### 5.3.7 连通性测试

```
POST /api/v1/instances/{instance_id}/test
```

**请求体：**

```json
{
  "test_type": "full",
  "timeout": "30s"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "test_id": "test-001",
    "status": "running",
    "started_at": "2026-10-27T10:05:00Z"
  }
}
```

#### 5.3.8 查询测试结果

```
GET /api/v1/instances/{instance_id}/test/{test_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "test_id": "test-001",
    "status": "passed",
    "connectivity": { "status": "ok", "latency_ms": 15 },
    "authentication": { "status": "ok" },
    "collection": { "status": "ok", "metric_count": 42 },
    "completed_at": "2026-10-27T10:05:25Z"
  }
}
```

#### 5.3.9 IP 自动推荐

```
POST /api/v1/instances/recommend-zone
```

**请求体：**

```json
{
  "host": "10.0.1.5"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "host": "10.0.1.5",
    "matched": true,
    "recommendations": [
      {
        "zone_id": "zone-east-1",
        "zone_name": "华东区",
        "system_id": "bs-ecommerce",
        "system_name": "电商平台",
        "segment": { "segment_id": 1, "cidr": "10.0.1.0/24" },
        "confidence": "exact"
      }
    ]
  }
}
```

#### 5.3.10 批量导入

```
POST /api/v1/zones/{zone_id}/instances/import
```

**请求体：**

```json
{
  "format": "csv",
  "data": "name,resource_type_id,host,port,labels\nMySQL-01,infra-db-mysql,10.0.1.5,3306,\"{env:prod}\"\nLinux-01,infra-os-linux-centos,10.0.1.10,9100,\"{env:prod}\""
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "total": 2,
    "success": 2,
    "failed": 0,
    "errors": [],
    "imported_ids": ["inst-001", "inst-002"]
  }
}
```

#### 5.3.11 跨网区实例查询

```
GET /api/v1/instances
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| system_id | string | 否 | 按业务系统过滤 |
| zone_id | string | 否 | 按网区过滤 |
| resource_type_id | string | 否 | 按资源类型过滤 |
| status | string | 否 | 按状态过滤 |
| host | string | 否 | 按 host 精确匹配 |
| keyword | string | 否 | 搜索名称或 host |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

---

## 六、设计决策

### 6.1 多 Sheet 详情页设计

**决策：** 实例详情 API 返回结构化的多 Sheet 数据（basic_info / collection / alert / metrics），前端按 Sheet 渲染标签页。

**理由：**
- 实例信息维度多（基本信息、采集配置、告警策略、指标覆盖），单页展示信息过载
- Sheet 化设计支持按需加载，减少首次渲染时间
- 每个 Sheet 可独立刷新，采集任务状态和指标数据变化频繁，需要实时更新
- 与资源类型的 `detail_panels` 定义对应，支持动态扩展

### 6.2 IP 关联策略

**决策：** 创建实例时，根据 host 字段自动匹配网段并推荐网区，但不强制绑定（用户可手动指定）。

**理由：**
- 自动推荐减少用户手动选择网区的工作量
- 不强制绑定保留灵活性：某些实例可能使用域名而非 IP
- 推荐结果附带置信度（exact/subnet），帮助用户判断

### 6.3 extra 字段的类型约束

**决策：** `extra` 字段存储类型专属扩展属性，其 Schema 由关联的 `resource_type.attributes` 定义，平台负责校验。

**理由：**
- 不同资源类型需要不同的扩展属性（MySQL 需要 version/database，Linux 需要 kernel_version）
- JSON 字段提供灵活性，Schema 校验保证数据质量
- 前端可根据 Schema 动态渲染表单字段

### 6.4 状态生命周期设计

**决策：** 四态模型（pending → active → suspended → decommissioned），decommissioned 为不可逆终态。

**理由：**
- pending：实例创建后需要首次测试验证，通过后才进入 active
- active：正常运行状态，采集任务可执行
- suspended：临时停用，保留配置但暂停采集，适用于维护窗口
- decommissioned：永久停用，数据保留但不可操作，满足审计需求
- 不可逆终态防止误操作导致数据丢失

### 6.5 级联删除策略

**决策：** 删除实例时级联删除采集任务和告警策略绑定，但凭据不级联删除。

**理由：**
- 采集任务与实例是强绑定关系，实例不存在则任务无意义
- 告警策略通过关联表绑定，删除关联记录即可，策略本身保留
- 凭据可能被多个实例引用，不可级联删除（外键设置为 ON DELETE SET NULL）

### 6.6 实例与凭据的引用关系

**决策：** 实例通过 `credential_id` 引用凭据，凭据数据不存储在实例表中。

**理由：**
- 凭据是敏感数据，独立管理便于加密存储和访问控制
- 同一凭据可被多个实例引用（如同一网区的多台 MySQL 使用相同账号）
- 凭据更新（版本变更）不影响实例配置
- 详见 M4 凭据管理模块
