# T4 服务发现

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：T4 | 职责：管理平台的服务发现机制，支持 http_sd / file_sd / dns_sd / consul_sd / kubernetes_sd 五种协议，管理发现配置、目标刷新、自动创建/更新采集任务，实现监控目标的动态感知。

---

## 一、概述

服务发现模块是平台实现自动化监控覆盖的核心组件。在大规模多网区环境中，手动维护监控目标列表不可持续——服务实例的创建、销毁、扩缩容都是动态发生的。本模块通过多种服务发现协议，自动感知网区内的监控目标变化，并将发现的目标同步为采集任务。

本模块的核心价值在于：

- **多协议支持**：覆盖 HTTP 文件、DNS、Consul、Kubernetes 等主流服务发现方式。
- **目标自动同步**：发现新目标自动创建采集任务，目标消失自动清理，减少人工干预。
- **配置集中管理**：所有服务发现配置统一存储在 service_discovery 表中，按 zone 隔离。
- **刷新可控**：每个发现源独立配置 refresh_interval，平衡实时性与系统负载。
- **与采集任务联动**：服务发现是采集任务的上游触发器，发现的 target 自动驱动 T1 模块。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 服务发现配置管理 | service_discovery 表的 CRUD |
| 多协议适配 | http_sd / file_sd / dns_sd / consul_sd / kubernetes_sd 的协议适配 |
| 目标发现与解析 | 从各协议获取 target 列表并解析为标准格式 |
| 目标刷新调度 | 按 refresh_interval 周期性刷新目标列表 |
| 采集任务联动 | 发现目标变化时自动创建/更新/删除采集任务 |
| http_sd 端点暴露 | 为采集器提供 HTTP 拉取目标列表的端点 |
| file_sd 文件生成 | 为采集器生成 JSON 格式的目标文件 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 采集任务的生命周期管理 | T1 采集任务模块 |
| 采集器节点的配置下发 | 采集器适配层 |
| DNS 服务器 / Consul 集群 / K8s API 的运维 | 外部基础设施 |
| 目标的连通性测试 | T1 采集任务模块（last_test_result） |
| 凭据管理 | 资源管理模块（Credential） |

---

## 三、功能清单

| 编号 | 功能 | 优先级 | 说明 |
|------|------|--------|------|
| F4.1 | 服务发现配置创建 | P0 | 指定 zone + protocol + config + refresh_interval |
| F4.2 | 服务发现配置查询 | P0 | 按 zone / protocol / status 查询列表 |
| F4.3 | 服务发现配置更新 | P0 | 更新配置、刷新间隔等 |
| F4.4 | 服务发现配置删除 | P1 | 删除前需先禁用，并清理关联的采集任务 |
| F4.5 | 服务发现启用/禁用 | P0 | 运行时启停发现过程 |
| F4.6 | http_sd 端点 | P0 | 平台暴露 HTTP 端点，采集器定期 GET 拉取 target 列表 |
| F4.7 | file_sd 文件生成 | P1 | 生成 JSON 文件写入采集器可访问的本地目录，采集器定期读取 |
| F4.8 | dns_sd 发现 | P1 | 通过 DNS SRV / A 记录发现目标 |
| F4.9 | consul_sd 发现 | P1 | 从 Consul 服务目录发现目标 |
| F4.10 | kubernetes_sd 发现 | P2 | 通过 K8s API 发现 Pod / Service / Endpoint 目标 |
| F4.11 | 目标刷新调度 | P0 | 按 refresh_interval 周期性执行目标发现 |
| F4.12 | 目标变更检测 | P0 | 对比前后两次发现结果，检测新增/变更/删除的目标 |
| F4.13 | 采集任务自动创建 | P0 | 发现新目标时，按关联的 task_type 自动创建采集任务 |
| F4.14 | 采集任务自动更新 | P1 | 目标信息变更时更新对应采集任务的配置 |
| F4.15 | 采集任务自动清理 | P1 | 目标消失时禁用或删除对应采集任务 |
| F4.16 | 目标元信息解析 | P0 | 将各协议返回的原始数据解析为标准 target 格式 |
| F4.17 | 发现日志 | P2 | 记录每次发现的 target 数量、耗时、错误 |
| F4.18 | 手动触发发现 | P1 | 立即执行一次目标发现，不等待 refresh_interval |

---

## 四、核心数据模型

### 4.1 service_discovery（服务发现配置）

```sql
CREATE TABLE service_discovery (
    discovery_id    VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '所属网区',
    name            VARCHAR(128) NOT NULL COMMENT '发现配置名称',
    protocol        VARCHAR(32) NOT NULL COMMENT 'http_sd/file_sd/dns_sd/consul_sd/kubernetes_sd',
    config          JSON NOT NULL COMMENT '协议专属配置',
    refresh_interval INT NOT NULL DEFAULT 30 COMMENT '目标刷新间隔(秒)',
    task_type_id    VARCHAR(64) COMMENT '关联的任务类型，用于自动创建采集任务',
    enabled         BOOLEAN DEFAULT TRUE,
    status          VARCHAR(32) DEFAULT 'inactive' COMMENT 'inactive/active/error',
    last_discovery_at TIMESTAMP NULL COMMENT '最近一次发现时间',
    last_target_count INT DEFAULT 0 COMMENT '最近一次发现的目标数',
    last_error      TEXT COMMENT '最近一次错误信息',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id),
    FOREIGN KEY (task_type_id) REFERENCES task_type(type_id)
);

CREATE INDEX idx_sd_zone ON service_discovery(zone_id);
CREATE INDEX idx_sd_protocol ON service_discovery(protocol);
CREATE INDEX idx_sd_status ON service_discovery(status);
CREATE INDEX idx_sd_enabled ON service_discovery(enabled);
```

**各协议 config 结构：**

**http_sd config：**

```json
{
  "listen_path": "/api/v1/sd/{zone_id}/targets",
  "auth_required": false,
  "response_format": "prometheus_http_sd"
}
```

**file_sd config：**

```json
{
  "output_dir": "/etc/prometheus/sd/",
  "file_pattern": "{zone_id}-{discovery_id}.json",
  "atomic_write": true
}
```

**dns_sd config：**

```json
{
  "dns_server": "10.0.0.53",
  "query_type": "SRV",
  "names": ["_mysql._tcp.service.consul", "_redis._tcp.service.consul"],
  "port": 9104,
  "refresh_interval": 30
}
```

**consul_sd config：**

```json
{
  "server": "consul.service.internal:8500",
  "token": "***",
  "services": ["mysql", "redis", "nginx"],
  "tags": ["monitor-enabled"],
  "datacenter": "dc-east-1",
  "node_meta": { "env": "prod" }
}
```

**kubernetes_sd config：**

```json
{
  "api_server": "https://k8s-api.internal:6443",
  "bearer_token": "***",
  "role": "endpoints",
  "namespaces": { "names": ["monitoring", "production"] },
  "selectors": [
    {
      "role": "pod",
      "label": "app=mysql,monitor=true"
    }
  ],
  "tls_config": {
    "ca_file": "/etc/ssl/k8s-ca.crt",
    "insecure_skip_verify": false
  }
}
```

### 4.2 sd_target（发现目标缓存，扩展）

```sql
CREATE TABLE sd_target (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    discovery_id    VARCHAR(64) NOT NULL COMMENT '所属服务发现配置',
    target_address  VARCHAR(256) NOT NULL COMMENT '目标地址 host:port',
    labels          JSON COMMENT '目标标签',
    meta            JSON COMMENT '协议元信息 (__meta_consul_tags 等)',
    task_id         VARCHAR(64) COMMENT '关联的采集任务ID',
    status          VARCHAR(32) DEFAULT 'active' COMMENT 'active/stale/removed',
    discovered_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (discovery_id) REFERENCES service_discovery(discovery_id) ON DELETE CASCADE,
    UNIQUE KEY uk_discovery_target (discovery_id, target_address)
);

CREATE INDEX idx_sdt_discovery ON sd_target(discovery_id);
CREATE INDEX idx_sdt_status ON sd_target(status);
CREATE INDEX idx_sdt_task ON sd_target(task_id);
```

### 4.3 协议与标准 target 格式映射

所有协议发现的目标最终转换为统一的标准 target 格式：

```json
{
  "targets": ["10.0.1.5:9104"],
  "labels": {
    "job": "mysql-standard",
    "instance": "inst-mysql-001",
    "zone": "zone-east-1",
    "source_type": "mysql",
    "__meta_discovery_id": "sd-dns-001",
    "__meta_discovery_protocol": "dns_sd",
    "__meta_task_type_id": "mysql-standard"
  }
}
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| 资源管理（Zone） | 获取 zone 信息、网段列表 | 内部服务调用 |
| T1 采集任务（TaskType） | 获取关联的 task_type 用于自动创建采集任务 | 内部服务调用 |
| 资源管理（Instance） | 目标发现后创建/关联 instance 记录 | 内部服务调用 |
| 外部 DNS 服务器 | dns_sd 协议的查询目标 | DNS 协议 |
| 外部 Consul 集群 | consul_sd 协议的服务目录 | HTTP API |
| 外部 K8s API Server | kubernetes_sd 协议的资源发现 | HTTPS API |

### 5.2 下游提供

| 消费方 | 提供内容 | 交互方式 |
|--------|----------|----------|
| 采集器（vmagent/Prometheus） | http_sd 端点返回 target 列表 | HTTP GET（采集器拉取） |
| 采集器（Prometheus） | file_sd 生成的 JSON 文件 | 文件系统读取 |
| T1 采集任务模块 | 自动创建/更新/删除采集任务 | 内部服务调用 |
| 前端 UI | 服务发现配置与目标列表 | REST API |

### 5.3 对外 API

#### 5.3.1 服务发现配置管理

**创建服务发现配置**

```
POST /api/v1/zones/{zone_id}/service-discoveries
```

请求体（dns_sd 示例）：

```json
{
  "name": "MySQL DNS 发现",
  "protocol": "dns_sd",
  "config": {
    "dns_server": "10.0.0.53",
    "query_type": "SRV",
    "names": ["_mysql._tcp.service.consul"],
    "port": 9104
  },
  "refresh_interval": 30,
  "task_type_id": "mysql-standard"
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "discovery_id": "sd-dns-001",
    "zone_id": "zone-east-1",
    "protocol": "dns_sd",
    "status": "inactive",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**查询服务发现列表**

```
GET /api/v1/zones/{zone_id}/service-discoveries?protocol=dns_sd&enabled=true&page=1&page_size=20
```

**查询单个服务发现配置**

```
GET /api/v1/service-discoveries/{discovery_id}
```

**更新服务发现配置**

```
PUT /api/v1/service-discoveries/{discovery_id}
```

**删除服务发现配置**

```
DELETE /api/v1/service-discoveries/{discovery_id}
```

> 前置校验：需先禁用（enabled=false）且无关联活跃采集任务。

**启用服务发现**

```
POST /api/v1/service-discoveries/{discovery_id}/enable
```

**禁用服务发现**

```
POST /api/v1/service-discoveries/{discovery_id}/disable
```

**手动触发发现**

```
POST /api/v1/service-discoveries/{discovery_id}/refresh
```

响应：

```json
{
  "code": 0,
  "data": {
    "discovery_id": "sd-dns-001",
    "triggered_at": "2026-10-27T10:05:00Z",
    "targets_found": 12,
    "new_targets": 3,
    "removed_targets": 1,
    "duration_ms": 245
  }
}
```

#### 5.3.2 目标查询

**查询发现目标列表**

```
GET /api/v1/service-discoveries/{discovery_id}/targets?status=active&page=1&page_size=50
```

响应：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "target_address": "10.0.1.5:9104",
        "labels": {
          "job": "mysql-standard",
          "instance": "inst-mysql-001",
          "zone": "zone-east-1"
        },
        "task_id": "ct-001",
        "status": "active",
        "discovered_at": "2026-10-27T10:00:00Z"
      }
    ],
    "total": 12,
    "page": 1,
    "page_size": 50
  }
}
```

#### 5.3.3 http_sd 端点（采集器拉取）

**采集器拉取目标列表**

```
GET /api/v1/sd/{zone_id}/targets
```

> 此端点供 vmagent / Prometheus 等采集器定期拉取，返回标准 Prometheus http_sd 格式。

响应：

```json
[
  {
    "targets": ["10.0.1.5:9104", "10.0.1.6:9104"],
    "labels": {
      "job": "mysql-standard",
      "zone": "zone-east-1",
      "__meta_discovery_id": "sd-dns-001",
      "__meta_discovery_protocol": "dns_sd"
    }
  },
  {
    "targets": ["10.0.1.10:9121"],
    "labels": {
      "job": "redis-standard",
      "zone": "zone-east-1",
      "__meta_discovery_id": "sd-consul-001",
      "__meta_discovery_protocol": "consul_sd"
    }
  }
]
```

#### 5.3.4 file_sd 管理

**触发文件生成**

```
POST /api/v1/service-discoveries/{discovery_id}/file-sd/generate
```

**查询生成的文件信息**

```
GET /api/v1/service-discoveries/{discovery_id}/file-sd/status
```

响应：

```json
{
  "code": 0,
  "data": {
    "discovery_id": "sd-file-001",
    "file_path": "/etc/prometheus/sd/zone-east-1-sd-file-001.json",
    "target_count": 15,
    "last_generated_at": "2026-10-27T10:05:00Z",
    "file_size_bytes": 2048
  }
}
```

---

## 六、设计决策

### 6.1 多协议统一抽象

**决策**：定义统一的 ServiceDiscovery 接口，各协议实现各自的发现逻辑，输出标准化的 target 列表。

**接口设计**：

```go
type ServiceDiscovery interface {
    // Protocol 返回协议标识
    Protocol() string

    // Discover 执行一次目标发现
    Discover(ctx context.Context, config *ServiceDiscoveryConfig) ([]*SDTarget, error)

    // ValidateConfig 校验协议配置合法性
    ValidateConfig(config json.RawMessage) error
}

type SDTarget struct {
    Address    string            // host:port
    Labels     map[string]string // 标准标签
    Meta       map[string]string // 协议元信息 (__meta_xxx)
}
```

**理由**：
- 统一接口便于新增协议（如 Nacos、Eureka）只需实现接口。
- 标准化的 target 格式使得下游采集任务创建逻辑与协议无关。
- 各协议的差异封装在各自的实现中，不污染核心逻辑。

### 6.2 http_sd 作为默认协议

**决策**：http_sd 是平台原生支持的默认服务发现协议，采集器通过 HTTP 拉取目标列表。

**架构**：

```
采集器 (vmagent/Prometheus)
    │
    │  每 refresh_interval 秒
    ▼
GET /api/v1/sd/{zone_id}/targets
    │
    ▼
平台 SD 模块
    │
    ├── 查询该 zone 下所有启用的 service_discovery
    ├── 合并各协议的 sd_target 缓存
    └── 返回标准 http_sd JSON 数组
```

**理由**：
- http_sd 是 Prometheus 生态标准协议，vmagent / Alloy / Prometheus 均原生支持。
- 平台作为 HTTP 服务端，采集器作为客户端，架构简单可靠。
- 无需文件系统交互或额外中间件依赖。
- 支持按 zone 聚合多个发现源的结果。

### 6.3 目标变更驱动采集任务联动

**决策**：服务发现检测到目标变化时，自动触发 T1 模块创建/更新/删除采集任务。

**联动流程**：

```
SD 刷新周期到达
    │
    ▼
执行 Discover() 获取新 target 列表
    │
    ▼
与 sd_target 缓存对比
    │
    ├── 新增 target
    │     └── 创建 instance (如不存在)
    │           └── 创建 collection_task (关联 task_type_id)
    │                 └── 触发配置生成与下发
    │
    ├── 变更 target (地址/标签变化)
    │     └── 更新 instance 信息
    │           └── 更新 collection_task.config
    │                 └── 触发配置重新生成
    │
    └── 消失 target
          └── 标记 collection_task.enabled = false
                └── 触发配置移除
```

**理由**：
- 全自动化减少运维人工操作。
- 关联 task_type_id 确保自动创建的任务有正确的采集配置模板。
- 消失目标不立即删除任务（先禁用），保留历史数据与配置，支持人工确认后再清理。

### 6.4 刷新调度策略

**决策**：每个 service_discovery 独立配置 refresh_interval，由统一的调度器驱动。

**调度策略**：

| 协议 | 默认间隔 | 建议范围 | 说明 |
|------|----------|----------|------|
| http_sd | 30s | 10s - 5min | 采集器主动拉取，间隔由采集器控制 |
| file_sd | 60s | 30s - 10min | 文件写入后采集器自动检测 |
| dns_sd | 30s | 15s - 5min | DNS TTL 通常 30s-300s |
| consul_sd | 30s | 10s - 5min | Consul watch 可实时，但轮询更稳定 |
| kubernetes_sd | 60s | 30s - 10min | K8s API 有 rate limit 约束 |

**理由**：
- 独立间隔允许按协议特性和业务需求灵活配置。
- 统一调度器避免每个发现源独立起 goroutine，便于管理与监控。
- 调度器支持 jitter（随机偏移）避免多个发现源同时刷新造成的瞬时负载。

### 6.5 file_sd 原子写入

**决策**：file_sd 生成文件时采用原子写入（先写临时文件再 rename），避免采集器读到不完整文件。

**理由**：
- Prometheus 读取 file_sd 文件时不做文件锁保护。
- 直接写入可能导致采集器读到半写状态的文件。
- 原子写入（write tmp + rename）是 POSIX 保证的原子操作。
- 写入失败时旧文件不受影响，采集器继续使用旧目标列表。

### 6.6 发现目标缓存

**决策**：使用 sd_target 表缓存最近一次发现的目标列表，作为变更检测的基准。

**理由**：
- 缓存避免每次刷新都查询外部系统（DNS / Consul / K8s API）。
- 前后两次缓存对比即可精确检测目标变化。
- 缓存中包含 task_id 关联，便于快速定位目标对应的采集任务。
- status 字段（active / stale / removed）支持优雅下线：先标记 stale，多次刷新确认后才标记 removed。
