# 数据模型设计

> 多网区监控采集管理平台全量数据模型，覆盖核心资源层、任务管理层、告警层、基础设施层、业务系统层、数据存储层、服务发现与定时任务。

---

## 一、核心实体关系

```
Zone (网区)
  |-- NetworkSegment (网段) x N
  |-- ZoneCoordinator (区协调器, gossip) x 1  ← 管理本区所有job分配
  |-- CollectorNode (采集节点) x N            ← 由协调器分配owned任务
  |-- RCWorker (vmalert节点) x N              ← 由协调器分配owned规则
  |-- Credential (凭据) x N
  |-- DataSink (数据存储) x N
  |-- ServiceDiscovery (服务发现) x N
  |
  +-- Instance (实例) x N
        |-- resource_type_id --> ResourceType (资源类型树)
        |-- business_system_id --> BusinessSystem (业务系统)
        |-- CollectionTask (采集任务) x N
        |     +-- task_type_id --> TaskType (任务类型)
        |     +-- data_sink_id --> DataSink (目标存储)
        |
        |-- RCTask (RC任务) x N
        |     +-- RCRule (规则) x N
        |     (归属哪个vmalert由gossip一致性哈希决定，不存worker_id)
        |
        +-- AlertStrategy (告警策略) x N
              |-- metric_id --> MetricDefinition (指标)
              +-- alert_group_id --> AlertGroup (告警组)

告警流水线:
  vmalert(×N/区) → 区AM(去重) → 核心AM(策略路由) → Kafka → 通知

ResourceType (资源类型树，全局)
  |-- parent_id --> ResourceType (自引用)
  |-- ResourceTypeMetric (公共指标) x N
  |     +-- metric_id --> MetricDefinition

BusinessSystem (业务系统，全局)
  |-- BusinessSystemMember (成员) x N
  |     +-- user_id --> User
  +-- BusinessSystemZone (网区关联) x N
        +-- zone_id --> Zone

ScheduledTask (定时任务)
  |-- business_system_id --> BusinessSystem (可选关联)
  +-- ScheduledTaskExecution (执行记录) x N

User (用户，全局)
  |-- UserAuthSource (认证源)
  +-- UserNotificationTarget (通知目标) x N
        +-- channel_id --> NotificationChannel (通知渠道)
```

---

## 二、核心资源层

### 2.1 Zone（网区）

```sql
CREATE TABLE zone (
    zone_id         VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    collector_type  VARCHAR(32) NOT NULL COMMENT '主采集器: vmagent/prometheus/alloy',
    remote_write_url VARCHAR(512),
    default_sink_id VARCHAR(64) COMMENT '默认数据存储，引用 data_sink.sink_id',
    config          JSON,
    status          VARCHAR(32) DEFAULT 'active',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 2.2 NetworkSegment（网段）

```sql
CREATE TABLE network_segment (
    segment_id      BIGINT AUTO_INCREMENT PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    cidr            VARCHAR(45) NOT NULL COMMENT 'CIDR格式: 10.0.1.0/24',
    name            VARCHAR(128) COMMENT '网段名称',
    description     TEXT,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id) ON DELETE CASCADE,
    UNIQUE KEY uk_zone_cidr (zone_id, cidr)
);

CREATE INDEX idx_segment_cidr ON network_segment(cidr);
```

### 2.3 ResourceType（资源类型树）

```sql
CREATE TABLE resource_type (
    type_id         VARCHAR(64) PRIMARY KEY,
    parent_id       VARCHAR(64) COMMENT '父类型ID, NULL=根节点',
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    icon            VARCHAR(64),
    sort_order      INT DEFAULT 0 COMMENT '同级排序',
    attributes      JSON COMMENT '类型专属属性Schema',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES resource_type(type_id) ON DELETE SET NULL
);

CREATE INDEX idx_rt_parent ON resource_type(parent_id);
```

### 2.4 ResourceTypeMetric（类型公共指标）

```sql
CREATE TABLE resource_type_metric (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    type_id         VARCHAR(64) NOT NULL,
    metric_id       VARCHAR(64) NOT NULL COMMENT '引用指标定义',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (type_id) REFERENCES resource_type(type_id) ON DELETE CASCADE,
    UNIQUE KEY uk_type_metric (type_id, metric_id)
);
```

### 2.5 Instance（实例）

```sql
CREATE TABLE instance (
    instance_id     VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    resource_type_id VARCHAR(64) NOT NULL,
    business_system_id VARCHAR(64) COMMENT '所属业务系统，引用 business_system.system_id',
    name            VARCHAR(256) NOT NULL,
    host            VARCHAR(256) NOT NULL COMMENT 'IP或域名',
    port            INT,
    credential_id   VARCHAR(64),
    labels          JSON,
    extra           JSON COMMENT '类型专属扩展属性',
    status          VARCHAR(32) DEFAULT 'pending' COMMENT 'pending/active/suspended/decommissioned',
    last_test_at    TIMESTAMP NULL,
    last_test_result VARCHAR(32),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id),
    FOREIGN KEY (resource_type_id) REFERENCES resource_type(type_id)
);

CREATE INDEX idx_inst_zone ON instance(zone_id);
CREATE INDEX idx_inst_type ON instance(resource_type_id);
CREATE INDEX idx_inst_host ON instance(host);
CREATE INDEX idx_inst_status ON instance(status);
```

### 2.6 Credential（凭据）

```sql
CREATE TABLE credential (
    credential_id   VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    cred_type       VARCHAR(32) NOT NULL COMMENT 'basic_auth/ssh_key/snmp/bearer/tls',
    encrypted_data  TEXT NOT NULL COMMENT 'AES加密后的凭据数据',
    version         INT DEFAULT 1,
    expires_at      TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);
```

---

## 三、任务管理层

### 3.1 TaskType（采集任务类型）

```sql
CREATE TABLE task_type (
    type_id         VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    source_type     VARCHAR(32) NOT NULL COMMENT '适用资源类型',
    collector       VARCHAR(32) NOT NULL COMMENT '默认采集器',
    form_schema     JSON NOT NULL COMMENT '表单Schema',
    detail_panels   JSON COMMENT '详情页面板',
    declared_metrics JSON COMMENT '声明的指标列表',
    config_template JSON NOT NULL COMMENT '各采集器配置模板',
    is_builtin      BOOLEAN DEFAULT FALSE,
    version         INT DEFAULT 1,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 3.2 CollectionTask（采集任务）

```sql
CREATE TABLE collection_task (
    task_id         VARCHAR(64) PRIMARY KEY,
    instance_id     VARCHAR(64) NOT NULL,
    type_id         VARCHAR(64) NOT NULL,
    module          VARCHAR(32) DEFAULT 'metrics' COMMENT '所属功能模块: metrics/traces/logs/network',
    node_id         VARCHAR(64) COMMENT '分配的采集器节点',
    data_sink_id    VARCHAR(64) COMMENT '目标存储，未指定则用网区默认；引用 data_sink.sink_id',
    name            VARCHAR(256) NOT NULL,
    config          JSON NOT NULL,
    enabled         BOOLEAN DEFAULT TRUE,
    status          VARCHAR(32) DEFAULT 'pending',
    last_scrape_at  TIMESTAMP NULL,
    last_error      TEXT,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (instance_id) REFERENCES instance(instance_id) ON DELETE CASCADE,
    FOREIGN KEY (type_id) REFERENCES task_type(type_id)
);

CREATE INDEX idx_ct_instance ON collection_task(instance_id);
CREATE INDEX idx_ct_type ON collection_task(type_id);
CREATE INDEX idx_ct_module ON collection_task(module);
CREATE INDEX idx_ct_node ON collection_task(node_id);
CREATE INDEX idx_ct_status ON collection_task(status);
```

### 3.3 RCTask（RC任务）

```sql
CREATE TABLE rc_task (
    task_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(256) NOT NULL,
    description     TEXT,
    strategy_id     VARCHAR(64) COMMENT '来源告警策略',
    query_address   VARCHAR(256) COMMENT '规则评估查询地址(覆盖集群默认)',
    enabled         BOOLEAN DEFAULT TRUE,
    status          VARCHAR(32) DEFAULT 'draft' COMMENT 'draft/active/paused/archived',
    rule_group      VARCHAR(128) COMMENT 'Prometheus rule group名称',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_rct_zone ON rc_task(zone_id);
CREATE INDEX idx_rct_strategy ON rc_task(strategy_id);
```

### 3.4 RCWorker（vmalert 工作节点 — 区协调器分配）

```sql
CREATE TABLE rc_worker (
    worker_id       VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '所属网区',
    name            VARCHAR(128) NOT NULL COMMENT '显示名称',
    address         VARCHAR(256) NOT NULL COMMENT 'vmalert HTTP地址(规则拉取/热加载)',
    query_address   VARCHAR(256) NOT NULL COMMENT '查询地址(规则评估)',
    capacity        INT DEFAULT 500 COMMENT '可承载最大规则数',
    owned_rules     INT DEFAULT 0 COMMENT '当前拥有规则数(协调器分配后)',
    status          VARCHAR(32) DEFAULT 'unknown' COMMENT 'healthy/degraded/offline',
    last_heartbeat  TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_rcw_zone ON rc_worker(zone_id);
CREATE INDEX idx_rcw_status ON rc_worker(status);
```

### 3.5 ZoneCoordinator（区协调器）

```sql
CREATE TABLE zone_coordinator (
    coordinator_id  VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '负责网区',
    name            VARCHAR(128) NOT NULL COMMENT '显示名称',
    address         VARCHAR(256) NOT NULL COMMENT '协调器HTTP地址',
    gossip_addr     VARCHAR(256) NOT NULL COMMENT 'gossip通信地址',
    role            VARCHAR(32) DEFAULT 'leader' COMMENT 'leader/standby(高可用)',
    managed_nodes   INT DEFAULT 0 COMMENT '管理的节点总数(vmalert+vmagent+prometheus)',
    status          VARCHAR(32) DEFAULT 'unknown' COMMENT 'active/standby/offline',
    last_heartbeat  TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_zc_zone ON zone_coordinator(zone_id);
CREATE INDEX idx_zc_status ON zone_coordinator(status);
```

> **区协调器**：每区一个（可主备），运行 gossip 协议，统一管理本区 vmalert/vmagent/prometheus 的 job 分配。一致性哈希打散规则/任务到各节点，节点故障时自动 rebalance。T2 只管规则发布，I1 只管节点注册，实际分配决策由区协调器完成。

### 3.5 RCRule（RC规则）

```sql
CREATE TABLE rc_rule (
    rule_id         VARCHAR(64) PRIMARY KEY,
    task_id         VARCHAR(64) NOT NULL,
    module          VARCHAR(32) DEFAULT 'metrics' COMMENT '所属功能模块: metrics/traces/logs/network',
    alert_name      VARCHAR(256) NOT NULL,
    expr            TEXT NOT NULL COMMENT 'PromQL表达式(变量已替换)',
    duration        VARCHAR(32) NOT NULL COMMENT '持续时间: 5m, 10s',
    severity        VARCHAR(32) NOT NULL COMMENT 'critical/warning/info',
    labels          JSON COMMENT '附加标签',
    annotations     JSON COMMENT 'summary, description模板',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES rc_task(task_id) ON DELETE CASCADE
);

CREATE INDEX idx_rule_task ON rc_rule(task_id);
CREATE INDEX idx_rule_module ON rc_rule(module);
```

### 3.6 MetricDefinition（指标定义）

```sql
CREATE TABLE metric_definition (
    metric_id       VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL COMMENT '指标名',
    category        VARCHAR(32) NOT NULL COMMENT 'event/metric',
    metric_type     VARCHAR(32) COMMENT 'gauge/counter/histogram/summary',
    help            TEXT,
    unit            VARCHAR(64),
    promql_template TEXT COMMENT 'PromQL模板, 含$variable占位符',
    variables       JSON COMMENT '变量定义: [{name, type, default}]',
    labels          JSON COMMENT '标签定义',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_metric_category ON metric_definition(category);
CREATE INDEX idx_metric_name ON metric_definition(name);
```

### 3.7 CommonQuery（常用查询集）

```sql
CREATE TABLE common_query (
    query_id        VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL,
    description     TEXT,
    metric_id       VARCHAR(64) NOT NULL,
    promql          TEXT NOT NULL COMMENT '完整PromQL(可含变量)',
    variables       JSON COMMENT '变量定义',
    category        VARCHAR(64) COMMENT '分类标签',
    is_builtin      BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (metric_id) REFERENCES metric_definition(metric_id)
);
```

---

## 四、告警层

### 4.1 AlertStrategy（告警策略）

```sql
CREATE TABLE alert_strategy (
    strategy_id     VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL,
    description     TEXT,
    metric_id       VARCHAR(64) NOT NULL COMMENT '关联指标',
    expr_template   TEXT NOT NULL COMMENT '告警表达式模板',
    duration        VARCHAR(32) DEFAULT '5m',
    severity        VARCHAR(32) NOT NULL DEFAULT 'warning',
    scope_type      VARCHAR(32) NOT NULL COMMENT 'instance/type/zone',
    scope_value     TEXT COMMENT '作用域值(JSON数组)',
    alert_group_id  VARCHAR(64) COMMENT '告警组',
    labels          JSON,
    annotations     JSON,
    enabled         BOOLEAN DEFAULT TRUE,
    status          VARCHAR(32) DEFAULT 'draft',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (metric_id) REFERENCES metric_definition(metric_id)
);

CREATE INDEX idx_strategy_metric ON alert_strategy(metric_id);
CREATE INDEX idx_strategy_group ON alert_strategy(alert_group_id);
CREATE INDEX idx_strategy_status ON alert_strategy(status);
```

### 4.2 AlertStrategyInstance（策略-实例关联）

```sql
CREATE TABLE alert_strategy_instance (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    strategy_id     VARCHAR(64) NOT NULL,
    instance_id     VARCHAR(64) NOT NULL,
    zone_id         VARCHAR(64) NOT NULL COMMENT '用于按DC拆分',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (strategy_id) REFERENCES alert_strategy(strategy_id) ON DELETE CASCADE,
    FOREIGN KEY (instance_id) REFERENCES instance(instance_id) ON DELETE CASCADE,
    UNIQUE KEY uk_strategy_instance (strategy_id, instance_id)
);

CREATE INDEX idx_asi_zone ON alert_strategy_instance(zone_id);
```

### 4.3 AlertGroup（告警组）

```sql
CREATE TABLE alert_group (
    group_id        VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    parent_id       VARCHAR(64) COMMENT '父组(升级链)',
    escalation_minutes INT COMMENT '升级超时(分钟)',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES alert_group(group_id)
);
```

### 4.4 AlertGroupMember（告警组成员）

```sql
CREATE TABLE alert_group_member (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    group_id        VARCHAR(64) NOT NULL,
    user_id         VARCHAR(64) NOT NULL,
    role            VARCHAR(32) DEFAULT 'member' COMMENT 'member/manager',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES alert_group(group_id) ON DELETE CASCADE,
    UNIQUE KEY uk_group_user (group_id, user_id)
);
```

### 4.5 NotificationChannel（通知渠道）

```sql
CREATE TABLE notification_channel (
    channel_id      VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    channel_type    VARCHAR(32) NOT NULL COMMENT 'email/sms/feishu/dingtalk/wechat/webhook',
    config          JSON NOT NULL COMMENT '渠道配置(加密存储敏感字段)',
    enabled         BOOLEAN DEFAULT TRUE,
    last_test_at    TIMESTAMP NULL,
    last_test_result VARCHAR(32),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 4.6 AlertEvent（告警事件，从Kafka消费后落库）

```sql
CREATE TABLE alert_event (
    event_id        BIGINT AUTO_INCREMENT PRIMARY KEY,
    alert_name      VARCHAR(256) NOT NULL,
    zone_id         VARCHAR(64),
    instance_id     VARCHAR(64),
    severity        VARCHAR(32),
    status          VARCHAR(32) COMMENT 'firing/resolved',
    labels          JSON,
    annotations     JSON,
    starts_at       TIMESTAMP,
    ends_at         TIMESTAMP NULL,
    fingerprint     VARCHAR(128),
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ae_zone ON alert_event(zone_id);
CREATE INDEX idx_ae_status ON alert_event(status);
CREATE INDEX idx_ae_time ON alert_event(starts_at);
CREATE INDEX idx_ae_fingerprint ON alert_event(fingerprint);
```

---

## 五、基础设施层

### 5.1 CollectorNode（采集器节点）

```sql
CREATE TABLE collector_node (
    node_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    node_type       VARCHAR(32) NOT NULL COMMENT 'vmagent/vmalert/otel-collector/exporter/prometheus/alloy',
    address         VARCHAR(256) NOT NULL COMMENT 'http://ip:port',
    health_endpoint VARCHAR(256) DEFAULT '/-/healthy',
    labels          JSON,
    status          VARCHAR(32) DEFAULT 'unknown' COMMENT 'healthy/degraded/offline',
    task_count      INT DEFAULT 0,
    last_heartbeat  TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_cn_zone ON collector_node(zone_id);
CREATE INDEX idx_cn_status ON collector_node(status);
```

### 5.2 User（用户）

```sql
CREATE TABLE user (
    user_id         VARCHAR(64) PRIMARY KEY,
    username        VARCHAR(128) NOT NULL UNIQUE,
    display_name    VARCHAR(256),
    email           VARCHAR(256),
    phone           VARCHAR(32),
    role            VARCHAR(32) DEFAULT 'viewer' COMMENT 'admin/operator/viewer',
    status          VARCHAR(32) DEFAULT 'active',
    last_login_at   TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 5.3 UserAuthSource（用户认证源）

```sql
CREATE TABLE user_auth_source (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL,
    source_type     VARCHAR(32) NOT NULL COMMENT 'local/ldap/oauth/oidc',
    external_id     VARCHAR(256) COMMENT 'LDAP DN / OAuth subject',
    password_hash   VARCHAR(256) COMMENT '仅local类型',
    extra           JSON,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
           UNIQUE KEY uk_user_source (user_id, source_type)
);
```

### 5.4 UserNotificationTarget（用户通知目标）

```sql
CREATE TABLE user_notification_target (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL,
    channel_type    VARCHAR(32) NOT NULL COMMENT 'email/sms/feishu/dingtalk/wechat',
    target_value    VARCHAR(256) NOT NULL COMMENT '具体地址/ID',
    is_primary      BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_unt_user ON user_notification_target(user_id);
```

---

## 六、状态缓存（Redis）

```
# 全局概览
HSET global:summary
  total_zones       "5"
  total_instances   "1200"
  total_tasks       "3600"
  active_alerts     "12"
  healthy_nodes     "15"
  offline_nodes     "1"

# 网区概览
HSET zone:{zone_id}:overview
  instance_count    "240"
  task_running      "720"
  task_error        "3"
  active_alerts     "2"
  healthy_nodes     "3"
  last_updated      "1727000100"

# 节点状态
HSET node:{node_id}:status
  status            "healthy"
  task_count        "120"
  uptime            "864000"
  last_heartbeat    "1727000090"

# 实例状态
HSET instance:{instance_id}:status
  status            "active"
  task_count        "3"
  last_scrape       "1727000085"
  error_count       "0"
```

---

## 七、业务系统层

### 7.1 BusinessSystem（业务系统）

```sql
-- 业务系统
CREATE TABLE business_system (
    system_id       VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    status          VARCHAR(32) DEFAULT 'active' COMMENT 'active/inactive',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 7.2 BusinessSystemMember（业务系统成员）

```sql
-- 业务系统成员
CREATE TABLE business_system_member (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    system_id       VARCHAR(64) NOT NULL,
    user_id         VARCHAR(64) NOT NULL,
    role            VARCHAR(32) NOT NULL DEFAULT 'viewer' COMMENT 'admin/operator/viewer',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (system_id) REFERENCES business_system(system_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    UNIQUE KEY uk_system_user (system_id, user_id)
);
```

### 7.3 BusinessSystemZone（业务系统-网区关联）

```sql
-- 业务系统-网区关联
CREATE TABLE business_system_zone (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    system_id       VARCHAR(64) NOT NULL,
    zone_id         VARCHAR(64) NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (system_id) REFERENCES business_system(system_id) ON DELETE CASCADE,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id) ON DELETE CASCADE,
    UNIQUE KEY uk_system_zone (system_id, zone_id)
);
```

---

## 八、数据存储层

### 8.1 DataSink（数据存储）

```sql
-- 数据存储
CREATE TABLE data_sink (
    sink_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    sink_type       VARCHAR(32) NOT NULL COMMENT 'victoriametrics/elasticsearch/loki/openobserve/s3/custom',
    protocol        VARCHAR(32) NOT NULL COMMENT 'prometheus/elasticsearch/loki/s3/http',
    config          JSON NOT NULL COMMENT '连接配置(加密敏感字段)',
    is_default      BOOLEAN DEFAULT FALSE COMMENT '是否为网区默认存储',
    status          VARCHAR(32) DEFAULT 'active' COMMENT 'active/inactive/error',
    last_check_at   TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_sink_zone ON data_sink(zone_id);
CREATE INDEX idx_sink_type ON data_sink(sink_type);
```

> zone 表中的 `default_sink_id` 字段引用 `data_sink.sink_id`，表示该网区的默认数据存储目标。当 `collection_task.data_sink_id` 未指定时，系统自动回退到对应网区的 `default_sink_id`。

---

## 九、服务发现与定时任务

### 9.1 ServiceDiscovery（服务发现配置）

```sql
-- 服务发现配置
CREATE TABLE service_discovery (
    discovery_id    VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    protocol        VARCHAR(32) NOT NULL COMMENT 'http_sd/file_sd/dns_sd/consul_sd/kubernetes_sd',
    config          JSON NOT NULL COMMENT '协议专属配置',
    refresh_interval VARCHAR(32) DEFAULT '30s' COMMENT '刷新间隔',
    enabled         BOOLEAN DEFAULT TRUE,
    status          VARCHAR(32) DEFAULT 'active' COMMENT 'active/error/paused',
    last_refresh_at TIMESTAMP NULL,
    discovered_count INT DEFAULT 0 COMMENT '最近发现的目标数',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_sd_zone ON service_discovery(zone_id);
CREATE INDEX idx_sd_protocol ON service_discovery(protocol);
```

### 9.2 ScheduledTask（定时任务）

```sql
-- 定时任务
CREATE TABLE scheduled_task (
    task_id         VARCHAR(64) PRIMARY KEY,
    business_system_id VARCHAR(64) COMMENT '关联业务系统(NULL=全局)',
    name            VARCHAR(256) NOT NULL,
    task_type       VARCHAR(32) NOT NULL COMMENT 'status_aggregate/data_cleanup/report_generate/health_check/custom',
    cron_expr       VARCHAR(128) NOT NULL COMMENT 'Cron表达式',
    config          JSON COMMENT '任务配置',
    enabled         BOOLEAN DEFAULT TRUE,
    last_run_at     TIMESTAMP NULL,
    next_run_at     TIMESTAMP NULL,
    status          VARCHAR(32) DEFAULT 'idle' COMMENT 'idle/running/error',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_st_type ON scheduled_task(task_type);
CREATE INDEX idx_st_status ON scheduled_task(status);
CREATE INDEX idx_st_next_run ON scheduled_task(next_run_at);
```

### 9.3 ScheduledTaskExecution（定时任务执行记录）

```sql
-- 定时任务执行记录
CREATE TABLE scheduled_task_execution (
    execution_id    BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id         VARCHAR(64) NOT NULL,
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP NULL,
    duration_ms     INT,
    status          VARCHAR(32) NOT NULL COMMENT 'success/error/timeout',
    error_message   TEXT,
    FOREIGN KEY (task_id) REFERENCES scheduled_task(task_id) ON DELETE CASCADE
);

CREATE INDEX idx_ste_task ON scheduled_task_execution(task_id);
CREATE INDEX idx_ste_time ON scheduled_task_execution(started_at);
```
