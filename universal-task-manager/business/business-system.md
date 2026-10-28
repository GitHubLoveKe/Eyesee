# B1 业务系统管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：B1 | 职责：业务系统全生命周期管理，包括系统创建、成员权限管理、资源隔离与全览视图聚合

---

## 一、概述

业务系统（BusinessSystem）是多网区监控采集管理平台的**顶层组织单元**。所有核心资源（网区、实例、采集任务、告警策略、数据存储）均归属于某个业务系统，实现资源与权限的物理级隔离。

本模块承担以下核心职责：

- 业务系统的创建、查询、修改、停用与归档
- 成员邀请与角色分配（admin / operator / viewer）
- 业务系统间资源隔离保障
- 跨业务系统的全览视图数据聚合

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 业务系统 CRUD | 创建、查询、更新、停用、归档业务系统 |
| 成员管理 | 邀请用户、分配角色、移除成员 |
| 权限校验 | 提供权限查询接口，供其他模块鉴权 |
| 资源隔离 | 确保网区、实例、任务等资源按业务系统隔离 |
| 全览视图聚合 | 汇总用户有权限的所有业务系统资源概览数据 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 网区的具体管理 | M1 网区与网段管理 |
| 实例的 CRUD | M3 实例管理 |
| 采集任务调度 | T1 采集任务管理 |
| 告警策略配置 | A1 告警策略 |
| 用户认证（登录/SSO） | I2 用户管理 |
| 通知渠道管理 | A3 通知渠道 |

---

## 三、功能清单

| # | 功能 | 说明 | 优先级 |
|---|------|------|--------|
| F1 | 创建业务系统 | 填写名称、描述，初始化系统标识 | P1 |
| F2 | 查询业务系统列表 | 支持按名称、状态筛选，分页返回 | P1 |
| F3 | 查询业务系统详情 | 返回系统基本信息、成员数、资源统计 | P1 |
| F4 | 更新业务系统 | 修改名称、描述等基本信息 | P1 |
| F5 | 停用业务系统 | 软删除，保留数据但禁止新操作 | P2 |
| F6 | 归档业务系统 | 归档后数据只读，不可恢复 | P3 |
| F7 | 邀请成员 | 将已有用户加入业务系统并分配角色 | P1 |
| F8 | 修改成员角色 | 调整成员在业务系统中的角色 | P1 |
| F9 | 移除成员 | 将用户从业务系统中移除 | P1 |
| F10 | 查询成员列表 | 列出业务系统下所有成员及角色 | P1 |
| F11 | 资源隔离校验 | 查询资源时自动过滤非本业务系统数据 | P1 |
| F12 | 全览视图 | 聚合用户所属全部业务系统的资源概览 | P2 |
| F13 | 业务系统统计 | 网区数、实例数、任务数、告警数汇总 | P2 |

---

## 四、核心数据模型

### 4.1 business_system 表

```sql
CREATE TABLE business_system (
    system_id       VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL COMMENT '业务系统名称',
    description     TEXT COMMENT '业务系统描述',
    status          VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT 'active/suspended/archived',
    config          JSON COMMENT '扩展配置',
    created_by      VARCHAR(64) COMMENT '创建人',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_bs_status ON business_system(status);
CREATE INDEX idx_bs_name ON business_system(name);
```

### 4.2 business_system_member 表

```sql
CREATE TABLE business_system_member (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    system_id       VARCHAR(64) NOT NULL COMMENT '业务系统ID',
    user_id         VARCHAR(64) NOT NULL COMMENT '用户ID',
    role            VARCHAR(32) NOT NULL DEFAULT 'viewer' COMMENT 'admin/operator/viewer',
    joined_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (system_id) REFERENCES business_system(system_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    UNIQUE KEY uk_system_user (system_id, user_id)
);

CREATE INDEX idx_bsm_system ON business_system_member(system_id);
CREATE INDEX idx_bsm_user ON business_system_member(user_id);
```

### 4.3 关联关系

```
business_system (1) ──< (N) business_system_member >── (1) user
       │
       ├──< (N) zone (网区归属)
       │       ├──< (N) network_segment
       │       ├──< (N) instance
       │       ├──< (N) collector_node
       │       └──< (N) data_sink
       │
       └── 资源隔离边界：所有查询自动附加 system_id 过滤条件
```

### 4.4 zone 表扩展字段

为支持业务系统隔离，zone 表需新增 `system_id` 字段：

```sql
ALTER TABLE zone ADD COLUMN system_id VARCHAR(64) NOT NULL COMMENT '所属业务系统ID';
ALTER TABLE zone ADD FOREIGN KEY (system_id) REFERENCES business_system(system_id);
CREATE INDEX idx_zone_system ON zone(system_id);
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| I2 用户管理 | 用户信息查询（邀请成员时校验用户存在性） | 内部 API 调用 |
| Redis 状态缓存 | 全览视图读取实时统计数据 | 缓存读取 |

### 5.2 下游提供

| 消费模块 | 提供内容 | 交互方式 |
|----------|----------|----------|
| M1 网区与网段 | 业务系统下的网区列表、system_id 隔离 | 查询参数传递 |
| M3 实例管理 | 业务系统下的实例列表、权限校验 | 查询参数传递 |
| T1 采集任务 | 业务系统下的任务列表、权限过滤 | 查询参数传递 |
| A1 告警策略 | 业务系统下的告警列表、权限隔离 | 查询参数传递 |
| I3 状态缓存 | 业务系统维度的统计聚合 | 定时任务触发 |

### 5.3 对外 API

#### 5.3.1 创建业务系统

```
POST /api/v1/business-systems
```

**请求体：**

```json
{
  "system_id": "bs-ecommerce",
  "name": "电商平台",
  "description": "核心电商业务系统，覆盖华东/华南网区"
}
```

**响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "system_id": "bs-ecommerce",
    "name": "电商平台",
    "description": "核心电商业务系统，覆盖华东/华南网区",
    "status": "active",
    "created_by": "user-001",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.2 查询业务系统列表

```
GET /api/v1/business-systems
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 按名称/描述模糊搜索 |
| status | string | 否 | 按状态过滤：active/suspended/archived |
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "system_id": "bs-ecommerce",
        "name": "电商平台",
        "status": "active",
        "zone_count": 2,
        "instance_count": 200,
        "member_count": 5,
        "created_at": "2026-10-27T10:00:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询业务系统详情

```
GET /api/v1/business-systems/{system_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "system_id": "bs-ecommerce",
    "name": "电商平台",
    "description": "核心电商业务系统",
    "status": "active",
    "statistics": {
      "zone_count": 2,
      "instance_count": 200,
      "task_count": 600,
      "active_alerts": 3,
      "member_count": 5
    },
    "created_by": "user-001",
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T12:00:00Z"
  }
}
```

#### 5.3.4 更新业务系统

```
PUT /api/v1/business-systems/{system_id}
```

**请求体：**

```json
{
  "name": "电商平台（新）",
  "description": "更新后的描述"
}
```

#### 5.3.5 停用业务系统

```
POST /api/v1/business-systems/{system_id}/suspend
```

**说明：** 停用后，该业务系统下所有资源变为只读状态，不可创建新资源。

#### 5.3.6 归档业务系统

```
POST /api/v1/business-systems/{system_id}/archive
```

**说明：** 归档为终态操作，数据永久保留但不可恢复。归档前需确认系统下无活跃采集任务。

#### 5.3.7 邀请成员

```
POST /api/v1/business-systems/{system_id}/members
```

**请求体：**

```json
{
  "user_id": "user-002",
  "role": "operator"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "system_id": "bs-ecommerce",
    "user_id": "user-002",
    "role": "operator",
    "joined_at": "2026-10-27T10:30:00Z"
  }
}
```

#### 5.3.8 查询成员列表

```
GET /api/v1/business-systems/{system_id}/members
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role | string | 否 | 按角色过滤：admin/operator/viewer |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "user_id": "user-001",
        "username": "alice",
        "display_name": "Alice Wang",
        "role": "admin",
        "joined_at": "2026-10-27T10:00:00Z"
      },
      {
        "id": 2,
        "user_id": "user-002",
        "username": "bob",
        "display_name": "Bob Li",
        "role": "operator",
        "joined_at": "2026-10-27T10:30:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.9 修改成员角色

```
PUT /api/v1/business-systems/{system_id}/members/{member_id}
```

**请求体：**

```json
{
  "role": "admin"
}
```

#### 5.3.10 移除成员

```
DELETE /api/v1/business-systems/{system_id}/members/{member_id}
```

#### 5.3.11 全览视图

```
GET /api/v1/business-systems/overview
```

**说明：** 根据当前用户所属的全部业务系统，聚合返回资源概览。

**响应：**

```json
{
  "code": 0,
  "data": {
    "summary": {
      "total_systems": 2,
      "total_zones": 3,
      "total_instances": 250,
      "total_tasks": 750,
      "active_alerts": 3
    },
    "systems": [
      {
        "system_id": "bs-ecommerce",
        "name": "电商平台",
        "status": "active",
        "zones": [
          {
            "zone_id": "zone-east-1",
            "name": "华东区",
            "instance_count": 120,
            "task_running": 360,
            "active_alerts": 2
          },
          {
            "zone_id": "zone-south-1",
            "name": "华南区",
            "instance_count": 80,
            "task_running": 240,
            "active_alerts": 0
          }
        ]
      },
      {
        "system_id": "bs-payment",
        "name": "支付系统",
        "status": "active",
        "zones": [
          {
            "zone_id": "zone-beijing-1",
            "name": "北京区",
            "instance_count": 50,
            "task_running": 150,
            "active_alerts": 1
          }
        ]
      }
    ]
  }
}
```

#### 5.3.12 查询当前用户的业务系统

```
GET /api/v1/business-systems/my
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "systems": [
      {
        "system_id": "bs-ecommerce",
        "name": "电商平台",
        "role": "admin",
        "status": "active"
      },
      {
        "system_id": "bs-payment",
        "name": "支付系统",
        "role": "viewer",
        "status": "active"
      }
    ]
  }
}
```

---

## 六、设计决策

### 6.1 业务系统作为资源隔离的顶层边界

**决策：** 所有核心资源（zone、instance、task、alert）均通过 `system_id` 归属到业务系统，查询时自动附加隔离条件。

**理由：**
- 多团队/多项目场景下，资源隔离是刚需，物理级隔离比逻辑权限过滤更可靠
- 业务系统作为顶层组织单元，概念清晰，用户易于理解
- 为未来可能的数据分库分表预留扩展空间

### 6.2 角色模型设计

**决策：** 采用三级角色模型（admin / operator / viewer），而非细粒度权限点。

**理由：**
- 社区开源项目追求简洁易用，三级角色覆盖 90% 以上场景
- admin 可管理成员与全部配置，operator 可操作资源但不可管理成员，viewer 只读
- 后续可通过 RBAC 扩展为细粒度权限，不影响当前模型

### 6.3 全览视图的实现策略

**决策：** 全览视图通过 Redis 缓存的预聚合数据构建，而非实时查询数据库。

**理由：**
- 用户可能属于多个业务系统，实时聚合查询代价高
- I3 状态缓存模块已定时聚合各维度统计数据到 Redis
- 全览视图读取 Redis 缓存，响应延迟 < 50ms，满足仪表盘交互需求

### 6.4 停用与归档的两阶段设计

**决策：** 提供停用（suspend）和归档（archive）两种非活跃状态，而非直接删除。

**理由：**
- 停用为临时状态，可随时恢复，适用于业务系统短期下线场景
- 归档为终态操作，数据永久保留但不可恢复，适用于业务系统永久下线
- 软删除策略确保历史数据不丢失，满足审计需求

### 6.5 成员关系独立表设计

**决策：** 使用 `business_system_member` 关联表，而非在 `user` 表中存储 JSON 数组。

**理由：**
- 关联表支持独立查询、索引和分页，性能优于 JSON 字段解析
- 便于统计每个业务系统的成员数量
- 支持未来扩展成员级别的细粒度配置（如通知偏好）
