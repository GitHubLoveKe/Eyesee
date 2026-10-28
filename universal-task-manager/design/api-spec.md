# API 规范

> 通用采集任务管理平台 RESTful API 规范。平台包含 19 个模块，分为业务层 (B)、模型层 (M)、任务层 (T)、告警层 (A)、基础设施层 (I) 五层。

---

## 一、通用约定

### 1.1 基础路径

```
http(s)://{host}:{port}/api/v1/
```

### 1.2 请求格式

- Content-Type: `application/json`
- 认证方式: Bearer Token（Header: `Authorization: Bearer {token}`）

### 1.3 响应格式

**成功响应：**

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

**错误响应：**

```json
{
  "code": 40001,
  "message": "Invalid parameter: zone_id is required",
  "details": {
    "field": "zone_id",
    "reason": "missing"
  }
}
```

### 1.4 分页

```json
{
  "code": 0,
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

查询参数：`?page=1&page_size=20`

### 1.5 模块总览

| 层级 | 模块编号 | 模块名称 | 主要实体 |
|------|----------|----------|----------|
| B - 业务层 | B1 | 业务系统管理 | BusinessSystem |
| M - 模型层 | M1 | 网区与网段 | Zone, NetworkSegment |
| | M2 | 资源类型树 | ResourceType |
| | M3 | 实例管理 | Instance |
| | M4 | 凭据管理 | Credential |
| | M5 | 数据存储管理 | DataSink |
| T - 任务层 | T1 | 采集任务 | CollectionTask |
| | T2 | RC 任务 | RCTask |
| | T3 | 指标管理 | MetricDefinition |
| | T4 | 服务发现 | ServiceDiscovery |
| | T5 | 定时任务 | ScheduledTask |
| A - 告警层 | A1 | 告警策略 | AlertStrategy |
| | A2 | 告警组 | AlertGroup |
| | A3 | 通知渠道 | NotificationChannel |
| | A4 | 告警收敛 | AlertConvergence |
| I - 基础设施层 | I1 | 采集器与协调器 | CollectorNode, ZoneCoordinator |
| | I2 | 用户管理 | User |
| | I3 | 状态缓存 | StatusCache |

---

## 二、业务系统管理 API (B1)

### 2.1 创建业务系统

```
POST /api/v1/business-systems
```

**请求体：**

```json
{
  "name": "核心交易系统",
  "description": "负责订单、支付等核心业务",
  "owner": "张三",
  "team": "交易团队",
  "members": [
    { "user_id": "user-001", "role": "admin" },
    { "user_id": "user-002", "role": "viewer" }
  ],
  "labels": {
    "department": "技术部",
    "level": "P0"
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "system_id": "bs-001",
    "name": "核心交易系统",
    "owner": "张三",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

### 2.2 查询业务系统列表

```
GET /api/v1/business-systems
```

| 参数 | 类型 | 说明 |
|------|------|------|
| keyword | string | 搜索名称 |
| owner | string | 按负责人过滤 |
| page | int | 页码 |
| page_size | int | 每页数量 |

### 2.3 查询单个业务系统

```
GET /api/v1/business-systems/{system_id}
```

### 2.4 更新业务系统

```
PUT /api/v1/business-systems/{system_id}
```

### 2.5 删除业务系统

```
DELETE /api/v1/business-systems/{system_id}
```

**注意：** 删除前需确保该业务系统下无关联实例和采集任务。

### 2.6 管理成员

```
PUT /api/v1/business-systems/{system_id}/members
```

**请求体：**

```json
{
  "members": [
    { "user_id": "user-001", "role": "admin" },
    { "user_id": "user-003", "role": "editor" }
  ]
}
```

---

## 三、网区与网段管理 API (M1)

### 3.1 创建网区

```
POST /api/v1/zones
```

**请求体：**

```json
{
  "name": "华东区",
  "description": "华东地区监控网区",
  "coordinator_enabled": true,
  "gossip_port": 7946,
  "config": {
    "heartbeat_interval": "5s",
    "gossip_nodes": 3
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "zone_id": "zone-east-1",
    "name": "华东区",
    "description": "华东地区监控网区",
    "coordinator_enabled": true,
    "status": "active",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

### 3.2 查询网区列表

```
GET /api/v1/zones
```

| 参数 | 类型 | 说明 |
|------|------|------|
| status | string | 按状态过滤 (active / inactive) |
| page | int | 页码 |
| page_size | int | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "zone_id": "zone-east-1",
        "name": "华东区",
        "instance_count": 42,
        "collector_count": 8,
        "coordinator_status": "leader",
        "status": "active"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

### 3.3 查询单个网区

```
GET /api/v1/zones/{zone_id}
```

### 3.4 更新网区

```
PUT /api/v1/zones/{zone_id}
```

### 3.5 删除网区

```
DELETE /api/v1/zones/{zone_id}
```

**注意：** 删除前需确保该网区下无实例和采集任务。

### 3.6 网段管理

#### 创建网段

```
POST /api/v1/zones/{zone_id}/network-segments
```

**请求体：**

```json
{
  "name": "生产网段 A",
  "cidr": "10.0.1.0/24",
  "vlan_id": 100,
  "gateway": "10.0.1.1",
  "description": "生产环境 A 段"
}
```

#### 查询网段列表

```
GET /api/v1/zones/{zone_id}/network-segments
```

#### 更新网段

```
PUT /api/v1/zones/{zone_id}/network-segments/{segment_id}
```

#### 删除网段

```
DELETE /api/v1/zones/{zone_id}/network-segments/{segment_id}
```

---

## 四、资源类型树 API (M2)

资源类型以树形结构组织，支持多级分类。例如：`数据库 > 关系型 > MySQL`。

### 4.1 创建资源类型

```
POST /api/v1/resource-types
```

**请求体：**

```json
{
  "name": "MySQL",
  "code": "mysql",
  "parent_id": "type-rdbms",
  "icon": "mysql",
  "config_schema": {
    "fields": [
      { "key": "port", "type": "integer", "default": 3306 },
      { "key": "database", "type": "string", "required": false }
    ]
  },
  "supported_collectors": ["vmagent", "alloy"]
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "type_id": "type-mysql",
    "name": "MySQL",
    "code": "mysql",
    "parent_id": "type-rdbms",
    "full_path": "数据库/关系型/MySQL",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

### 4.2 查询资源类型树

```
GET /api/v1/resource-types/tree
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "tree": [
      {
        "type_id": "type-database",
        "name": "数据库",
        "code": "database",
        "children": [
          {
            "type_id": "type-rdbms",
            "name": "关系型",
            "code": "rdbms",
            "children": [
              { "type_id": "type-mysql", "name": "MySQL", "code": "mysql", "children": [] },
              { "type_id": "type-postgresql", "name": "PostgreSQL", "code": "postgresql", "children": [] }
            ]
          }
        ]
      }
    ]
  }
}
```

### 4.3 查询资源类型列表（扁平）

```
GET /api/v1/resource-types
```

| 参数 | 类型 | 说明 |
|------|------|------|
| parent_id | string | 按父类型过滤 |
| code | string | 按编码精确查询 |
| keyword | string | 搜索名称 |

### 4.4 查询单个资源类型

```
GET /api/v1/resource-types/{type_id}
```

### 4.5 更新资源类型

```
PUT /api/v1/resource-types/{type_id}
```

### 4.6 删除资源类型

```
DELETE /api/v1/resource-types/{type_id}
```

**注意：** 删除前需确保无子类型且无实例引用。

---

## 五、实例管理 API (M3)

实例（Instance）是被监控的目标对象，如一台 MySQL 数据库、一个 Redis 节点等。

### 5.1 创建实例

```
POST /api/v1/zones/{zone_id}/instances
```

**请求体：**

```json
{
  "name": "MySQL 生产主库",
  "resource_type": "mysql",
  "host": "10.0.1.5",
  "port": 3306,
  "credential_id": "cred-mysql-001",
  "network_segment_id": "seg-001",
  "business_system_id": "bs-001",
  "labels": {
    "env": "prod",
    "role": "primary"
  },
  "extra": {
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
    "instance_id": "inst-001",
    "zone_id": "zone-east-1",
    "name": "MySQL 生产主库",
    "resource_type": "mysql",
    "host": "10.0.1.5",
    "port": 3306,
    "status": "unknown",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

### 5.2 查询实例列表

```
GET /api/v1/zones/{zone_id}/instances
```

| 参数 | 类型 | 说明 |
|------|------|------|
| resource_type | string | 按资源类型过滤 |
| status | string | 按状态过滤 (healthy / unhealthy / unknown) |
| label | string | 按标签过滤，格式: `key=value` |
| business_system_id | string | 按业务系统过滤 |
| keyword | string | 搜索名称或主机 |
| page | int | 页码 |
| page_size | int | 每页数量 |

### 5.3 查询单个实例

```
GET /api/v1/instances/{instance_id}
```

### 5.4 更新实例

```
PUT /api/v1/instances/{instance_id}
```

### 5.5 删除实例

```
DELETE /api/v1/instances/{instance_id}
```

**注意：** 删除前需确保无关联的采集任务。

### 5.6 连通性测试

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

### 5.7 查询测试结果

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

---

## 六、凭据管理 API (M4)

凭据用于存储实例的连接认证信息（用户名/密码、SSH 密钥、Token 等），与实例解耦管理。

### 6.1 创建凭据

```
POST /api/v1/credentials
```

**请求体：**

```json
{
  "name": "MySQL 生产账号",
  "credential_type": "password",
  "zone_id": "zone-east-1",
  "data": {
    "username": "monitor",
    "password": "encrypted:AES256:..."
  },
  "expires_at": "2027-10-27T00:00:00Z"
}
```

> `credential_type` 可选值：`password` | `ssh_key` | `token` | `tls_cert`

**响应：**

```json
{
  "code": 0,
  "data": {
    "credential_id": "cred-001",
    "name": "MySQL 生产账号",
    "credential_type": "password",
    "zone_id": "zone-east-1",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

> 注意：响应中不返回凭据明文数据。

### 6.2 查询凭据列表

```
GET /api/v1/credentials
```

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| credential_type | string | 按凭据类型过滤 |
| keyword | string | 搜索名称 |

### 6.3 查询单个凭据

```
GET /api/v1/credentials/{credential_id}
```

### 6.4 更新凭据

```
PUT /api/v1/credentials/{credential_id}
```

### 6.5 删除凭据

```
DELETE /api/v1/credentials/{credential_id}
```

**注意：** 删除前需确保无实例引用该凭据。

---

## 七、数据存储管理 API (M5)

DataSink 抽象了监控数据的存储后端，支持多种存储类型。

### 7.1 创建数据存储

```
POST /api/v1/data-sinks
```

**请求体：**

```json
{
  "name": "华东 VM 集群",
  "sink_type": "victoriametrics",
  "zone_id": "zone-east-1",
  "config": {
    "url": "http://vmstorage-east:8428",
    "write_concurrency": 4,
    "retention_days": 30
  }
}
```

> `sink_type` 可选值：`victoriametrics` | `elasticsearch` | `loki` | `openobserve` | `s3`

**响应：**

```json
{
  "code": 0,
  "data": {
    "sink_id": "sink-vm-001",
    "name": "华东 VM 集群",
    "sink_type": "victoriametrics",
    "zone_id": "zone-east-1",
    "status": "active",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

### 7.2 查询数据存储列表

```
GET /api/v1/data-sinks
```

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| sink_type | string | 按存储类型过滤 |
| status | string | 按状态过滤 |

### 7.3 查询单个数据存储

```
GET /api/v1/data-sinks/{sink_id}
```

### 7.4 更新数据存储

```
PUT /api/v1/data-sinks/{sink_id}
```

### 7.5 删除数据存储

```
DELETE /api/v1/data-sinks/{sink_id}
```

### 7.6 连通性检测

```
POST /api/v1/data-sinks/{sink_id}/health
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "status": "healthy",
    "latency_ms": 12,
    "checked_at": "2026-10-27T10:10:00Z"
  }
}
```

---

## 八、采集任务管理 API (T1)

### 8.1 创建采集任务

```
POST /api/v1/instances/{instance_id}/collection-tasks
```

**请求体：**

```json
{
  "name": "MySQL 生产采集",
  "resource_type": "mysql",
  "config": {
    "scrape_interval": "15s",
    "enable_slow_query": false,
    "databases": ["mysql", "performance_schema"]
  },
  "sink_id": "sink-vm-001",
  "enabled": true
}
```

> **重要：** `node_id` 为只读字段，由 Zone Coordinator 通过一致性哈希自动分配，创建时不可指定。

**响应：**

```json
{
  "code": 0,
  "data": {
    "task_id": "ct-001",
    "instance_id": "inst-001",
    "zone_id": "zone-east-1",
    "name": "MySQL 生产采集",
    "resource_type": "mysql",
    "node_id": null,
    "enabled": true,
    "status": "pending",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

> `node_id` 初始为 `null`，Zone Coordinator 完成分配后自动填充。

### 8.2 查询实例的采集任务列表

```
GET /api/v1/instances/{instance_id}/collection-tasks
```

### 8.3 查询采集任务列表（全局）

```
GET /api/v1/collection-tasks
```

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| instance_id | string | 按实例过滤 |
| resource_type | string | 按资源类型过滤 |
| node_id | string | 按采集器节点过滤 |
| status | string | 按状态过滤 |
| enabled | boolean | 按启用状态过滤 |
| keyword | string | 搜索名称 |

### 8.4 查询单个采集任务

```
GET /api/v1/collection-tasks/{task_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "task_id": "ct-001",
    "instance_id": "inst-001",
    "zone_id": "zone-east-1",
    "node_id": "collector-03",
    "name": "MySQL 生产采集",
    "resource_type": "mysql",
    "config": { "scrape_interval": "15s" },
    "sink_id": "sink-vm-001",
    "enabled": true,
    "status": "running",
    "last_scrape_at": "2026-10-27T10:30:15Z",
    "last_error": null,
    "created_at": "2026-10-27T09:00:00Z"
  }
}
```

### 8.5 更新采集任务

```
PUT /api/v1/collection-tasks/{task_id}
```

> `node_id` 不可通过此接口修改。

### 8.6 删除采集任务

```
DELETE /api/v1/collection-tasks/{task_id}
```

### 8.7 启用 / 禁用任务

```
POST /api/v1/collection-tasks/{task_id}/enable
POST /api/v1/collection-tasks/{task_id}/disable
```

### 8.8 查询任务状态

```
GET /api/v1/collection-tasks/{task_id}/status
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "task_id": "ct-001",
    "status": "running",
    "node_id": "collector-03",
    "zone_id": "zone-east-1",
    "target": "10.0.1.5:9104",
    "scrape_count_24h": 5760,
    "error_count_24h": 0,
    "last_success_at": "2026-10-27T10:30:15Z",
    "avg_scrape_duration_ms": 125
  }
}
```

---

## 九、RC 任务管理 API (T2)

RC（Rule Collection）任务用于基于指标数据的规则计算与评估。

### 9.1 创建 RC 任务

```
POST /api/v1/rc-tasks
```

**请求体：**

```json
{
  "name": "MySQL 连接数告警规则",
  "zone_id": "zone-east-1",
  "sink_id": "sink-vm-001",
  "rules": [
    {
      "rule_name": "high_connections",
      "expr": "mysql_global_status_connections > 1000",
      "severity": "warning",
      "duration": "5m",
      "labels": { "team": "dba" },
      "annotations": {
        "summary": "MySQL 连接数过高",
        "description": "实例 {{ $labels.instance }} 连接数 {{ $value }}"
      }
    }
  ],
  "eval_interval": "30s",
  "enabled": true
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "rc_task_id": "rc-001",
    "name": "MySQL 连接数告警规则",
    "zone_id": "zone-east-1",
    "rule_count": 1,
    "eval_interval": "30s",
    "enabled": true,
    "status": "running",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

### 9.2 查询 RC 任务列表

```
GET /api/v1/rc-tasks
```

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| enabled | boolean | 按启用状态过滤 |
| keyword | string | 搜索名称 |

### 9.3 查询单个 RC 任务

```
GET /api/v1/rc-tasks/{rc_task_id}
```

### 9.4 更新 RC 任务

```
PUT /api/v1/rc-tasks/{rc_task_id}
```

### 9.5 删除 RC 任务

```
DELETE /api/v1/rc-tasks/{rc_task_id}
```

### 9.6 规则管理

#### 添加规则

```
POST /api/v1/rc-tasks/{rc_task_id}/rules
```

**请求体：**

```json
{
  "rule_name": "slow_queries",
  "expr": "rate(mysql_slow_queries_total[5m]) > 10",
  "severity": "critical",
  "duration": "2m",
  "labels": {},
  "annotations": {
    "summary": "慢查询过多"
  }
}
```

#### 查询规则列表

```
GET /api/v1/rc-tasks/{rc_task_id}/rules
```

#### 更新规则

```
PUT /api/v1/rc-tasks/{rc_task_id}/rules/{rule_id}
```

#### 删除规则

```
DELETE /api/v1/rc-tasks/{rc_task_id}/rules/{rule_id}
```

---

## 十、指标管理 API (T3)

管理采集任务产出的指标定义，包括指标名称、类型、标签等元数据。

### 10.1 查询指标定义列表

```
GET /api/v1/metric-definitions
```

| 参数 | 类型 | 说明 |
|------|------|------|
| resource_type | string | 按资源类型过滤 |
| task_id | string | 按采集任务过滤 |
| metric_type | string | 按指标类型过滤 (gauge / counter / histogram / summary) |
| keyword | string | 搜索指标名称 |

### 10.2 查询单个指标定义

```
GET /api/v1/metric-definitions/{metric_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "metric_id": "metric-001",
    "name": "mysql_global_status_connections",
    "metric_type": "gauge",
    "help": "当前连接数",
    "resource_type": "mysql",
    "labels": ["instance", "zone"],
    "task_id": "ct-001"
  }
}
```

### 10.3 创建 / 更新指标定义

```
POST /api/v1/metric-definitions
PUT /api/v1/metric-definitions/{metric_id}
```

### 10.4 删除指标定义

```
DELETE /api/v1/metric-definitions/{metric_id}
```

---

## 十一、服务发现 API (T4)

为采集器提供 HTTP-based 服务发现（http_sd），兼容 Prometheus http_sd 协议。

### 11.1 http_sd 接口

```
GET /api/v1/sd/{zone_id}/targets
```

| 参数 | 类型 | 说明 |
|------|------|------|
| resource_type | string | 按资源类型过滤 |
| node_id | string | 按采集器节点过滤 |

**响应（Prometheus http_sd 格式）：**

```json
[
  {
    "targets": ["10.0.1.5:9104"],
    "labels": {
      "job": "mysql-inst-001",
      "instance": "inst-001",
      "resource_type": "mysql",
      "__meta_task_id": "ct-001",
      "__meta_zone_id": "zone-east-1",
      "__meta_node_id": "collector-03"
    }
  },
  {
    "targets": ["10.0.1.6:9104"],
    "labels": {
      "job": "mysql-inst-002",
      "instance": "inst-002",
      "resource_type": "mysql",
      "__meta_task_id": "ct-002",
      "__meta_zone_id": "zone-east-1",
      "__meta_node_id": "collector-03"
    }
  }
]
```

> 该接口由 Zone Coordinator 提供，采集器通过 HTTP 轮询获取分配给本节点的 targets。

---

## 十二、定时任务 API (T5)

管理周期性执行的调度任务，如数据清理、报表生成、同步等。

### 12.1 创建定时任务

```
POST /api/v1/scheduled-tasks
```

**请求体：**

```json
{
  "name": "每日指标归档",
  "task_type": "metric_archive",
  "zone_id": "zone-east-1",
  "cron": "0 2 * * *",
  "config": {
    "source_sink_id": "sink-vm-001",
    "target_sink_id": "sink-s3-001",
    "retention_days": 90
  },
  "enabled": true
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "scheduled_task_id": "st-001",
    "name": "每日指标归档",
    "task_type": "metric_archive",
    "cron": "0 2 * * *",
    "next_run_at": "2026-10-28T02:00:00Z",
    "enabled": true,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

### 12.2 查询定时任务列表

```
GET /api/v1/scheduled-tasks
```

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| task_type | string | 按任务类型过滤 |
| enabled | boolean | 按启用状态过滤 |

### 12.3 查询单个定时任务

```
GET /api/v1/scheduled-tasks/{scheduled_task_id}
```

### 12.4 更新定时任务

```
PUT /api/v1/scheduled-tasks/{scheduled_task_id}
```

### 12.5 删除定时任务

```
DELETE /api/v1/scheduled-tasks/{scheduled_task_id}
```

### 12.6 启用 / 禁用

```
POST /api/v1/scheduled-tasks/{scheduled_task_id}/enable
POST /api/v1/scheduled-tasks/{scheduled_task_id}/disable
```

### 12.7 手动触发执行

```
POST /api/v1/scheduled-tasks/{scheduled_task_id}/trigger
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "execution_id": "exec-001",
    "status": "running",
    "started_at": "2026-10-27T10:15:00Z"
  }
}
```

### 12.8 查询执行历史

```
GET /api/v1/scheduled-tasks/{scheduled_task_id}/executions
```

---

## 十三、告警策略 API (A1)

### 13.1 创建告警策略

```
POST /api/v1/alert-strategies
```

**请求体：**

```json
{
  "name": "MySQL 高连接数",
  "zone_id": "zone-east-1",
  "rc_task_id": "rc-001",
  "rule_name": "high_connections",
  "severity": "warning",
  "alert_group_id": "ag-dba",
  "notification_channel_id": "nc-dingtalk",
  "inhibit_rules": [
    { "source_match": { "severity": "critical" }, "target_match": { "severity": "warning" } }
  ],
  "enabled": true
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "strategy_id": "as-001",
    "name": "MySQL 高连接数",
    "severity": "warning",
    "enabled": true,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

### 13.2 查询告警策略列表

```
GET /api/v1/alert-strategies
```

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| severity | string | 按严重级别过滤 (critical / warning / info) |
| enabled | boolean | 按启用状态过滤 |
| keyword | string | 搜索名称 |

### 13.3 查询单个告警策略

```
GET /api/v1/alert-strategies/{strategy_id}
```

### 13.4 更新告警策略

```
PUT /api/v1/alert-strategies/{strategy_id}
```

### 13.5 删除告警策略

```
DELETE /api/v1/alert-strategies/{strategy_id}
```

---

## 十四、告警组 API (A2)

告警组定义了一组告警接收人，用于告警路由。

### 14.1 创建告警组

```
POST /api/v1/alert-groups
```

**请求体：**

```json
{
  "name": "DBA 值班组",
  "description": "数据库管理员值班组",
  "members": [
    { "user_id": "user-001", "role": "receiver" },
    { "user_id": "user-002", "role": "escalation" }
  ],
  "routing": {
    "group_by": ["alertname", "zone_id"],
    "group_wait": "30s",
    "group_interval": "5m",
    "repeat_interval": "4h"
  }
}
```

### 14.2 查询告警组列表

```
GET /api/v1/alert-groups
```

### 14.3 查询单个告警组

```
GET /api/v1/alert-groups/{alert_group_id}
```

### 14.4 更新告警组

```
PUT /api/v1/alert-groups/{alert_group_id}
```

### 14.5 删除告警组

```
DELETE /api/v1/alert-groups/{alert_group_id}
```

---

## 十五、通知渠道 API (A3)

### 15.1 创建通知渠道

```
POST /api/v1/notification-channels
```

**请求体：**

```json
{
  "name": "钉钉告警群",
  "channel_type": "dingtalk",
  "config": {
    "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    "secret": "encrypted:AES256:...",
    "at_mobiles": ["13800138000"],
    "msg_type": "markdown"
  }
}
```

> `channel_type` 可选值：`dingtalk` | `wechat` | `email` | `slack` | `webhook` | `pagerduty` | `opsgenie`

### 15.2 查询通知渠道列表

```
GET /api/v1/notification-channels
```

| 参数 | 类型 | 说明 |
|------|------|------|
| channel_type | string | 按渠道类型过滤 |
| keyword | string | 搜索名称 |

### 15.3 查询单个通知渠道

```
GET /api/v1/notification-channels/{channel_id}
```

### 15.4 更新通知渠道

```
PUT /api/v1/notification-channels/{channel_id}
```

### 15.5 删除通知渠道

```
DELETE /api/v1/notification-channels/{channel_id}
```

### 15.6 发送测试通知

```
POST /api/v1/notification-channels/{channel_id}/test
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "status": "sent",
    "sent_at": "2026-10-27T10:20:00Z"
  }
}
```

---

## 十六、告警收敛 API (A4)

告警收敛模块提供告警查询、确认、静默等操作。平台采用两级 Alertmanager 架构：Zone AM（去重） -> Core AM（路由） -> Kafka。

### 16.1 查询告警列表

```
GET /api/v1/alerts
```

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| severity | string | 按严重级别过滤 |
| status | string | 按状态过滤 (firing / resolved / silenced) |
| alertname | string | 按告警名称过滤 |
| instance_id | string | 按实例过滤 |
| start_time | string | 起始时间 (RFC3339) |
| end_time | string | 结束时间 (RFC3339) |
| page | int | 页码 |
| page_size | int | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "alert_id": "alert-001",
        "alertname": "high_connections",
        "severity": "warning",
        "status": "firing",
        "zone_id": "zone-east-1",
        "instance_id": "inst-001",
        "labels": { "team": "dba" },
        "annotations": {
          "summary": "MySQL 连接数过高",
          "description": "实例 inst-001 连接数 1523"
        },
        "starts_at": "2026-10-27T10:00:00Z",
        "ends_at": null,
        "generator_url": "http://rc-001/rules/high_connections"
      }
    ],
    "total": 15,
    "page": 1,
    "page_size": 20
  }
}
```

### 16.2 查询单个告警详情

```
GET /api/v1/alerts/{alert_id}
```

### 16.3 确认告警

```
POST /api/v1/alerts/{alert_id}/acknowledge
```

**请求体：**

```json
{
  "user_id": "user-001",
  "comment": "已确认，正在处理"
}
```

### 16.4 批量确认告警

```
POST /api/v1/alerts/batch-acknowledge
```

**请求体：**

```json
{
  "alert_ids": ["alert-001", "alert-002"],
  "user_id": "user-001",
  "comment": "批量确认"
}
```

### 16.5 创建静默

```
POST /api/v1/silences
```

**请求体：**

```json
{
  "matchers": [
    { "name": "alertname", "value": "high_connections", "is_regex": false },
    { "name": "zone_id", "value": "zone-east-1", "is_regex": false }
  ],
  "starts_at": "2026-10-27T10:00:00Z",
  "ends_at": "2026-10-27T14:00:00Z",
  "created_by": "user-001",
  "comment": "维护窗口，临时静默"
}
```

### 16.6 查询静默列表

```
GET /api/v1/silences
```

| 参数 | 类型 | 说明 |
|------|------|------|
| state | string | 按状态过滤 (active / pending / expired) |
| created_by | string | 按创建人过滤 |

### 16.7 删除静默（解除静默）

```
DELETE /api/v1/silences/{silence_id}
```

---

## 十七、采集器与协调器 API (I1)

### 17.1 采集器节点管理

#### 注册采集器节点

```
POST /api/v1/collector-nodes
```

**请求体：**

```json
{
  "name": "collector-east-03",
  "zone_id": "zone-east-1",
  "node_type": "vmagent",
  "address": "10.0.2.3:8429",
  "labels": {
    "rack": "A3",
    "capacity": "high"
  }
}
```

> `node_type` 可选值：`vmagent` | `alloy` | `prometheus` | `custom`

**响应：**

```json
{
  "code": 0,
  "data": {
    "node_id": "collector-03",
    "name": "collector-east-03",
    "zone_id": "zone-east-1",
    "node_type": "vmagent",
    "status": "online",
    "task_count": 0,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 查询采集器节点列表

```
GET /api/v1/collector-nodes
```

| 参数 | 类型 | 说明 |
|------|------|------|
| zone_id | string | 按网区过滤 |
| node_type | string | 按采集器类型过滤 |
| status | string | 按状态过滤 (online / offline / degraded) |

#### 查询单个采集器节点

```
GET /api/v1/collector-nodes/{node_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "node_id": "collector-03",
    "name": "collector-east-03",
    "zone_id": "zone-east-1",
    "node_type": "vmagent",
    "address": "10.0.2.3:8429",
    "status": "online",
    "task_count": 15,
    "cpu_usage": 0.35,
    "memory_usage": 0.62,
    "last_heartbeat_at": "2026-10-27T10:30:00Z",
    "created_at": "2026-10-27T08:00:00Z"
  }
}
```

#### 更新采集器节点

```
PUT /api/v1/collector-nodes/{node_id}
```

#### 注销采集器节点

```
DELETE /api/v1/collector-nodes/{node_id}
```

**注意：** 注销前需确保节点上无运行中的采集任务。

#### 查询节点上的采集任务

```
GET /api/v1/collector-nodes/{node_id}/tasks
```

### 17.2 Zone Coordinator 管理

Zone Coordinator 在每个网区运行，通过 gossip 协议维护集群成员关系，使用一致性哈希将采集任务分配到采集器节点。

#### 查询协调器状态

```
GET /api/v1/zones/{zone_id}/coordinator
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "zone_id": "zone-east-1",
    "leader_node": "collector-01",
    "members": [
      { "node_id": "collector-01", "address": "10.0.2.1:7946", "state": "alive" },
      { "node_id": "collector-02", "address": "10.0.2.2:7946", "state": "alive" },
      { "node_id": "collector-03", "address": "10.0.2.3:7946", "state": "alive" }
    ],
    "total_tasks": 45,
    "hash_ring_size": 128,
    "last_rebalance_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 手动触发任务重平衡

```
POST /api/v1/zones/{zone_id}/coordinator/rebalance
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "status": "rebalancing",
    "tasks_affected": 5,
    "started_at": "2026-10-27T10:35:00Z"
  }
}
```

#### 查询任务分配视图

```
GET /api/v1/zones/{zone_id}/coordinator/assignments
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "assignments": [
      {
        "node_id": "collector-01",
        "task_ids": ["ct-001", "ct-002", "ct-003"],
        "hash_slots": [0, 1, 2, 30, 31]
      },
      {
        "node_id": "collector-02",
        "task_ids": ["ct-004", "ct-005"],
        "hash_slots": [42, 43, 44, 45]
      }
    ]
  }
}
```

---

## 十八、用户管理 API (I2)

### 18.1 创建用户

```
POST /api/v1/users
```

**请求体：**

```json
{
  "username": "zhangsan",
  "display_name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "role": "editor",
  "business_system_ids": ["bs-001"]
}
```

> `role` 可选值：`admin` | `editor` | `viewer`

### 18.2 查询用户列表

```
GET /api/v1/users
```

| 参数 | 类型 | 说明 |
|------|------|------|
| role | string | 按角色过滤 |
| keyword | string | 搜索用户名或姓名 |

### 18.3 查询单个用户

```
GET /api/v1/users/{user_id}
```

### 18.4 更新用户

```
PUT /api/v1/users/{user_id}
```

### 18.5 删除用户

```
DELETE /api/v1/users/{user_id}
```

### 18.6 重置密码

```
POST /api/v1/users/{user_id}/reset-password
```

### 18.7 用户登录

```
POST /api/v1/auth/login
```

**请求体：**

```json
{
  "username": "zhangsan",
  "password": "***"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-10-28T10:00:00Z",
    "user": {
      "user_id": "user-001",
      "username": "zhangsan",
      "role": "editor"
    }
  }
}
```

---

## 十九、状态缓存 API (I3)

状态缓存提供实例和任务实时状态的快速查询能力，数据由采集器上报并缓存在内存中。

### 19.1 查询实例状态

```
GET /api/v1/status/instances/{instance_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "instance_id": "inst-001",
    "status": "healthy",
    "uptime_seconds": 864000,
    "last_check_at": "2026-10-27T10:30:00Z",
    "metrics_summary": {
      "total_metrics": 42,
      "last_scrape_duration_ms": 125
    }
  }
}
```

### 19.2 查询网区状态概览

```
GET /api/v1/status/zones/{zone_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "zone_id": "zone-east-1",
    "total_instances": 42,
    "healthy_count": 38,
    "unhealthy_count": 2,
    "unknown_count": 2,
    "total_tasks": 128,
    "running_tasks": 120,
    "failed_tasks": 3,
    "pending_tasks": 5,
    "collector_nodes": {
      "total": 8,
      "online": 7,
      "offline": 1
    },
    "updated_at": "2026-10-27T10:30:00Z"
  }
}
```

### 19.3 批量查询实例状态

```
POST /api/v1/status/instances/batch
```

**请求体：**

```json
{
  "instance_ids": ["inst-001", "inst-002", "inst-003"]
}
```

### 19.4 刷新状态缓存

```
POST /api/v1/status/refresh
```

**请求体：**

```json
{
  "zone_id": "zone-east-1",
  "scope": "all"
}
```

> `scope` 可选值：`all` | `instances` | `tasks`

---

## 二十、批量操作 API

### 20.1 批量导入实例

```
POST /api/v1/zones/{zone_id}/instances/import
```

**请求体：**

```json
{
  "format": "csv",
  "data": "name,resource_type,host,port\nMySQL-01,mysql,10.0.1.5,3306\nRedis-01,redis,10.0.1.6,6379\n..."
}
```

> `format` 可选值：`csv` | `json`

### 20.2 批量导出实例

```
GET /api/v1/zones/{zone_id}/instances/export?format=csv
```

### 20.3 批量启用 / 禁用采集任务

```
POST /api/v1/collection-tasks/batch
```

**请求体：**

```json
{
  "action": "enable",
  "task_ids": ["ct-001", "ct-002", "ct-003"]
}
```

> `action` 可选值：`enable` | `disable`

### 20.4 批量删除

```
POST /api/v1/batch/delete
```

**请求体：**

```json
{
  "resource_type": "collection_task",
  "ids": ["ct-001", "ct-002"]
}
```

---

## 二十一、错误码

| 错误码 | 说明 |
|--------|------|
| 40001 | 参数缺失或格式错误 |
| 40002 | 资源不存在 |
| 40003 | 资源状态冲突 |
| 40101 | 认证失败或 Token 过期 |
| 40301 | 权限不足 |
| 40901 | 资源冲突（重复创建） |
| 42201 | 业务规则校验失败 |
| 42202 | 存在关联资源，无法删除 |
| 42203 | 凭据已过期 |
| 42204 | 网区协调器未就绪 |
| 50001 | 内部错误 |
| 50301 | 服务不可用 |
| 50302 | Zone Coordinator 不可达 |
