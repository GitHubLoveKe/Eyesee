# A3 通知渠道模块

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：A3 | 职责：管理多渠道通知投递能力，维护渠道配置与连通性，提供统一的通知分发接口（NotificationDispatcher）。

---

## 一、概述

通知渠道模块是告警层的投递执行层。它抽象了多种通知渠道（邮件、短信、飞书、钉钉、企业微信、Webhook），为上层告警路由提供统一的消息投递接口。每个渠道通过 `config` JSON 字段存储专属配置，敏感信息（如 API Token、密码）在存储时加密。

模块的核心抽象是 `NotificationDispatcher` 接口，它定义了渠道无关的通知投递契约。默认实现按渠道类型分发到对应的 Sender（EmailSender、FeishuSender 等），同时支持通过 Webhook 渠道扩展自定义投递逻辑。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 渠道 CRUD | 创建、查询、更新、删除通知渠道 |
| 渠道配置管理 | 维护各渠道的专属配置（JSON），加密敏感字段 |
| 连通性测试 | 对渠道执行测试消息发送，记录测试结果 |
| 通知投递 | 通过 NotificationDispatcher 接口向指定渠道发送告警通知 |
| 渠道适配器 | 为每种渠道类型实现对应的 Sender 适配器 |
| 投递记录 | 记录通知投递结果（成功/失败/重试） |
| 用户通知目标管理 | 维护 user_notification_target 关联 |

### 本模块不负责

| 职责 | 归属 |
|------|------|
| 决定"通知谁" | A2 告警组模块（路由解析） |
| 告警事件的生成 | A4 告警收敛模块（AlertEngine） |
| 告警策略与路由规则 | A1 告警策略模块 / A2 告警组模块 |
| 用户账号管理 | 基础设施层（User） |

---

## 三、功能清单

| 功能编号 | 功能名称 | 优先级 | 说明 |
|----------|----------|--------|------|
| A3-F01 | 渠道创建 | P0 | 创建通知渠道，指定类型与配置 |
| A3-F02 | 渠道查询 | P0 | 按类型、名称、启用状态等查询渠道列表 |
| A3-F03 | 渠道更新 | P0 | 修改渠道配置，自动加密敏感字段 |
| A3-F04 | 渠道删除 | P0 | 删除渠道，检查用户通知目标引用 |
| A3-F05 | 渠道启用/禁用 | P0 | 切换渠道启用状态 |
| A3-F06 | 连通性测试 | P0 | 向渠道发送测试消息，验证配置正确性 |
| A3-F07 | 测试结果记录 | P0 | 记录 last_test_at 和 last_test_result |
| A3-F08 | 通知投递 | P0 | 通过 NotificationDispatcher 发送告警通知 |
| A3-F09 | 投递重试 | P1 | 投递失败时自动重试，支持退避策略 |
| A3-F10 | 投递记录查询 | P1 | 查询历史投递记录与结果 |
| A3-F11 | 用户通知目标管理 | P0 | 维护用户与渠道的关联（user_notification_target） |
| A3-F12 | 渠道模板管理 | P2 | 为各渠道类型维护消息模板 |
| A3-F13 | 投递限流 | P1 | 防止告警风暴时的大量通知轰炸 |
| A3-F14 | Webhook 签名验证 | P1 | 为 Webhook 渠道提供签名校验能力 |

---

## 四、核心数据模型

### 4.1 notification_channel（通知渠道表）

> 引用 data-model.md 4.5 节定义。

```sql
CREATE TABLE notification_channel (
    channel_id      VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    channel_type    VARCHAR(32) NOT NULL COMMENT 'email/sms/feishu/dingtalk/wechat/webhook',
    config          JSON NOT NULL COMMENT '渠道配置，敏感字段加密存储',
    enabled         BOOLEAN DEFAULT TRUE,
    last_test_at    TIMESTAMP NULL COMMENT '最近一次连通性测试时间',
    last_test_result VARCHAR(32) COMMENT '最近测试结果: success/failed/timeout',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**渠道类型与配置结构：**

#### Email 渠道

```json
{
  "channel_type": "email",
  "config": {
    "smtp_host": "smtp.company.com",
    "smtp_port": 465,
    "smtp_tls": true,
    "smtp_user": "alert@company.com",
    "smtp_password": "***",
    "from_address": "alert@company.com",
    "from_name": "监控平台告警"
  }
}
```

#### SMS 渠道

```json
{
  "channel_type": "sms",
  "config": {
    "provider": "aliyun",
    "access_key_id": "***",
    "access_key_secret": "***",
    "sign_name": "监控平台",
    "template_code": "SMS_001"
  }
}
```

#### 飞书（Feishu）渠道

```json
{
  "channel_type": "feishu",
  "config": {
    "webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
    "secret": "***",
    "msg_type": "interactive"
  }
}
```

#### 钉钉（DingTalk）渠道

```json
{
  "channel_type": "dingtalk",
  "config": {
    "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    "secret": "***",
    "msg_type": "actionCard",
    "at_mobiles": []
  }
}
```

#### 企业微信（WeChat）渠道

```json
{
  "channel_type": "wechat",
  "config": {
    "webhook_url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx",
    "msg_type": "markdown"
  }
}
```

#### Webhook 渠道

```json
{
  "channel_type": "webhook",
  "config": {
    "url": "https://api.company.com/alerts",
    "method": "POST",
    "headers": {
      "Authorization": "Bearer ***",
      "Content-Type": "application/json"
    },
    "timeout": "10s",
    "retry_count": 3,
    "signature_key": "X-Signature",
    "signature_secret": "***"
  }
}
```

### 4.2 user_notification_target（用户通知目标表）

> 引用 data-model.md 5.4 节定义。

```sql
CREATE TABLE user_notification_target (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL,
    channel_type    VARCHAR(32) NOT NULL COMMENT 'email/sms/feishu/dingtalk/wechat',
    target_value    VARCHAR(256) NOT NULL COMMENT '具体地址/ID，如邮箱地址、手机号、飞书open_id',
    is_primary      BOOLEAN DEFAULT FALSE COMMENT '是否为首选通知方式',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_unt_user ON user_notification_target(user_id);
```

**说明：**

- 一个用户可以有多个通知目标（邮箱、手机、飞书等）。
- `is_primary=true` 的目标为首选通知方式，告警通知优先通过首选渠道投递。
- `channel_type` 与 `notification_channel.channel_type` 对应，但此处不直接引用 `channel_id`，因为用户通知目标是"渠道类型 + 地址"的组合，而非绑定到某个具体渠道实例。

### 4.3 通知投递记录表（建议新增）

```sql
CREATE TABLE notification_delivery (
    delivery_id     BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id        BIGINT NOT NULL COMMENT '关联告警事件ID',
    channel_type    VARCHAR(32) NOT NULL,
    target_value    VARCHAR(256) NOT NULL,
    user_id         VARCHAR(64),
    status          VARCHAR(32) NOT NULL COMMENT 'pending/sent/failed/retrying',
    attempt_count   INT DEFAULT 0 COMMENT '投递尝试次数',
    last_attempt_at TIMESTAMP NULL,
    error_message   TEXT,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_nd_event ON notification_delivery(event_id);
CREATE INDEX idx_nd_status ON notification_delivery(status);
CREATE INDEX idx_nd_time ON notification_delivery(created_at);
```

---

## 五、接口与交互

### 5.1 上游依赖

| 上游模块 | 依赖接口 | 用途 |
|----------|----------|------|
| A2 告警组模块 | 路由解析结果 | 获取需要通知的用户列表及其首选通知渠道 |
| A4 告警收敛模块 | 告警事件数据 | 获取告警事件详情，组装通知消息内容 |

### 5.2 下游提供

| 下游模块 | 提供接口 | 用途 |
|----------|----------|------|
| 外部系统 | 邮件/短信/IM 投递 | 通过各渠道将告警通知投递到用户 |
| 审计日志 | 投递记录 | 通知投递的可追溯记录 |

### 5.3 对外 API

#### 5.3.1 创建通知渠道

```
POST /api/v1/notification-channels
```

**请求体（Email 示例）：**

```json
{
  "name": "公司邮件服务",
  "channel_type": "email",
  "config": {
    "smtp_host": "smtp.company.com",
    "smtp_port": 465,
    "smtp_tls": true,
    "smtp_user": "alert@company.com",
    "smtp_password": "encrypted_password_here",
    "from_address": "alert@company.com",
    "from_name": "监控平台告警"
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "channel_id": "ch-email-001",
    "name": "公司邮件服务",
    "channel_type": "email",
    "enabled": true,
    "last_test_result": null,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**注意：** 请求体中的敏感字段（如密码、Token、Secret）在存储时由服务端自动加密，响应中不返回敏感字段的明文。

#### 5.3.2 查询通知渠道列表

```
GET /api/v1/notification-channels
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| channel_type | string | 否 | 按渠道类型过滤 |
| enabled | boolean | 否 | 按启用状态过滤 |
| keyword | string | 否 | 按名称模糊搜索 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "channel_id": "ch-email-001",
        "name": "公司邮件服务",
        "channel_type": "email",
        "enabled": true,
        "last_test_at": "2026-10-27T09:30:00Z",
        "last_test_result": "success",
        "created_at": "2026-10-27T10:00:00Z"
      },
      {
        "channel_id": "ch-feishu-001",
        "name": "运维飞书群",
        "channel_type": "feishu",
        "enabled": true,
        "last_test_at": "2026-10-26T15:00:00Z",
        "last_test_result": "success",
        "created_at": "2026-10-25T08:00:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询单个通知渠道

```
GET /api/v1/notification-channels/{channel_id}
```

**响应：** 返回渠道详情，敏感字段以掩码形式展示（如 `***word`）。

#### 5.3.4 更新通知渠道

```
PUT /api/v1/notification-channels/{channel_id}
```

**请求体：** 同创建，字段可选。

#### 5.3.5 删除通知渠道

```
DELETE /api/v1/notification-channels/{channel_id}
```

**前置校验：** 检查是否有 user_notification_target 引用该渠道类型。若有用户配置了该类型的通知目标，提示先清理。

#### 5.3.6 启用/禁用渠道

```
POST /api/v1/notification-channels/{channel_id}/enable
POST /api/v1/notification-channels/{channel_id}/disable
```

#### 5.3.7 连通性测试

```
POST /api/v1/notification-channels/{channel_id}/test
```

**请求体（可选）：**

```json
{
  "target_value": "test@example.com",
  "message": "这是一条测试告警通知"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "channel_id": "ch-email-001",
    "test_status": "success",
    "test_message": "测试邮件已发送至 test@example.com",
    "latency_ms": 1250,
    "tested_at": "2026-10-27T10:30:00Z"
  }
}
```

**行为：**

1. 向指定目标发送测试消息。
2. 记录 `last_test_at` 和 `last_test_result`（success / failed / timeout）。
3. 若测试失败，返回错误详情（如 SMTP 连接超时、Token 无效等）。

#### 5.3.8 用户通知目标管理

```
# 添加用户通知目标
POST /api/v1/users/{user_id}/notification-targets
{
  "channel_type": "email",
  "target_value": "zhangsan@company.com",
  "is_primary": true
}

# 查询用户的通知目标列表
GET /api/v1/users/{user_id}/notification-targets

# 更新通知目标
PUT /api/v1/users/{user_id}/notification-targets/{target_id}
{
  "is_primary": false
}

# 删除通知目标
DELETE /api/v1/users/{user_id}/notification-targets/{target_id}
```

#### 5.3.9 通知投递接口（内部调用）

```
# 投递告警通知
POST /api/v1/notifications/dispatch
```

**请求体：**

```json
{
  "event_id": 12345,
  "recipients": [
    {
      "user_id": "user-001",
      "targets": [
        { "channel_type": "feishu", "target_value": "zhangsan_open_id" }
      ]
    }
  ],
  "notification": {
    "title": "[P1] MySQL连接数过高",
    "content": "实例 inst-mysql-001 当前连接数 480，超过阈值 500 的 90%",
    "severity": "critical",
    "zone_id": "zone-east-1",
    "instance_id": "inst-mysql-001",
    "starts_at": "2026-10-27T10:25:00Z",
    "alert_url": "https://monitor.company.com/alerts/12345"
  }
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "dispatch_id": "dispatch-001",
    "total_recipients": 1,
    "results": [
      {
        "user_id": "user-001",
        "channel_type": "feishu",
        "status": "sent",
        "delivery_id": 1001
      }
    ]
  }
}
```

#### 5.3.10 投递记录查询

```
GET /api/v1/notifications/deliveries
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| event_id | long | 否 | 按告警事件 ID 过滤 |
| status | string | 否 | 按投递状态过滤：pending/sent/failed/retrying |
| channel_type | string | 否 | 按渠道类型过滤 |
| user_id | string | 否 | 按接收用户过滤 |
| start_time | timestamp | 否 | 起始时间 |
| end_time | timestamp | 否 | 结束时间 |

---

## 六、设计决策

### 6.1 NotificationDispatcher 接口抽象

**决策：** 定义渠道无关的 `NotificationDispatcher` 接口，各渠道类型通过适配器模式实现。

```go
// NotificationDispatcher 通知投递抽象接口
type NotificationDispatcher interface {
    // Dispatch 投递通知到指定渠道
    Dispatch(ctx context.Context, req *DispatchRequest) (*DispatchResult, error)

    // BatchDispatch 批量投递通知
    BatchDispatch(ctx context.Context, reqs []*DispatchRequest) ([]*DispatchResult, error)

    // Test 测试渠道连通性
    Test(ctx context.Context, channelID string, target string) (*TestResult, error)

    // SupportedTypes 返回支持的渠道类型列表
    SupportedTypes() []string
}

// DispatchRequest 投递请求
type DispatchRequest struct {
    EventID     int64             `json:"event_id"`
    ChannelType string            `json:"channel_type"`
    Target      string            `json:"target"`
    Title       string            `json:"title"`
    Content     string            `json:"content"`
    Severity    string            `json:"severity"`
    Extra       map[string]string `json:"extra"`
}

// DispatchResult 投递结果
type DispatchResult struct {
    UserID       string    `json:"user_id"`
    ChannelType  string    `json:"channel_type"`
    Status       string    `json:"status"`
    DeliveryID   int64     `json:"delivery_id"`
    ErrorMessage string    `json:"error_message,omitempty"`
}
```

**渠道适配器接口：**

```go
// ChannelSender 渠道发送器
type ChannelSender interface {
    // Send 发送通知
    Send(ctx context.Context, config json.RawMessage, target string, message *NotificationMessage) error

    // TestConnectivity 测试连通性
    TestConnectivity(ctx context.Context, config json.RawMessage, target string) error

    // Type 渠道类型标识
    Type() string
}
```

**理由：**

- 接口抽象使得新增渠道类型只需实现 `ChannelSender` 接口并注册，无需修改调度逻辑。
- `NotificationDispatcher` 统一处理重试、限流、记录等横切关注点。
- 可替换设计：用户可实现自定义的 Dispatcher（如通过 RabbitMQ 异步投递）。

### 6.2 敏感字段加密策略

**决策：** 渠道配置中的敏感字段（密码、Token、Secret）在写入数据库前自动加密，读取时解密。

**实现方式：**

- 使用 AES-256-GCM 对称加密。
- 加密密钥由系统配置管理（如环境变量或 Vault），不存储在数据库中。
- 渠道配置 JSON 中的敏感字段通过预定义规则识别：

```go
var sensitiveFields = map[string][]string{
    "email":    {"smtp_password"},
    "sms":      {"access_key_id", "access_key_secret"},
    "feishu":   {"secret"},
    "dingtalk": {"secret"},
    "wechat":   {},
    "webhook":  {"signature_secret"},
}
```

- API 响应中敏感字段以掩码展示（如 `***word`），仅展示首尾字符。

**理由：**

- 防止数据库泄露时敏感信息暴露。
- 掩码展示避免前端意外泄露配置。
- 加密规则与渠道类型绑定，便于维护和审计。

### 6.3 连通性测试机制

**决策：** 每个渠道支持独立的连通性测试，测试结果记录到 `last_test_at` 和 `last_test_result`。

**测试流程：**

```
用户触发测试
    │
    ▼
从数据库读取渠道配置（解密敏感字段）
    │
    ▼
构造测试消息（包含时间戳标识）
    │
    ▼
调用对应 ChannelSender.TestConnectivity
    │
    ├── 邮件：发送测试邮件到指定地址
    ├── 短信：发送测试短信到指定手机号
    ├── 飞书/钉钉/微信：发送测试卡片消息到 Webhook
    └── Webhook：发送 POST 请求到目标 URL
    │
    ▼
记录测试结果（success/failed/timeout）
    │
    ▼
更新 last_test_at 和 last_test_result
```

**测试超时设置：**

| 渠道类型 | 默认超时 | 说明 |
|----------|----------|------|
| email | 10s | SMTP 连接 + 发送 |
| sms | 10s | API 调用 |
| feishu | 5s | Webhook POST |
| dingtalk | 5s | Webhook POST |
| wechat | 5s | Webhook POST |
| webhook | 15s | 可自定义 |

### 6.4 投递重试与退避策略

**决策：** 投递失败时自动重试，采用指数退避策略。

```
第 1 次重试：等待 30 秒
第 2 次重试：等待 2 分钟
第 3 次重试：等待 10 分钟
超过最大重试次数：标记为 failed，记录错误信息
```

**最大重试次数：** 默认 3 次，可在渠道配置中自定义（如 webhook 的 `retry_count`）。

**理由：**

- 网络抖动或临时故障通常可在短时间内恢复，重试机制提高投递成功率。
- 指数退避避免对目标服务造成压力。
- 超过重试上限后标记失败，避免无限重试占用资源。

### 6.5 通知限流（防风暴）

**决策：** 在 NotificationDispatcher 层实现限流，防止告警风暴时的大量通知轰炸。

**限流策略：**

| 维度 | 限制 | 说明 |
|------|------|------|
| 每用户每分钟 | 最多 10 条 | 防止单个用户被通知淹没 |
| 每渠道每分钟 | 最多 100 条 | 防止渠道服务过载或被封禁 |
| 全局每秒 | 最多 50 条 | 系统整体投递速率上限 |

**超限时行为：**

- 合并相似告警为一条汇总通知。
- 延迟非紧急告警（severity=info）的投递。
- 记录限流日志，便于事后分析。

### 6.6 Webhook 签名验证

**决策：** Webhook 渠道支持 HMAC-SHA256 签名验证，确保投递消息的完整性和来源可信。

**签名算法：**

```
signature = HMAC-SHA256(secret, timestamp + "." + body)
```

**签名通过 `X-Signature` Header 传递（Header 名称可在渠道配置中自定义）。**

**理由：**

- Webhook 接收方可通过校验签名确认消息来源。
- 防止中间人篡改告警内容。
- 签名密钥（signature_secret）属于敏感字段，加密存储。
