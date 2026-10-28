# M5 数据存储管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：M5 | 职责：多类型数据存储注册、协议适配、写入路由与健康检查

---

## 一、概述

数据存储（DataSink）是网区内接收采集数据的后端系统。平台支持多种存储类型（VictoriaMetrics、Elasticsearch、Loki、OpenObserve、S3/MinIO），每种存储通过统一的 DataSinkAdapter 接口适配，支持 metric / log / trace 多协议数据写入。

每个网区可注册多个数据存储，其中一个是默认存储。采集任务可指定目标存储，未指定时使用网区默认存储。

本模块承担以下核心职责：

- 多类型数据存储的注册、查询、更新与删除
- DataSinkAdapter 统一接口抽象
- 网区默认存储管理
- 数据写入路由逻辑
- 存储健康检查

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 存储注册 | 注册 DataSink 实例，配置连接参数 |
| 协议适配 | 通过 DataSinkAdapter 接口适配不同存储后端 |
| 默认存储 | 管理网区的默认数据存储 |
| 写入路由 | 根据采集任务配置决定数据写入目标 |
| 健康检查 | 定期检查存储后端的可用性 |
| 存储类型扩展 | 支持通过实现接口注册新的存储类型 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 网区管理 | M1 网区与网段管理 |
| 采集任务配置 | T1 采集任务管理 |
| 采集器适配器（生成配置） | 采集器适配器设计 |
| 数据查询与可视化 | 外部系统（Grafana / Kibana） |
| 数据归档与清理策略 | T5 定时任务 |

---

## 三、功能清单

| # | 功能 | 说明 | 优先级 |
|---|------|------|--------|
| F1 | 注册数据存储 | 指定类型、连接参数、协议，归属网区 | P1 |
| F2 | 查询存储列表 | 按网区、类型、协议过滤 | P1 |
| F3 | 查询存储详情 | 返回配置信息、健康状态、使用统计 | P1 |
| F4 | 更新存储配置 | 修改连接参数、协议等 | P1 |
| F5 | 删除存储 | 需确保无采集任务引用且非默认存储 | P1 |
| F6 | 设置默认存储 | 指定网区的默认数据存储 | P1 |
| F7 | 连接校验 | 注册时验证存储后端连通性 | P1 |
| F8 | 健康检查 | 定期检查存储后端健康状态 | P2 |
| F9 | 写入路由 | 根据采集任务配置决定数据写入目标 | P1 |
| F10 | 多协议支持 | metric / log / trace 多协议数据写入 | P2 |
| F11 | 存储使用统计 | 数据量、写入速率、查询延迟等 | P3 |

---

## 四、核心数据模型

### 4.1 data_sink 表

```sql
CREATE TABLE data_sink (
    sink_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '所属网区ID',
    name            VARCHAR(128) NOT NULL COMMENT '存储名称',
    sink_type       VARCHAR(32) NOT NULL COMMENT 'victoriametrics/elasticsearch/loki/openobserve/s3/custom',
    config          JSON NOT NULL COMMENT '连接配置(JSON)',
    is_default      BOOLEAN DEFAULT FALSE COMMENT '是否为网区默认存储',
    protocol        VARCHAR(32) NOT NULL DEFAULT 'metric' COMMENT '支持协议: metric/log/trace/multi',
    status          VARCHAR(32) DEFAULT 'active' COMMENT 'active/degraded/offline',
    description     TEXT COMMENT '存储描述',
    last_health_at  TIMESTAMP NULL COMMENT '最近健康检查时间',
    last_health_ok  BOOLEAN NULL COMMENT '最近健康检查是否通过',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_sink_zone ON data_sink(zone_id);
CREATE INDEX idx_sink_type ON data_sink(sink_type);
CREATE INDEX idx_sink_default ON data_sink(zone_id, is_default);
CREATE INDEX idx_sink_status ON data_sink(status);
```

### 4.2 存储类型与配置结构

#### VictoriaMetrics

```json
{
  "sink_type": "victoriametrics",
  "protocol": "metric",
  "config": {
    "url": "http://vm-prod:8428",
    "remote_write_url": "http://vm-prod:8428/api/v1/write",
    "remote_read_url": "http://vm-prod:8428/api/v1/query",
    "username": "",
    "password": "",
    "bearer_token": "",
    "timeout": "30s",
    "tls_config": {
      "insecure_skip_verify": false
    }
  }
}
```

#### Elasticsearch

```json
{
  "sink_type": "elasticsearch",
  "protocol": "log",
  "config": {
    "urls": ["http://es-node1:9200", "http://es-node2:9200"],
    "index_prefix": "monitor-logs",
    "username": "elastic",
    "password": "***",
    "api_key": "",
    "timeout": "30s",
    "bulk_size": 1000,
    "tls_config": {
      "insecure_skip_verify": false
    }
  }
}
```

#### Loki

```json
{
  "sink_type": "loki",
  "protocol": "log",
  "config": {
    "url": "http://loki:3100",
    "push_url": "http://loki:3100/loki/api/v1/push",
    "tenant_id": "",
    "username": "",
    "password": "",
    "batch_size": 1048576,
    "timeout": "10s",
    "labels": {
      "job": "default"
    }
  }
}
```

#### OpenObserve

```json
{
  "sink_type": "openobserve",
  "protocol": "multi",
  "config": {
    "url": "http://openobserve:5080",
    "organization": "default",
    "stream_name": "metrics",
    "username": "admin",
    "password": "***",
    "timeout": "30s"
  }
}
```

#### S3/MinIO

```json
{
  "sink_type": "s3",
  "protocol": "multi",
  "config": {
    "endpoint": "http://minio:9000",
    "bucket": "monitor-archive",
    "region": "us-east-1",
    "access_key": "AKIAIOSFODNN7EXAMPLE",
    "secret_key": "***",
    "prefix": "data/",
    "compression": "gzip",
    "partition": "day"
  }
}
```

#### Custom（自定义）

```json
{
  "sink_type": "custom",
  "protocol": "multi",
  "config": {
    "adapter_class": "com.example.MySinkAdapter",
    "url": "http://custom-backend:8080",
    "extra": {}
  }
}
```

### 4.3 关联关系

```
zone (1) ──< (N) data_sink
  │                │
  │                ├── is_default=true (网区默认存储)
  │                │
  │                └──< (N) collection_task (通过 data_sink_id 引用, 可选)
  │
  └── default_sink_id ──> data_sink (网区级别的默认存储引用)
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| M1 网区与网段 | zone_id 校验、网区存在性确认 | 内部 API 调用 |
| 基础设施层 | 存储后端连接（网络可达性） | 直接连接 |

### 5.2 下游提供

| 消费模块 | 提供内容 | 交互方式 |
|----------|----------|----------|
| T1 采集任务 | 存储目标列表、默认存储信息 | 查询 API |
| 采集器适配器 | 存储连接参数（生成 remote_write 配置） | 内部 API 调用 |
| I3 状态缓存 | 存储健康状态 | 定时任务触发 |
| A1 告警策略 | 存储查询接口（告警规则评估时查询历史数据） | 查询 API |

### 5.3 对外 API

#### 5.3.1 注册数据存储

```
POST /api/v1/zones/{zone_id}/data-sinks
```

**请求体（VictoriaMetrics 示例）：**

```json
{
  "sink_id": "sink-vm-east",
  "name": "VM 生产存储",
  "sink_type": "victoriametrics",
  "protocol": "metric",
  "is_default": true,
  "description": "华东区 VictoriaMetrics 生产存储",
  "config": {
    "url": "http://vm-prod:8428",
    "remote_write_url": "http://vm-prod:8428/api/v1/write",
    "remote_read_url": "http://vm-prod:8428/api/v1/query",
    "timeout": "30s"
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "sink_id": "sink-vm-east",
    "zone_id": "zone-east-1",
    "name": "VM 生产存储",
    "sink_type": "victoriametrics",
    "protocol": "metric",
    "is_default": true,
    "status": "active",
    "health_check": {
      "status": "passed",
      "latency_ms": 25,
      "checked_at": "2026-10-27T10:00:00Z"
    },
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**注册时自动执行：**
1. 校验连接参数格式
2. 尝试连接存储后端，验证可达性
3. 若 `is_default=true`，取消同网区其他存储的默认标记

#### 5.3.2 查询存储列表

```
GET /api/v1/zones/{zone_id}/data-sinks
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sink_type | string | 否 | 按存储类型过滤 |
| protocol | string | 否 | 按协议过滤：metric/log/trace/multi |
| status | string | 否 | 按状态过滤 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "sink_id": "sink-vm-east",
        "name": "VM 生产存储",
        "sink_type": "victoriametrics",
        "protocol": "metric",
        "is_default": true,
        "status": "active",
        "last_health_at": "2026-10-27T10:30:00Z",
        "last_health_ok": true,
        "task_count": 120,
        "created_at": "2026-10-27T10:00:00Z"
      },
      {
        "sink_id": "sink-es-east",
        "name": "ES 日志存储",
        "sink_type": "elasticsearch",
        "protocol": "log",
        "is_default": false,
        "status": "active",
        "last_health_at": "2026-10-27T10:30:00Z",
        "last_health_ok": true,
        "task_count": 45,
        "created_at": "2026-10-27T10:05:00Z"
      }
    ],
    "total": 2,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询存储详情

```
GET /api/v1/data-sinks/{sink_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "sink_id": "sink-vm-east",
    "zone_id": "zone-east-1",
    "zone_name": "华东区",
    "name": "VM 生产存储",
    "sink_type": "victoriametrics",
    "protocol": "metric",
    "is_default": true,
    "status": "active",
    "description": "华东区 VictoriaMetrics 生产存储",
    "config": {
      "url": "http://vm-prod:8428",
      "remote_write_url": "http://vm-prod:8428/api/v1/write",
      "timeout": "30s"
    },
    "health": {
      "status": "healthy",
      "last_check_at": "2026-10-27T10:30:00Z",
      "latency_ms": 25,
      "uptime": "720h",
      "consecutive_failures": 0
    },
    "statistics": {
      "total_tasks": 120,
      "write_rate_per_min": 50000,
      "storage_size_bytes": 107374182400,
      "query_avg_latency_ms": 150
    },
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T12:00:00Z"
  }
}
```

**注意：** 响应中的 `config` 字段会脱敏处理，隐藏密码、密钥等敏感字段。

#### 5.3.4 更新存储配置

```
PUT /api/v1/data-sinks/{sink_id}
```

**请求体：**

```json
{
  "name": "VM 生产存储（v2）",
  "config": {
    "url": "http://vm-prod-v2:8428",
    "remote_write_url": "http://vm-prod-v2:8428/api/v1/write",
    "timeout": "60s"
  }
}
```

#### 5.3.5 删除存储

```
DELETE /api/v1/data-sinks/{sink_id}
```

**前置校验：**
- 存储非默认存储（默认存储不可删除，需先更换默认）
- 无采集任务引用（task_count = 0）

#### 5.3.6 设置默认存储

```
POST /api/v1/zones/{zone_id}/data-sinks/{sink_id}/set-default
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "zone_id": "zone-east-1",
    "previous_default": "sink-vm-old",
    "current_default": "sink-vm-east"
  }
}
```

**说明：** 设置新默认存储时，自动取消原默认存储的 `is_default` 标记。每个网区有且仅有一个默认存储。

#### 5.3.7 连接校验

```
POST /api/v1/data-sinks/{sink_id}/test
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "test_id": "test-sink-001",
    "status": "passed",
    "connectivity": {
      "reachable": true,
      "latency_ms": 25
    },
    "authentication": {
      "valid": true
    },
    "write_test": {
      "success": true,
      "sample_count": 10
    },
    "tested_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.8 健康检查

```
GET /api/v1/data-sinks/{sink_id}/health
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "sink_id": "sink-vm-east",
    "status": "healthy",
    "checks": [
      {
        "name": "connectivity",
        "status": "ok",
        "latency_ms": 25
      },
      {
        "name": "disk_usage",
        "status": "ok",
        "value": "65%"
      },
      {
        "name": "write_throughput",
        "status": "ok",
        "value": "50000 samples/min"
      }
    ],
    "last_check_at": "2026-10-27T10:30:00Z",
    "consecutive_ok": 120,
    "consecutive_failures": 0
  }
}
```

#### 5.3.9 查询写入路由

```
GET /api/v1/tasks/{task_id}/data-sink
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "task_id": "task-001",
    "resolved_sink": {
      "sink_id": "sink-vm-east",
      "name": "VM 生产存储",
      "sink_type": "victoriametrics",
      "protocol": "metric",
      "resolution": "task_explicit"
    }
  }
}
```

**resolution 字段说明：**
- `task_explicit`：采集任务明确指定了目标存储
- `zone_default`：采集任务未指定，使用网区默认存储

---

## 六、设计决策

### 6.1 DataSinkAdapter 统一接口

**决策：** 所有存储后端通过统一的 `DataSinkAdapter` 接口适配，接口定义如下：

```go
// DataSinkAdapter 数据存储适配器接口
type DataSinkAdapter interface {
    // Type 返回存储类型标识
    Type() string

    // Validate 校验连接配置
    Validate(config DataSinkConfig) error

    // Write 写入数据（metric/log/trace）
    Write(data []DataPoint) error

    // Query 查询数据（用于告警规则评估）
    Query(query string, start, end time.Time) ([]DataPoint, error)

    // HealthCheck 检查存储健康状态
    HealthCheck() (*SinkHealthStatus, error)
}
```

**理由：**
- 统一接口使存储后端可替换，符合开闭原则
- 新增存储类型只需实现接口，不影响现有代码
- Write/Query 接口覆盖数据采集与告警评估两大核心场景
- HealthCheck 接口支持主动健康监控

### 6.2 网区默认存储机制

**决策：** 每个网区有且仅有一个默认存储（`is_default=true`），采集任务未指定目标时自动路由到默认存储。

**理由：**
- 简化配置：大多数场景下，一个网区的数据写入同一个存储
- 灵活性保留：采集任务可显式指定不同存储（如日志写 ES、指标写 VM）
- 默认存储可切换：更换存储后端时只需修改默认标记，无需更新所有任务

### 6.3 数据写入路由逻辑

**决策：** 路由优先级为：采集任务显式指定 > 网区默认存储。

```
采集任务 (CollectionTask)
    │
    ├── 指定 data_sink_id → 写入对应存储
    │
    └── 未指定 → 使用网区默认存储 (zone.default_sink_id)
```

**理由：**
- 两级路由策略简洁明了，覆盖 95% 的使用场景
- 显式指定适用于多存储场景（如 metric 写 VM、log 写 ES）
- 默认存储适用于单存储场景，减少配置负担

### 6.4 多协议支持策略

**决策：** 通过 `protocol` 字段标识存储支持的协议类型，支持 metric / log / trace / multi 四种值。

| 协议 | 适用存储 | 说明 |
|------|----------|------|
| `metric` | VictoriaMetrics, Prometheus | 时序指标数据 |
| `log` | Elasticsearch, Loki | 日志数据 |
| `trace` | Elasticsearch, Jaeger | 链路追踪数据 |
| `multi` | OpenObserve, S3 | 多协议混合数据 |

**理由：**
- 不同存储后端擅长的数据类型不同，protocol 字段明确标识能力边界
- `multi` 类型支持统一可观测平台（如 OpenObserve），一个存储接收多种数据
- 路由时可校验协议匹配：metric 类型的采集任务不应路由到仅支持 log 的存储

### 6.5 配置脱敏策略

**决策：** API 响应中的 `config` 字段自动脱敏，隐藏密码、密钥、Token 等敏感信息。

**理由：**
- 凭据类信息不应通过 API 明文传输
- 脱敏规则基于字段名匹配（password / secret / token / key 等关键词）
- 编辑配置时，前端显示占位符（如 `******`），用户不修改则保持原值
- 内部服务间调用可通过特殊 Header 获取完整配置（需鉴权）

### 6.6 健康检查机制

**决策：** 采用定时健康检查 + 手动触发检查的双模式。

**理由：**
- 定时检查（默认 60s 间隔）由 T5 定时任务模块调度，结果写入 Redis 缓存
- 手动检查通过 API 触发，适用于注册新存储或故障排查场景
- 健康状态三态：healthy / degraded / offline
- 连续失败次数（consecutive_failures）用于判断是否触发告警

### 6.7 内置存储适配器实现

| 存储类型 | 适配器 | 协议 | 说明 |
|----------|--------|------|------|
| VictoriaMetrics | `VictoriaMetricsAdapter` | metric | 通过 remote_write API 写入，PromQL 查询 |
| Elasticsearch | `ElasticsearchAdapter` | log/trace | 通过 Bulk API 写入，Query DSL 查询 |
| Loki | `LokiAdapter` | log | 通过 Push API 写入，LogQL 查询 |
| OpenObserve | `OpenObserveAdapter` | multi | 通过 HTTP API 写入，支持多协议 |
| S3/MinIO | `S3Adapter` | multi | 通过 S3 API 写入，支持归档场景 |
| Custom | `CustomAdapter` | 任意 | 用户实现 DataSinkAdapter 接口 |
