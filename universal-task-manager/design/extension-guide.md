# 扩展开发指南

> 描述如何通过注册新的任务类型来扩展系统，支持新的采集场景。

---

## 一、扩展场景

| 场景 | 说明 | 扩展示例 |
|------|------|----------|
| 新监控目标 | 系统尚未支持的中间件/设备 | 添加 Redis、MongoDB、Kafka 任务类型 |
| 自定义采集 | 用户特有的采集需求 | 添加自定义 SQL 采集任务类型 |
| 新采集器 | 接入新的采集组件 | 实现 Telegraf、Datadog Agent 适配器 |

---

## 二、添加新任务类型

### 2.1 任务类型定义结构

```json
{
  "type_id": "唯一标识",
  "name": "显示名称",
  "description": "类型描述",
  "source_type": "适用的资源类型",
  "collector": "默认采集器",

  "form_schema": { "表单定义" },
  "detail_panels": { "详情页面板定义" },
  "declared_metrics": { "声明的指标列表" },
  "config_template": { "各采集器的配置模板" }
}
```

### 2.2 表单 Schema 定义

`form_schema.fields` 数组定义前端表单字段：

```json
{
  "form_schema": {
    "fields": [
      {
        "key": "字段标识",
        "label": "显示标签",
        "type": "字段类型",
        "required": true/false,
        "default": "默认值",
        "options": [],
        "placeholder": "提示文本",
        "help": "帮助说明",
        "group": "分组名称"
      }
    ]
  }
}
```

**支持的字段类型：**

| type | 说明 | 适用场景 |
|------|------|----------|
| `text` | 单行文本 | 名称、路径 |
| `number` | 数字输入 | 端口、超时 |
| `select` | 下拉选择 | 间隔、协议 |
| `boolean` | 开关 | 启用/禁用 |
| `tags` | 标签输入 | 数据库列表 |
| `textarea` | 多行文本 | SQL 语句 |
| `password` | 密码输入 | 凭据 |
| `json` | JSON 编辑器 | 高级配置 |

### 2.3 详情页面板定义

`detail_panels` 定义任务详情页的展示布局：

```json
{
  "detail_panels": [
    {
      "key": "面板标识",
      "title": "面板标题",
      "type": "panel_type",
      "fields": ["field1", "field2"],
      "display": "展示方式"
    }
  ]
}
```

**面板类型：**

| type | 说明 |
|------|------|
| `info` | 信息展示，显示字段值 |
| `metric_list` | 指标列表，展示 declared_metrics |
| `log_viewer` | 日志查看器 |
| `chart` | 图表展示 |

### 2.4 指标声明

`declared_metrics` 声明该任务类型预期产出的指标：

```json
{
  "declared_metrics": [
    {
      "name": "指标名称",
      "type": "gauge|counter|histogram|summary",
      "help": "指标说明",
      "labels": ["标签列表"]
    }
  ]
}
```

### 2.5 配置模板

`config_template` 为每种采集器提供配置生成模板：

```json
{
  "config_template": {
    "vmagent": {
      "job_name": "{{type_id}}-{{instance_id}}",
      "scrape_interval": "{{config.scrape_interval}}",
      "static_configs": [{
        "targets": ["{{instance.host}}:{{instance.port}}"]
      }]
    },
    "alloy": {
      "prometheus.scrape": {
        "job_name": "{{type_id}}-{{instance_id}}",
        "scrape_interval": "{{config.scrape_interval}}"
      }
    }
  }
}
```

**模板变量：**

| 变量 | 说明 |
|------|------|
| `{{instance_id}}` | 实例 ID |
| `{{type_id}}` | 任务类型 ID |
| `{{task_id}}` | 任务 ID |
| `{{instance.host}}` | 实例主机 |
| `{{instance.port}}` | 实例端口 |
| `{{instance.credential_id}}` | 凭据 ID |
| `{{config.xxx}}` | 任务表单中的配置值 |
| `{{zone.id}}` | 网区 ID |

---

## 三、完整示例：Redis 任务类型

### 3.1 任务类型 JSON

```json
{
  "type_id": "redis-standard",
  "name": "Redis 标准采集",
  "description": "采集Redis实例的运行状态、内存使用、命令统计等指标",
  "source_type": "redis",
  "collector": "vmagent",

  "form_schema": {
    "fields": [
      {
        "key": "scrape_interval",
        "label": "采集间隔",
        "type": "select",
        "options": ["10s", "15s", "30s", "60s"],
        "default": "15s",
        "required": true,
        "group": "基础配置"
      },
      {
        "key": "redis_password",
        "label": "Redis 密码",
        "type": "password",
        "required": false,
        "placeholder": "如无需认证请留空",
        "group": "认证配置"
      },
      {
        "key": "enable_keyspace",
        "label": "启用键空间统计",
        "type": "boolean",
        "default": true,
        "group": "采集项"
      },
      {
        "key": "enable_slowlog",
        "label": "启用慢查询采集",
        "type": "boolean",
        "default": false,
        "group": "采集项"
      },
      {
        "key": "slowlog_max_len",
        "label": "慢查询最大条数",
        "type": "number",
        "default": 100,
        "min": 10,
        "max": 1000,
        "group": "采集项",
        "visible_when": "enable_slowlog=true"
      },
      {
        "key": "databases",
        "label": "监控数据库编号",
        "type": "tags",
        "default": ["0"],
        "placeholder": "输入数据库编号，回车添加",
        "group": "采集项"
      }
    ]
  },

  "detail_panels": [
    {
      "key": "connection",
      "title": "连接信息",
      "type": "info",
      "fields": ["host", "port", "auth_status"]
    },
    {
      "key": "collection",
      "title": "采集配置",
      "type": "info",
      "fields": ["scrape_interval", "enable_keyspace", "enable_slowlog", "databases"]
    },
    {
      "key": "metrics",
      "title": "指标覆盖",
      "type": "metric_list",
      "display": "table"
    },
    {
      "key": "logs",
      "title": "运行日志",
      "type": "log_viewer",
      "display": "scroll"
    }
  ],

  "declared_metrics": [
    { "name": "redis_up", "type": "gauge", "help": "Redis实例是否可达" },
    { "name": "redis_info", "type": "gauge", "help": "Redis INFO 信息", "labels": ["section", "key"] },
    { "name": "redis_connected_clients", "type": "gauge", "help": "已连接客户端数" },
    { "name": "redis_used_memory_bytes", "type": "gauge", "help": "已使用内存(字节)" },
    { "name": "redis_used_memory_peak_bytes", "type": "gauge", "help": "内存峰值(字节)" },
    { "name": "redis_total_commands_processed_total", "type": "counter", "help": "处理的命令总数" },
    { "name": "redis_keyspace_hits_total", "type": "counter", "help": "键命中次数" },
    { "name": "redis_keyspace_misses_total", "type": "counter", "help": "键未命中次数" },
    { "name": "redis_slowlog_length", "type": "gauge", "help": "慢查询日志长度" },
    { "name": "redis_db_keys", "type": "gauge", "help": "数据库键数量", "labels": ["db"] }
  ],

  "config_template": {
    "vmagent": {
      "job_name": "redis-{{instance_id}}",
      "scrape_interval": "{{config.scrape_interval}}",
      "static_configs": [{
        "targets": ["{{instance.host}}:{{instance.port}}"],
        "labels": {
          "instance": "{{instance_id}}",
          "zone": "{{zone.id}}"
        }
      }],
      "params": {
        "redis.password": ["{{config.redis_password}}"],
        "redis.keyspace": ["{{config.enable_keyspace}}"],
        "redis.slowlog": ["{{config.enable_slowlog}}"]
      }
    },
    "alloy": {
      "prometheus.scrape": {
        "job_name": "redis-{{instance_id}}",
        "scrape_interval": "{{config.scrape_interval}}",
        "targets": [{
          "__address__": "{{instance.host}}:{{instance.port}}",
          "instance": "{{instance_id}}"
        }],
        "forward_to": ["prometheus.remote_write.default.receiver"]
      }
    }
  }
}
```

### 3.2 注册 API 调用

```bash
curl -X POST http://localhost:8080/api/v1/task-types \
  -H "Content-Type: application/json" \
  -d @redis-task-type.json
```

### 3.3 创建实例

```bash
curl -X POST http://localhost:8080/api/v1/zones/zone-east-1/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Redis 生产集群",
    "resource_type": "redis",
    "host": "10.0.1.10",
    "port": 6379,
    "credential_id": "cred-redis-001",
    "labels": {"env": "prod", "role": "cache"}
  }'
```

### 3.4 创建采集任务

```bash
curl -X POST http://localhost:8080/api/v1/instances/inst-redis-001/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "type_id": "redis-standard",
    "name": "Redis 生产采集",
    "config": {
      "scrape_interval": "15s",
      "redis_password": "",
      "enable_keyspace": true,
      "enable_slowlog": true,
      "slowlog_max_len": 100,
      "databases": ["0", "1", "2"]
    }
  }'
```

---

## 四、自定义 SQL 采集任务类型

### 4.1 场景说明

用户有私有数据库，需要通过自定义 SQL 查询采集指标。

### 4.2 任务类型定义

```json
{
  "type_id": "custom-sql",
  "name": "自定义 SQL 采集",
  "description": "通过自定义 SQL 查询采集业务指标",
  "source_type": "database",
  "collector": "custom",

  "form_schema": {
    "fields": [
      {
        "key": "query",
        "label": "SQL 查询语句",
        "type": "textarea",
        "required": true,
        "placeholder": "SELECT value, timestamp FROM metrics WHERE ...",
        "help": "查询必须返回 value 和 timestamp 两列",
        "group": "查询配置"
      },
      {
        "key": "scrape_interval",
        "label": "采集间隔",
        "type": "select",
        "options": ["30s", "60s", "5m", "15m"],
        "default": "60s",
        "required": true,
        "group": "采集配置"
      },
      {
        "key": "metric_name",
        "label": "产出指标名",
        "type": "text",
        "required": true,
        "placeholder": "my_custom_metric",
        "group": "指标配置"
      },
      {
        "key": "metric_type",
        "label": "指标类型",
        "type": "select",
        "options": ["gauge", "counter"],
        "default": "gauge",
        "group": "指标配置"
      },
      {
        "key": "labels",
        "label": "附加标签",
        "type": "json",
        "default": "{}",
        "help": "JSON 格式的标签键值对",
        "group": "指标配置"
      }
    ]
  },

  "detail_panels": [
    {
      "key": "query_info",
      "title": "查询信息",
      "type": "info",
      "fields": ["query", "scrape_interval"]
    },
    {
      "key": "metric_info",
      "title": "指标信息",
      "type": "info",
      "fields": ["metric_name", "metric_type", "labels"]
    },
    {
      "key": "execution_log",
      "title": "执行日志",
      "type": "log_viewer"
    }
  ],

  "declared_metrics": [
    {
      "name": "{{config.metric_name}}",
      "type": "{{config.metric_type}}",
      "help": "自定义 SQL 采集指标"
    }
  ],

  "config_template": {
    "custom": {
      "type": "sql_collector",
      "query": "{{config.query}}",
      "interval": "{{config.scrape_interval}}",
      "output": {
        "metric_name": "{{config.metric_name}}",
        "metric_type": "{{config.metric_type}}",
        "labels": "{{config.labels}}"
      }
    }
  }
}
```

---

## 五、最佳实践

### 5.1 表单设计

- 将相关字段分组（`group`），便于用户理解
- 为必填字段设置 `required: true`
- 提供合理的 `default` 值
- 使用 `placeholder` 和 `help` 引导用户
- 使用 `visible_when` 实现条件显示

### 5.2 指标声明

- 指标名遵循 Prometheus 命名规范：`<namespace>_<subsystem>_<name>_<unit>`
- 为每个指标提供清晰的 `help` 说明
- 区分 gauge（瞬时值）和 counter（累计值）
- 声明所有可能出现的标签

### 5.3 配置模板

- 模板变量使用 `{{xxx}}` 语法
- 为不同采集器提供对应的配置模板
- 保持模板简洁，复杂逻辑放在适配器层

### 5.4 版本管理

- 任务类型支持版本化（通过 `updated_at` 追踪）
- 已创建的任务不受类型更新影响（快照语义）
- 如需强制更新所有任务，提供批量刷新 API

---

## 六、调试与测试

### 6.1 预览生成的配置

```bash
# 预览 vmagent 配置
curl http://localhost:8080/api/v1/tasks/{task_id}/config/vmagent

# 预览 Alloy 配置
curl http://localhost:8080/api/v1/tasks/{task_id}/config/alloy
```

### 6.2 验证指标声明

```bash
# 查询任务实际产出的指标
curl http://localhost:8080/api/v1/tasks/{task_id}/metrics/actual

# 对比声明的指标
curl http://localhost:8080/api/v1/task-types/{type_id}/metrics
```

### 6.3 模拟采集

```bash
# 触发一次测试采集
curl -X POST http://localhost:8080/api/v1/tasks/{task_id}/test
```
