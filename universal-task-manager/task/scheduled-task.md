# T5 定时任务

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：T5 | 职责：管理平台所有周期性后台作业，包括状态聚合、数据清理、报告生成、健康检查等内置任务类型以及用户自定义任务，提供 Cron 调度、执行记录、失败重试与任务编排能力。

---

## 一、概述

定时任务模块是平台的自动化运维引擎。它负责调度和执行各类周期性后台作业，从状态聚合（驱动 I3 状态缓存更新）、数据清理（过期数据回收）、报告生成（日报/周报），到健康检查（采集器与实例连通性巡检）以及用户自定义脚本任务。

本模块的核心价值在于：

- **统一调度**：所有周期性后台作业集中管理，避免分散的 cron job 和定时器。
- **Cron 表达式**：标准 Cron 语法定义调度周期，灵活且易理解。
- **执行可观测**：每次执行均有记录（scheduled_task_execution），含耗时、状态、输出。
- **失败重试**：内置指数退避重试机制，避免瞬时故障导致任务遗漏。
- **与状态缓存联动**：I3 状态缓存的聚合更新由本模块的 status_aggregate 任务驱动。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 定时任务 CRUD | 创建、查询、更新、删除定时任务 |
| Cron 调度 | 基于 Cron 表达式计算下次执行时间并触发执行 |
| 执行记录管理 | 记录每次执行的开始/结束时间、状态、输出、错误 |
| 失败重试 | 执行失败后按策略自动重试 |
| 内置任务类型 | status_aggregate / data_cleanup / report_generate / health_check |
| 自定义任务支持 | 用户通过 config 定义自定义执行逻辑 |
| 任务启停控制 | 运行时启用/禁用任务 |
| 并发控制 | 同一任务不重复执行（前一次未完成则跳过） |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 状态缓存的实际读写操作 | I3 状态缓存模块（本模块只负责调度触发） |
| 数据的物理删除 | 数据存储层（本模块发出清理指令） |
| 告警通知的发送 | 通知层（NotificationChannel） |
| 采集任务的执行 | T1 采集任务模块 |
| 服务发现的目标刷新 | T4 服务发现模块（有独立的刷新调度） |

---

## 三、功能清单

| 编号 | 功能 | 优先级 | 说明 |
|------|------|--------|------|
| F5.1 | 定时任务创建 | P0 | 指定名称、类型、Cron 表达式、配置 |
| F5.2 | 定时任务查询 | P0 | 按类型 / 状态 / 业务系统查询 |
| F5.3 | 定时任务更新 | P0 | 更新 Cron 表达式、配置等 |
| F5.4 | 定时任务删除 | P1 | 删除前需先禁用 |
| F5.5 | 任务启用/禁用 | P0 | 运行时启停调度 |
| F5.6 | Cron 调度引擎 | P0 | 解析 Cron 表达式，计算 next_run_at，到时触发执行 |
| F5.7 | 执行记录写入 | P0 | 每次执行写入 scheduled_task_execution 记录 |
| F5.8 | 执行状态跟踪 | P0 | running / success / failed / skipped / timeout |
| F5.9 | 失败重试 | P0 | 指数退避重试，最大重试次数可配置 |
| F5.10 | 并发控制 | P0 | 同一任务前一次未完成则跳过本次 |
| F5.11 | 超时控制 | P1 | 单次执行超时后强制终止 |
| F5.12 | 状态聚合任务 | P0 | 内置 status_aggregate 类型，驱动 I3 状态缓存更新 |
| F5.13 | 数据清理任务 | P1 | 内置 data_cleanup 类型，清理过期执行记录与历史数据 |
| F5.14 | 报告生成任务 | P1 | 内置 report_generate 类型，生成日报/周报/月报 |
| F5.15 | 健康检查任务 | P0 | 内置 health_check 类型，巡检采集器与实例连通性 |
| F5.16 | 自定义任务 | P2 | 用户通过 config 定义自定义执行逻辑（脚本/HTTP回调） |
| F5.17 | 手动触发 | P1 | 不等待 Cron 调度，立即执行一次 |
| F5.18 | 执行日志查询 | P0 | 查询历史执行记录，含状态、耗时、输出 |
| F5.19 | 任务统计 | P1 | 成功率、平均耗时、最近执行时间等统计信息 |
| F5.20 | 失败告警 | P1 | 连续失败超过阈值时触发告警通知 |

---

## 四、核心数据模型

### 4.1 scheduled_task（定时任务）

```sql
CREATE TABLE scheduled_task (
    task_id             VARCHAR(64) PRIMARY KEY,
    business_system_id  VARCHAR(64) COMMENT '所属业务系统',
    name                VARCHAR(256) NOT NULL COMMENT '任务名称',
    description         TEXT COMMENT '任务描述',
    task_type           VARCHAR(32) NOT NULL COMMENT 'status_aggregate/data_cleanup/report_generate/health_check/custom',
    cron_expr           VARCHAR(128) NOT NULL COMMENT 'Cron表达式',
    config              JSON NOT NULL COMMENT '任务配置(类型专属)',
    enabled             BOOLEAN DEFAULT TRUE,
    last_run_at         TIMESTAMP NULL COMMENT '最近一次执行时间',
    next_run_at         TIMESTAMP NULL COMMENT '下次计划执行时间',
    status              VARCHAR(32) DEFAULT 'idle' COMMENT 'idle/running/error/disabled',
    retry_policy        JSON COMMENT '重试策略 {max_retries, initial_interval, multiplier}',
    timeout_seconds     INT DEFAULT 300 COMMENT '单次执行超时(秒)',
    concurrency_policy  VARCHAR(32) DEFAULT 'skip' COMMENT 'skip/queue/replace',
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_st_type ON scheduled_task(task_type);
CREATE INDEX idx_st_status ON scheduled_task(status);
CREATE INDEX idx_st_enabled ON scheduled_task(enabled);
CREATE INDEX idx_st_next_run ON scheduled_task(next_run_at);
CREATE INDEX idx_st_business ON scheduled_task(business_system_id);
```

### 4.2 scheduled_task_execution（执行记录）

```sql
CREATE TABLE scheduled_task_execution (
    execution_id    VARCHAR(64) PRIMARY KEY,
    task_id         VARCHAR(64) NOT NULL COMMENT '所属定时任务',
    status          VARCHAR(32) NOT NULL COMMENT 'running/success/failed/skipped/timeout',
    started_at      TIMESTAMP NOT NULL COMMENT '执行开始时间',
    finished_at     TIMESTAMP NULL COMMENT '执行结束时间',
    duration_ms     INT NULL COMMENT '执行耗时(毫秒)',
    output          JSON COMMENT '执行输出(类型专属)',
    error_message   TEXT COMMENT '错误信息',
    retry_count     INT DEFAULT 0 COMMENT '已重试次数',
    triggered_by    VARCHAR(32) DEFAULT 'scheduler' COMMENT 'scheduler/manual',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (task_id) REFERENCES scheduled_task(task_id) ON DELETE CASCADE
);

CREATE INDEX idx_ste_task ON scheduled_task_execution(task_id);
CREATE INDEX idx_ste_status ON scheduled_task_execution(status);
CREATE INDEX idx_ste_started ON scheduled_task_execution(started_at);
```

### 4.3 任务类型与 config 结构

**status_aggregate（状态聚合）**

```json
{
  "scope": "zone",
  "scope_value": "zone-east-1",
  "aggregate_targets": ["instance_status", "task_status", "node_status", "alert_count"],
  "cache_ttl_seconds": 60,
  "notify_on_change": true
}
```

**data_cleanup（数据清理）**

```json
{
  "cleanup_targets": [
    {
      "table": "scheduled_task_execution",
      "retention_days": 30,
      "batch_size": 1000
    },
    {
      "table": "alert_event",
      "retention_days": 90,
      "batch_size": 500
    },
    {
      "table": "rc_rule_history",
      "retention_days": 180,
      "batch_size": 200
    }
  ],
  "dry_run": false,
  "vacuum_after_delete": false
}
```

**report_generate（报告生成）**

```json
{
  "report_type": "daily",
  "scope": "zone",
  "scope_value": "zone-east-1",
  "sections": ["overview", "alert_summary", "task_statistics", "capacity_trend"],
  "output_format": "html",
  "delivery": {
    "channels": ["email", "feishu"],
    "recipients": ["ops-team@example.com"]
  }
}
```

**health_check（健康检查）**

```json
{
  "check_targets": ["collector_nodes", "instances", "service_discoveries"],
  "zone_id": "zone-east-1",
  "check_timeout_seconds": 10,
  "unhealthy_threshold": 3,
  "notify_on_unhealthy": true,
  "auto_recover": false
}
```

**custom（自定义任务）**

```json
{
  "executor": "http_callback",
  "url": "http://internal-service:8080/api/custom-job",
  "method": "POST",
  "headers": { "Authorization": "Bearer ***" },
  "body": { "param1": "value1" },
  "success_condition": "response.code == 200",
  "timeout_seconds": 120
}
```

### 4.4 重试策略

```json
{
  "max_retries": 3,
  "initial_interval_seconds": 10,
  "multiplier": 2.0,
  "max_interval_seconds": 300,
  "retryable_errors": ["timeout", "connection_refused", "temporary_failure"]
}
```

**退避计算**：

| 重试次数 | 等待时间 |
|----------|----------|
| 第 1 次 | 10s |
| 第 2 次 | 20s (10 × 2.0) |
| 第 3 次 | 40s (20 × 2.0) |
| 超过 max_retries | 标记任务为 error，触发告警 |

### 4.5 状态枚举定义

**scheduled_task.status：**

| 状态 | 含义 | 允许转换 |
|------|------|----------|
| `idle` | 空闲，等待下次调度 | → running, disabled |
| `running` | 正在执行中 | → idle, error |
| `error` | 执行异常（重试耗尽） | → idle, disabled |
| `disabled` | 手动禁用 | → idle |

**scheduled_task_execution.status：**

| 状态 | 含义 |
|------|------|
| `running` | 执行中 |
| `success` | 执行成功 |
| `failed` | 执行失败（将重试） |
| `skipped` | 跳过（并发控制或禁用状态） |
| `timeout` | 执行超时 |

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| I3 状态缓存 | status_aggregate 任务读写缓存数据 | 内部服务调用 |
| 数据存储层 | data_cleanup 任务清理过期数据 | 内部服务调用 |
| T1 采集任务 | health_check 任务检查采集任务状态 | 内部服务调用 |
| T2 RC 任务 | health_check 任务检查规则下发状态 | 内部服务调用 |
| T4 服务发现 | health_check 任务检查服务发现状态 | 内部服务调用 |
| 通知层 | 报告投递、失败告警 | 内部服务调用 |

### 5.2 下游提供

| 消费方 | 提供内容 | 交互方式 |
|--------|----------|----------|
| I3 状态缓存 | 驱动状态聚合更新 | 内部服务调用 |
| 前端 UI | 执行记录与统计信息 | REST API |
| 运维人员 | 报告（HTML / PDF） | 邮件 / IM 推送 |

### 5.3 对外 API

#### 5.3.1 定时任务管理

**创建定时任务**

```
POST /api/v1/scheduled-tasks
```

请求体（健康检查示例）：

```json
{
  "name": "华东区采集器健康巡检",
  "description": "每5分钟检查华东区所有采集器节点的健康状态",
  "task_type": "health_check",
  "cron_expr": "*/5 * * * *",
  "config": {
    "check_targets": ["collector_nodes"],
    "zone_id": "zone-east-1",
    "check_timeout_seconds": 10,
    "unhealthy_threshold": 3,
    "notify_on_unhealthy": true
  },
  "retry_policy": {
    "max_retries": 2,
    "initial_interval_seconds": 30,
    "multiplier": 2.0
  },
  "timeout_seconds": 60,
  "concurrency_policy": "skip"
}
```

响应：

```json
{
  "code": 0,
  "data": {
    "task_id": "st-health-001",
    "name": "华东区采集器健康巡检",
    "task_type": "health_check",
    "cron_expr": "*/5 * * * *",
    "next_run_at": "2026-10-27T10:05:00Z",
    "status": "idle",
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**查询定时任务列表**

```
GET /api/v1/scheduled-tasks?task_type=health_check&status=idle&enabled=true&business_system_id=biz-001&page=1&page_size=20
```

**查询单个定时任务**

```
GET /api/v1/scheduled-tasks/{task_id}
```

**更新定时任务**

```
PUT /api/v1/scheduled-tasks/{task_id}
```

> 更新 cron_expr 后自动重新计算 next_run_at。

**删除定时任务**

```
DELETE /api/v1/scheduled-tasks/{task_id}
```

> 前置校验：需先禁用（enabled=false）且无正在执行的实例。

**启用任务**

```
POST /api/v1/scheduled-tasks/{task_id}/enable
```

**禁用任务**

```
POST /api/v1/scheduled-tasks/{task_id}/disable
```

**手动触发执行**

```
POST /api/v1/scheduled-tasks/{task_id}/trigger
```

响应：

```json
{
  "code": 0,
  "data": {
    "execution_id": "exec-001",
    "task_id": "st-health-001",
    "triggered_by": "manual",
    "status": "running",
    "started_at": "2026-10-27T10:02:00Z"
  }
}
```

#### 5.3.2 执行记录管理

**查询执行记录列表**

```
GET /api/v1/scheduled-tasks/{task_id}/executions?status=failed&page=1&page_size=20
```

响应：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "execution_id": "exec-001",
        "task_id": "st-health-001",
        "status": "failed",
        "started_at": "2026-10-27T10:00:00Z",
        "finished_at": "2026-10-27T10:00:45Z",
        "duration_ms": 45000,
        "retry_count": 2,
        "error_message": "Connection refused: collector-node-east-2:9090",
        "triggered_by": "scheduler"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

**查询单个执行记录**

```
GET /api/v1/scheduled-task-executions/{execution_id}
```

**查询执行输出**

```
GET /api/v1/scheduled-task-executions/{execution_id}/output
```

#### 5.3.3 统计与监控

**任务执行统计**

```
GET /api/v1/scheduled-tasks/{task_id}/statistics?period=7d
```

响应：

```json
{
  "code": 0,
  "data": {
    "task_id": "st-health-001",
    "period": "7d",
    "total_executions": 2016,
    "success_count": 2010,
    "failed_count": 4,
    "skipped_count": 2,
    "success_rate": 0.997,
    "avg_duration_ms": 12500,
    "max_duration_ms": 45000,
    "min_duration_ms": 8000,
    "p95_duration_ms": 25000,
    "last_run_at": "2026-10-27T10:00:00Z",
    "last_status": "success",
    "next_run_at": "2026-10-27T10:05:00Z"
  }
}
```

**全局任务概览**

```
GET /api/v1/scheduled-tasks/overview
```

响应：

```json
{
  "code": 0,
  "data": {
    "total_tasks": 25,
    "enabled_tasks": 22,
    "running_now": 3,
    "error_tasks": 1,
    "by_type": {
      "status_aggregate": 5,
      "data_cleanup": 3,
      "report_generate": 4,
      "health_check": 10,
      "custom": 3
    },
    "last_24h": {
      "total_executions": 1850,
      "success_rate": 0.995,
      "avg_duration_ms": 15000
    }
  }
}
```

---

## 六、设计决策

### 6.1 内置任务类型 + 自定义扩展

**决策**：提供 status_aggregate / data_cleanup / report_generate / health_check 四种内置类型，同时支持 custom 类型通过 HTTP 回调或脚本执行自定义逻辑。

**理由**：
- 内置类型覆盖平台 80% 的运维场景，开箱即用。
- custom 类型保留扩展性，用户可对接自有系统。
- 每种内置类型有专属的 config 结构和 output 格式，语义明确。
- custom 类型通过 executor 字段区分 http_callback / shell_script 等执行方式。

### 6.2 与状态缓存（I3）的关系

**决策**：I3 状态缓存的聚合更新由 T5 的 status_aggregate 任务周期性调度驱动。

**交互流程**：

```
T5 调度器 (每 60 秒)
    │
    ▼
status_aggregate 任务执行
    │
    ├── 读取各 zone 的 instance / task / node 状态
    ├── 聚合计算: total_instances, task_running, task_error, active_alerts...
    ├── 写入 Redis: HSET zone:{zone_id}:overview ...
    └── 写入 Redis: HSET global:summary ...
```

**理由**：
- 状态缓存需要定期刷新以保持数据新鲜度。
- 由 T5 统一调度，避免独立的定时器散落各处。
- 聚合逻辑作为任务执行，可配置执行频率（如 30s / 60s / 5min）。
- 执行结果写入 Redis，前端直接读取展示。

### 6.3 失败重试机制

**决策**：采用指数退避重试策略，配置最大重试次数、初始间隔和退避乘数。

**重试流程**：

```
执行失败
    │
    ▼
检查 retry_count < max_retries ?
    │
    ├── 是 → 等待 initial_interval × multiplier^retry_count
    │         └── 重新执行
    │               ├── 成功 → 记录 success, 重置 retry_count
    │               └── 失败 → retry_count++, 继续重试
    │
    └── 否 → 标记任务 status = error
              └── 触发告警通知
              └── 等待下次 Cron 调度重新执行
```

**理由**：
- 指数退避避免频繁重试加重系统负担。
- 最大重试次数防止无限重试。
- 重试耗尽后标记 error 并告警，确保问题被关注。
- 下次 Cron 调度周期到达时自动重新尝试，恢复自愈。

### 6.4 并发控制策略

**决策**：支持 skip / queue / replace 三种并发控制策略，默认 skip。

| 策略 | 行为 | 适用场景 |
|------|------|----------|
| `skip` | 前一次未完成，跳过本次调度 | 健康检查、状态聚合（允许漏执行） |
| `queue` | 排队等待前一次完成后执行 | 报告生成（不可漏执行） |
| `replace` | 终止前一次，立即执行本次 | 数据清理（以最新配置为准） |

**理由**：
- 不同任务对执行完整性的要求不同。
- skip 策略最安全，避免任务堆积。
- queue 策略保证每次调度都被执行。
- replace 策略适用于配置变更后需要立即生效的场景。

### 6.5 Cron 表达式支持

**决策**：使用标准 5 字段 Cron 表达式（分 时 日 月 周），支持特殊字符 `*/,-`。

**示例**：

| Cron 表达式 | 含义 |
|-------------|------|
| `*/5 * * * *` | 每 5 分钟 |
| `0 * * * *` | 每小时整点 |
| `0 0 * * *` | 每天 00:00 |
| `0 9 * * 1-5` | 工作日 09:00 |
| `0 0 1 * *` | 每月 1 日 00:00 |
| `*/30 * * * *` | 每 30 分钟 |

**调度引擎流程**：

```
调度循环 (每秒检查)
    │
    ▼
遍历所有 enabled=true 的 scheduled_task
    │
    ├── next_run_at <= now ?
    │     ├── 是 → 提交执行
    │     │         ├── 更新 last_run_at = now
    │     │         ├── 计算 next_run_at (CronParser.next(cron_expr, now))
    │     │         └── 写入 execution 记录
    │     │
    │     └── 否 → 跳过
    │
    └── 继续下一个任务
```

### 6.6 执行记录保留策略

**决策**：执行记录默认保留 30 天，由 data_cleanup 类型的定时任务自动清理。

**理由**：
- 执行记录用于故障排查与统计分析，需要一定保留期。
- 无限增长的执行记录会占用大量存储空间。
- 通过内置的 data_cleanup 任务自动清理，无需人工干预。
- 保留天数可在 data_cleanup 任务的 config 中配置。

### 6.7 任务编排（未来扩展）

**决策**：当前版本不支持任务间的依赖编排，但数据模型预留了扩展空间。

**理由**：
- v1.0 聚焦单任务调度，降低复杂度。
- 未来可通过增加 task_dependency 表实现 DAG 编排。
- 当前可通过 custom 类型的 HTTP 回调串联多个任务，实现简易编排。
- 状态聚合任务（status_aggregate）天然在健康检查任务之后执行，通过 Cron 时间错开实现隐式顺序。
