# M1 网区与网段管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：M1 | 职责：网区全生命周期管理与网段分配，提供 IP 自动推荐与冲突检测能力

---

## 一、概述

网区（Zone）是物理或逻辑的监控部署域，代表一个独立的数据采集环境（如华东区、北京区、DMZ 区）。网段（NetworkSegment）是分配给网区的 IP 范围（CIDR 格式），用于自动关联实例与网区、检测 IP 冲突。

本模块承担以下核心职责：

- 网区的创建、查询、修改与状态管理
- 网段的分配、查询与释放
- 根据实例 IP 自动推荐所属网区
- 跨网区 IP 冲突检测，确保同一 IP 不被多个网区的不同网段覆盖

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 网区 CRUD | 创建、查询、更新、删除网区 |
| 网段管理 | 分配 CIDR 网段到网区，查询网区下的网段列表 |
| IP 自动推荐 | 根据实例 host（IP）匹配网段，推荐所属网区 |
| 冲突检测 | 校验同一 IP 不可被多个网区的不同网段同时覆盖 |
| 网区状态管理 | active / degraded / offline 状态流转 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 业务系统隔离 | B1 业务系统管理 |
| 实例的具体管理 | M3 实例管理 |
| 采集器节点管理 | I1 采集器管理 |
| 数据存储注册 | M5 数据存储管理 |
| 采集任务分配 | T1 采集任务管理 |

---

## 三、功能清单

| # | 功能 | 说明 | 优先级 |
|---|------|------|--------|
| F1 | 创建网区 | 指定名称、采集器类型、remote_write_url，归属业务系统 | P1 |
| F2 | 查询网区列表 | 支持按业务系统、状态、采集器类型筛选 | P1 |
| F3 | 查询网区详情 | 返回网区基本信息、网段列表、资源统计 | P1 |
| F4 | 更新网区 | 修改名称、描述、配置等 | P1 |
| F5 | 删除网区 | 需确保网区下无实例和采集任务 | P1 |
| F6 | 分配网段 | 为网区添加 CIDR 格式网段 | P1 |
| F7 | 查询网段列表 | 按网区查询已分配的所有网段 | P1 |
| F8 | 释放网段 | 从网区移除网段，需确保无实例引用 | P2 |
| F9 | IP 自动推荐 | 输入实例 IP，匹配网段并推荐所属网区 | P1 |
| F10 | 冲突检测 | 新建/修改网段时校验 IP 不跨网区重叠 | P1 |
| F11 | 网区概览统计 | 实例数、任务数、采集器健康状态等 | P2 |

---

## 四、核心数据模型

### 4.1 zone 表

```sql
CREATE TABLE zone (
    zone_id         VARCHAR(64) PRIMARY KEY,
    system_id       VARCHAR(64) NOT NULL COMMENT '所属业务系统ID',
    name            VARCHAR(128) NOT NULL COMMENT '网区名称',
    description     TEXT COMMENT '网区描述',
    collector_type  VARCHAR(32) NOT NULL COMMENT '主采集器: vmagent/prometheus/alloy',
    remote_write_url VARCHAR(512) COMMENT '远程写入地址',
    default_sink_id VARCHAR(64) COMMENT '默认数据存储ID',
    config          JSON COMMENT '网区扩展配置',
    status          VARCHAR(32) DEFAULT 'active' COMMENT 'active/degraded/offline',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (system_id) REFERENCES business_system(system_id),
    FOREIGN KEY (default_sink_id) REFERENCES data_sink(sink_id)
);

CREATE INDEX idx_zone_system ON zone(system_id);
CREATE INDEX idx_zone_status ON zone(status);
```

### 4.2 network_segment 表

```sql
CREATE TABLE network_segment (
    segment_id      BIGINT AUTO_INCREMENT PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '所属网区ID',
    cidr            VARCHAR(45) NOT NULL COMMENT 'CIDR格式: 10.0.1.0/24',
    name            VARCHAR(128) COMMENT '网段名称',
    description     TEXT COMMENT '网段描述',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id) ON DELETE CASCADE,
    UNIQUE KEY uk_zone_cidr (zone_id, cidr)
);

CREATE INDEX idx_segment_zone ON network_segment(zone_id);
CREATE INDEX idx_segment_cidr ON network_segment(cidr);
```

### 4.3 关联关系

```
business_system (1) ──< (N) zone
                          │
                          ├──< (N) network_segment (CIDR 网段)
                          ├──< (N) collector_node
                          ├──< (N) data_sink
                          └──< (N) instance (通过 zone_id 关联)
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| B1 业务系统管理 | system_id 校验、业务系统存在性确认 | 内部 API 调用 |
| M5 数据存储管理 | default_sink_id 校验 | 内部 API 调用 |

### 5.2 下游提供

| 消费模块 | 提供内容 | 交互方式 |
|----------|----------|----------|
| M3 实例管理 | 网区列表、IP 推荐网区、网段匹配 | 查询 API |
| T1 采集任务 | 网区下的采集器节点列表 | 查询 API |
| I1 采集器管理 | 网区与采集器的归属关系 | 数据关联 |
| I3 状态缓存 | 网区维度的统计聚合 | 定时任务触发 |

### 5.3 对外 API

#### 5.3.1 创建网区

```
POST /api/v1/business-systems/{system_id}/zones
```

**请求体：**

```json
{
  "zone_id": "zone-east-1",
  "name": "华东区",
  "description": "华东地区数据采集区",
  "collector_type": "vmagent",
  "remote_write_url": "http://vmagent-east:8429/api/v1/write",
  "config": {
    "timeout": "30s",
    "retry_count": 3
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "zone_id": "zone-east-1",
    "system_id": "bs-ecommerce",
    "name": "华东区",
    "collector_type": "vmagent",
    "status": "active",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.2 查询网区列表

```
GET /api/v1/business-systems/{system_id}/zones
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| collector_type | string | 否 | 按采集器类型过滤 |
| status | string | 否 | 按状态过滤 |
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
        "zone_id": "zone-east-1",
        "name": "华东区",
        "collector_type": "vmagent",
        "segment_count": 3,
        "instance_count": 120,
        "task_count": 360,
        "healthy_nodes": 3,
        "total_nodes": 3,
        "status": "active"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询网区详情

```
GET /api/v1/zones/{zone_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "zone_id": "zone-east-1",
    "system_id": "bs-ecommerce",
    "name": "华东区",
    "description": "华东地区数据采集区",
    "collector_type": "vmagent",
    "remote_write_url": "http://vmagent-east:8429/api/v1/write",
    "default_sink_id": "sink-vm-east",
    "config": { "timeout": "30s", "retry_count": 3 },
    "status": "active",
    "statistics": {
      "segment_count": 3,
      "instance_count": 120,
      "task_count": 360,
      "collector_nodes": 3,
      "data_sinks": 2
    },
    "segments": [
      { "segment_id": 1, "cidr": "10.0.1.0/24", "name": "生产网段A" },
      { "segment_id": 2, "cidr": "10.0.2.0/24", "name": "生产网段B" },
      { "segment_id": 3, "cidr": "10.0.3.0/24", "name": "测试网段" }
    ],
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T12:00:00Z"
  }
}
```

#### 5.3.4 更新网区

```
PUT /api/v1/zones/{zone_id}
```

**请求体：**

```json
{
  "name": "华东区（更新）",
  "description": "更新后的描述",
  "remote_write_url": "http://vmagent-east-v2:8429/api/v1/write",
  "config": { "timeout": "60s" }
}
```

#### 5.3.5 删除网区

```
DELETE /api/v1/zones/{zone_id}
```

**前置校验：**
- 网区下无实例（instance_count = 0）
- 网区下无活跃采集任务（task_running = 0）
- 网区下无采集器节点（collector_node_count = 0）

#### 5.3.6 分配网段

```
POST /api/v1/zones/{zone_id}/segments
```

**请求体：**

```json
{
  "cidr": "10.0.1.0/24",
  "name": "生产网段A",
  "description": "华东区生产环境网段"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "segment_id": 1,
    "zone_id": "zone-east-1",
    "cidr": "10.0.1.0/24",
    "name": "生产网段A",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**校验规则：**
- CIDR 格式合法性校验（IPv4/IPv6）
- 同一网区内 CIDR 不可重复（uk_zone_cidr 唯一约束）
- 跨网区 CIDR 冲突检测（见 F10）

#### 5.3.7 查询网段列表

```
GET /api/v1/zones/{zone_id}/segments
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "segment_id": 1,
        "zone_id": "zone-east-1",
        "cidr": "10.0.1.0/24",
        "name": "生产网段A",
        "instance_count": 45,
        "created_at": "2026-10-27T10:00:00Z"
      }
    ],
    "total": 3
  }
}
```

#### 5.3.8 释放网段

```
DELETE /api/v1/zones/{zone_id}/segments/{segment_id}
```

**前置校验：** 该网段下无实例引用（通过 IP 匹配检查）。

#### 5.3.9 IP 自动推荐

```
POST /api/v1/zones/recommend
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
        "segment_id": 1,
        "cidr": "10.0.1.0/24",
        "segment_name": "生产网段A",
        "confidence": "exact"
      }
    ]
  }
}
```

**匹配逻辑：**
1. 将输入 IP 与所有网段的 CIDR 进行包含关系计算
2. 返回所有匹配的网段及其所属网区
3. confidence 字段标识匹配精度：`exact`（精确匹配）/ `subnet`（子网匹配）

#### 5.3.10 冲突检测

```
POST /api/v1/zones/segments/check-conflict
```

**请求体：**

```json
{
  "cidr": "10.0.1.0/24",
  "exclude_zone_id": "zone-east-1"
}
```

**响应（无冲突）：**

```json
{
  "code": 0,
  "data": {
    "has_conflict": false,
    "conflicts": []
  }
}
```

**响应（有冲突）：**

```json
{
  "code": 0,
  "data": {
    "has_conflict": true,
    "conflicts": [
      {
        "zone_id": "zone-north-1",
        "zone_name": "华北区",
        "segment_id": 5,
        "cidr": "10.0.1.0/24",
        "overlap_type": "exact"
      }
    ]
  }
}
```

**冲突类型：**
- `exact`：CIDR 完全相同
- `overlap`：CIDR 存在 IP 重叠（子网包含关系）

---

## 六、设计决策

### 6.1 CIDR 格式作为网段标准

**决策：** 网段使用标准 CIDR 格式（如 `10.0.1.0/24`），支持 IPv4 和 IPv6。

**理由：**
- CIDR 是网络工程领域的通用标准，运维人员易于理解
- 便于进行 IP 包含关系计算（推荐网区、冲突检测）
- Go/Java/Python 等主流语言均有成熟的 CIDR 解析库

### 6.2 跨网区冲突检测策略

**决策：** 不同业务系统的网区之间允许 IP 重叠，同一业务系统内的网区之间严格禁止。

**理由：**
- 不同业务系统可能管理物理隔离的网络环境，IP 地址空间可能复用
- 同一业务系统内 IP 必须唯一，否则实例归属会产生歧义
- 冲突检测粒度控制在业务系统级别，兼顾灵活性与安全性

### 6.3 IP 推荐算法

**决策：** 采用最长前缀匹配（Longest Prefix Match）策略，当多个网段均包含目标 IP 时，推荐掩码最长（最精确）的网段。

**理由：**
- 与路由器选路算法一致，行为可预期
- 最精确匹配意味着管理粒度最细，推荐结果最合理
- 返回所有匹配结果（而非仅一个），便于用户确认

### 6.4 网区状态模型

**决策：** 网区支持 active / degraded / offline 三态，状态流转由采集器健康检测驱动。

**理由：**
- active：所有采集器节点健康
- degraded：部分采集器节点异常，但仍有节点可用
- offline：所有采集器节点不可达
- 三态模型简洁，满足运维监控需求，避免过度设计

### 6.5 网区与业务系统的归属关系

**决策：** zone 表通过 `system_id` 外键关联业务系统，而非使用独立的关联表。

**理由：**
- 每个网区严格归属一个业务系统，不存在多对多关系
- 外键约束确保数据一致性
- 查询时通过索引快速过滤，性能优于 JOIN 关联表
