# I2 用户管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：I2 | 职责：用户全生命周期管理、多源认证、角色权限、通知目标管理

---

## 一、概述

用户管理模块是基础设施层的身份与权限中枢，负责管理平台所有用户的注册、认证、授权与通知配置。系统支持多源认证（local / LDAP / OAuth / OIDC），通过 user_auth_source 表实现一个用户绑定多种认证方式。模块同时管理用户的通知目标（邮件、短信、飞书、钉钉、微信），并与告警组（A2）和业务系统（B1）协作完成告警通知投递与资源权限隔离。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 用户 CRUD | 用户的创建、查询、修改、停用/删除 |
| 多源认证 | local 密码认证、LDAP 绑定认证、OAuth/OIDC 第三方认证 |
| 认证源管理 | 用户与认证源的绑定关系维护 |
| 角色权限 | admin / operator / viewer 三级角色的权限判定 |
| 通知目标管理 | 用户多渠道通知地址的 CRUD |
| 登录会话 | 用户登录态管理、Token 签发与刷新 |

### 本模块不负责

| 不负责项 | 归属模块 |
|----------|----------|
| 告警组的创建与管理 | A2 告警组 |
| 业务系统成员管理 | B1 业务系统管理 |
| 通知渠道的配置与投递 | A3 通知渠道 |
| LDAP 服务器 / OAuth Provider 运维 | 外部系统 |
| 告警事件处理 | A1 告警策略 / A4 告警收敛 |

---

## 三、功能清单

| # | 功能 | 优先级 | 阶段 | 说明 |
|---|------|--------|------|------|
| F1 | 用户创建 | P0 | Phase 4 | 创建用户账号，指定角色与基本信息 |
| F2 | 用户查询 | P0 | Phase 4 | 列表查询（分页/筛选）、详情查询 |
| F3 | 用户信息修改 | P0 | Phase 4 | 修改显示名、邮箱、手机等 |
| F4 | 用户停用/启用 | P0 | Phase 4 | 停用后禁止登录，保留数据 |
| F5 | 用户删除 | P1 | Phase 4 | 软删除，级联清理认证源与通知目标 |
| F6 | 本地密码认证 | P0 | Phase 4 | 用户名 + password_hash 校验 |
| F7 | LDAP 认证 | P1 | Phase 4 | LDAP DN 绑定验证，自动关联/创建用户 |
| F8 | OAuth 认证 | P1 | Phase 4 | OAuth 2.0 授权码流程，subject 关联 |
| F9 | OIDC 认证 | P1 | Phase 4 | OpenID Connect，ID Token 验证 |
| F10 | 认证源绑定管理 | P0 | Phase 4 | 为用户添加/移除认证源 |
| F11 | 角色权限判定 | P0 | Phase 4 | admin/operator/viewer 权限校验 |
| F12 | 通知目标管理 | P1 | Phase 4 | 用户通知地址 CRUD，设置主通知渠道 |
| F13 | 登录/登出 | P0 | Phase 4 | 登录签发 Token，登出失效 Token |
| F14 | 密码修改 | P0 | Phase 4 | 本地用户修改密码 |
| F15 | 密码重置 | P1 | Phase 4 | 管理员重置用户密码 |

---

## 四、核心数据模型

### 4.1 user（用户表）

> 引用 data-model.md 5.2 节定义

```sql
CREATE TABLE user (
    user_id         VARCHAR(64) PRIMARY KEY,
    username        VARCHAR(128) NOT NULL UNIQUE,
    display_name    VARCHAR(256),
    email           VARCHAR(256),
    phone           VARCHAR(32),
    role            VARCHAR(32) DEFAULT 'viewer' COMMENT 'admin/operator/viewer',
    status          VARCHAR(32) DEFAULT 'active' COMMENT 'active/disabled/deleted',
    last_login_at   TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_username ON user(username);
CREATE INDEX idx_user_email ON user(email);
CREATE INDEX idx_user_role ON user(role);
CREATE INDEX idx_user_status ON user(status);
```

### 4.2 user_auth_source（用户认证源表）

> 引用 data-model.md 5.3 节定义

```sql
CREATE TABLE user_auth_source (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL,
    source_type     VARCHAR(32) NOT NULL COMMENT 'local/ldap/oauth/oidc',
    external_id     VARCHAR(256) COMMENT 'LDAP DN / OAuth subject / OIDC sub',
    password_hash   VARCHAR(256) COMMENT '仅local类型使用bcrypt存储',
    extra           JSON COMMENT '扩展信息，如OAuth token、LDAP属性映射',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_source (user_id, source_type)
);

CREATE INDEX idx_uas_user ON user_auth_source(user_id);
CREATE INDEX idx_uas_external ON user_auth_source(source_type, external_id);
```

### 4.3 user_notification_target（用户通知目标表）

> 引用 data-model.md 5.4 节定义

```sql
CREATE TABLE user_notification_target (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL,
    channel_type    VARCHAR(32) NOT NULL COMMENT 'email/sms/feishu/dingtalk/wechat/webhook',
    target_value    VARCHAR(256) NOT NULL COMMENT '具体地址/ID',
    is_primary      BOOLEAN DEFAULT FALSE COMMENT '是否为主通知渠道',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_unt_user ON user_notification_target(user_id);
CREATE INDEX idx_unt_channel ON user_notification_target(channel_type);
```

### 4.4 关联表引用（已在其他模块定义）

```sql
-- alert_group_member (A2 告警组) 关联用户
-- 字段: group_id, user_id, role(member/manager)
-- 告警触发时，通过告警组 -> 成员 -> 通知目标 完成通知投递

-- business_system_member (B1 业务系统) 关联用户
-- 字段: system_id, user_id, role(admin/operator/viewer)
-- 控制用户在业务系统内的操作权限
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| A3 通知渠道 | 通知渠道配置（SMTP/飞书webhook等） | 内部函数调用 |
| B1 业务系统 | 用户-业务系统成员关系 | 事件驱动 |
| A2 告警组 | 用户-告警组成员关系 | 事件驱动 |

### 5.2 下游提供

| 消费方 | 提供内容 | 交互方式 |
|--------|----------|----------|
| 所有模块 | 用户身份校验（认证/鉴权） | 中间件拦截 |
| A3 通知渠道 | 用户通知目标列表 | 内部函数调用 |
| A2 告警组 | 告警组成员列表 | 内部函数调用 |
| B1 业务系统 | 业务系统成员与角色 | 内部函数调用 |
| 前端 | 当前登录用户信息、权限 | REST API |
| I3 状态缓存 | 用户操作审计日志 | 事件驱动 |

### 5.3 对外 API

#### 5.3.1 用户登录

```
POST /api/v1/auth/login
```

**请求体（本地认证）：**

```json
{
  "auth_type": "local",
  "username": "admin",
  "password": "***"
}
```

**请求体（LDAP 认证）：**

```json
{
  "auth_type": "ldap",
  "username": "zhangsan",
  "password": "***"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 7200,
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
    "user": {
      "user_id": "user-001",
      "username": "admin",
      "display_name": "系统管理员",
      "role": "admin"
    }
  }
}
```

#### 5.3.2 OAuth/OIDC 登录

```
GET /api/v1/auth/oauth/{provider}/authorize
```

**流程：**

```
用户 → 点击「GitHub 登录」
     → GET /api/v1/auth/oauth/github/authorize
     → 302 重定向到 GitHub 授权页
     → 用户授权后回调
     → GET /api/v1/auth/oauth/github/callback?code=xxx
     → 平台用 code 换取 access_token
     → 获取用户信息（subject / email）
     → 查找或创建 user_auth_source 关联
     → 签发平台 JWT Token
     → 302 重定向到前端，携带 Token
```

#### 5.3.3 Token 刷新

```
POST /api/v1/auth/refresh
```

**请求体：**

```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g..."
}
```

#### 5.3.4 用户登出

```
POST /api/v1/auth/logout
```

#### 5.3.5 创建用户

```
POST /api/v1/users
```

**权限：** admin

**请求体：**

```json
{
  "username": "zhangsan",
  "display_name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "role": "operator",
  "auth_sources": [
    {
      "source_type": "local",
      "password": "***"
    }
  ],
  "notification_targets": [
    {
      "channel_type": "email",
      "target_value": "zhangsan@example.com",
      "is_primary": true
    },
    {
      "channel_type": "feishu",
      "target_value": "ou_zhangsan_id",
      "is_primary": false
    }
  ]
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "user_id": "user-002",
    "username": "zhangsan",
    "display_name": "张三",
    "role": "operator",
    "status": "active",
    "auth_sources": ["local"],
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.6 查询用户列表

```
GET /api/v1/users
```

**权限：** admin

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| role | string | 按角色过滤：admin/operator/viewer |
| status | string | 按状态过滤：active/disabled |
| keyword | string | 搜索用户名/显示名/邮箱 |
| page | int | 页码 |
| page_size | int | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "user_id": "user-001",
        "username": "admin",
        "display_name": "系统管理员",
        "email": "admin@example.com",
        "role": "admin",
        "status": "active",
        "auth_sources": ["local"],
        "last_login_at": "2026-10-27T09:00:00Z"
      }
    ],
    "total": 15,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.7 查询用户详情

```
GET /api/v1/users/{user_id}
```

**权限：** admin 或本人

**响应：**

```json
{
  "code": 0,
  "data": {
    "user_id": "user-002",
    "username": "zhangsan",
    "display_name": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "role": "operator",
    "status": "active",
    "last_login_at": "2026-10-27T08:30:00Z",
    "auth_sources": [
      {
        "source_type": "local",
        "external_id": null,
        "created_at": "2026-10-27T10:00:00Z"
      },
      {
        "source_type": "ldap",
        "external_id": "cn=zhangsan,ou=people,dc=example,dc=com",
        "created_at": "2026-10-27T10:05:00Z"
      }
    ],
    "notification_targets": [
      {
        "id": 1,
        "channel_type": "email",
        "target_value": "zhangsan@example.com",
        "is_primary": true
      },
      {
        "id": 2,
        "channel_type": "feishu",
        "target_value": "ou_zhangsan_id",
        "is_primary": false
      }
    ],
    "business_systems": [
      {"system_id": "bs-001", "name": "电商平台", "role": "operator"},
      {"system_id": "bs-002", "name": "支付系统", "role": "viewer"}
    ],
    "alert_groups": [
      {"group_id": "ag-001", "name": "DBA值班组", "role": "member"}
    ],
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T10:05:00Z"
  }
}
```

#### 5.3.8 更新用户信息

```
PUT /api/v1/users/{user_id}
```

**权限：** admin 或本人（本人仅可修改 display_name、email、phone）

**请求体：**

```json
{
  "display_name": "张三（已更新）",
  "email": "zhangsan_new@example.com",
  "phone": "13900139000",
  "role": "operator"
}
```

#### 5.3.9 停用/启用用户

```
PUT /api/v1/users/{user_id}/status
```

**权限：** admin

**请求体：**

```json
{
  "status": "disabled"
}
```

#### 5.3.10 删除用户

```
DELETE /api/v1/users/{user_id}
```

**权限：** admin

**行为：** 软删除（status 标记为 deleted），级联删除 user_auth_source 和 user_notification_target 记录，同时从所有 alert_group_member 和 business_system_member 中移除。

#### 5.3.11 修改密码

```
PUT /api/v1/users/{user_id}/password
```

**权限：** 本人（需提供旧密码）或 admin（可强制重置）

**请求体：**

```json
{
  "old_password": "***",
  "new_password": "***"
}
```

#### 5.3.12 管理认证源

```
POST /api/v1/users/{user_id}/auth-sources
```

**权限：** admin 或本人

**请求体（绑定 LDAP）：**

```json
{
  "source_type": "ldap",
  "external_id": "cn=zhangsan,ou=people,dc=example,dc=com",
  "extra": {"ldap_server": "ldap://ldap.example.com"}
}
```

```
DELETE /api/v1/users/{user_id}/auth-sources/{source_type}
```

#### 5.3.13 管理通知目标

```
POST /api/v1/users/{user_id}/notification-targets
```

**请求体：**

```json
{
  "channel_type": "dingtalk",
  "target_value": "dingtalk_user_id",
  "is_primary": false
}
```

```
GET /api/v1/users/{user_id}/notification-targets
PUT /api/v1/users/{user_id}/notification-targets/{id}
DELETE /api/v1/users/{user_id}/notification-targets/{id}
```

#### 5.3.14 获取当前登录用户

```
GET /api/v1/auth/me
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "user_id": "user-002",
    "username": "zhangsan",
    "display_name": "张三",
    "email": "zhangsan@example.com",
    "role": "operator",
    "permissions": ["task:create", "task:update", "instance:view", "zone:view"]
  }
}
```

---

## 六、设计决策

### 6.1 多源认证架构

**决策：** 采用 user + user_auth_source 分离模型，一个用户可绑定多种认证方式。

```
user (user_id=001, username=zhangsan)
  ├── user_auth_source (source_type=local, password_hash=$2b$12$xxx)
  ├── user_auth_source (source_type=ldap, external_id=cn=zhangsan,...)
  └── user_auth_source (source_type=oauth, external_id=github:12345)
```

**理由：**
- 用户身份与认证方式解耦，同一用户可同时支持密码登录和 SSO 登录
- 认证源可独立扩展，新增认证方式不影响用户核心数据
- `uk_user_source (user_id, source_type)` 约束保证每种认证方式唯一绑定

**认证流程：**

| 认证方式 | 流程 |
|----------|------|
| local | 用户名查找 user → 查 user_auth_source(local) → bcrypt 校验 password_hash |
| ldap | 用户名查找 user → 若无则自动创建 → 用 LDAP DN + 密码绑定验证 → 关联 user_auth_source(ldap) |
| oauth | 授权码换 token → 获取 subject → 查 user_auth_source(oauth, external_id) → 关联 user |
| oidc | 同 oauth，额外验证 ID Token 签名与 issuer |

### 6.2 角色权限模型

**决策：** 采用 RBAC（基于角色的访问控制），三级预定义角色。

| 角色 | 权限范围 | 典型用户 |
|------|----------|----------|
| **admin** | 全局管理：用户管理、系统配置、所有资源的 CRUD | 平台管理员 |
| **operator** | 操作权限：采集任务/实例/网区的 CRUD，不可管理用户和系统配置 | 运维工程师 |
| **viewer** | 只读权限：查看所有资源与告警，不可修改 | 业务方/管理层 |

**权限与业务系统的交叉：**

```
全局角色: admin
  └── 拥有所有业务系统的 admin 权限

全局角色: operator
  ├── 业务系统 B1: operator (可操作 B1 内资源)
  └── 业务系统 B2: viewer (只读 B2 内资源)

全局角色: viewer
  └── 所有业务系统均为 viewer
```

**权限校验顺序：**

1. 全局 admin → 直接放行所有操作
2. 全局 operator/viewer → 检查业务系统级角色（business_system_member.role）
3. 业务系统 admin → 可管理该系统内所有资源
4. 业务系统 operator → 可创建/修改任务，不可管理成员
5. 业务系统 viewer → 仅可查看

### 6.3 密码安全策略

| 策略 | 说明 |
|------|------|
| 哈希算法 | bcrypt（cost=12） |
| 密码长度 | 最少 8 位 |
| 密码复杂度 | 至少包含大小写字母和数字（推荐包含特殊字符） |
| 登录失败锁定 | 连续 5 次失败后锁定 15 分钟 |
| Token 有效期 | access_token: 2h, refresh_token: 7d |

### 6.4 通知目标与告警投递

**投递链路：**

```
告警触发 (A1 告警策略)
    │
    ▼
告警组 (A2 alert_group)
    │
    ▼
告警组成员 (alert_group_member → user_id)
    │
    ▼
用户通知目标 (user_notification_target)
    │
    ├── is_primary=true → 主渠道优先投递
    │
    └── is_primary=false → 备用渠道投递（主渠道失败时）
    │
    ▼
通知渠道 (A3 notification_channel)
    │
    ├── email → SMTP 发送
    ├── sms → 短信网关
    ├── feishu → 飞书机器人 webhook
    ├── dingtalk → 钉钉机器人 webhook
    ├── wechat → 企业微信 webhook
    └── webhook → 自定义 HTTP 回调
```

**通知目标去重：** 同一用户在同一告警事件中只接收一次通知，即使该用户属于多个告警组。

### 6.5 与告警组的关系

```
alert_group (告警组)
  ├── alert_group_member
  │     ├── user_id → user (关联用户)
  │     └── role: member / manager
  │
  └── alert_strategy (关联告警策略)
        └── 告警触发时通知该组所有成员
```

- 用户可属于多个告警组
- 告警组 manager 可管理组成员（添加/移除用户）
- 告警组支持升级链（parent_id），超时未认领则升级到父组

### 6.6 与业务系统的关系

```
business_system (业务系统)
  └── business_system_member
        ├── user_id → user (关联用户)
        └── role: admin / operator / viewer
```

- 用户可属于多个业务系统，在不同系统中角色可不同
- 业务系统 admin 可管理该系统的成员
- 资源（网区、实例、任务）通过业务系统隔离，用户只能看到有权限的业务系统内的资源

### 6.7 LDAP 集成设计

**配置模型：**

```json
{
  "ldap": {
    "server": "ldap://ldap.example.com:389",
    "base_dn": "dc=example,dc=com",
    "bind_dn": "cn=admin,dc=example,dc=com",
    "bind_password": "***",
    "user_search_filter": "(uid={username})",
    "user_attribute_mapping": {
      "username": "uid",
      "display_name": "cn",
      "email": "mail",
      "phone": "telephoneNumber"
    },
    "auto_create_user": true,
    "default_role": "viewer"
  }
}
```

**首次登录自动创建流程：**

1. 用户以 LDAP 方式登录
2. 平台用 bind_dn 连接 LDAP，搜索用户
3. 找到用户后尝试用用户密码绑定验证
4. 验证通过后检查 user_auth_source 是否已关联
5. 若未关联 → 自动创建 user + user_auth_source(ldap)
6. 属性映射：从 LDAP 属性中提取 display_name、email 等写入 user 表

### 6.8 OAuth/OIDC 集成设计

**支持的 Provider 配置：**

```json
{
  "oauth_providers": [
    {
      "name": "github",
      "type": "oauth",
      "client_id": "xxx",
      "client_secret": "***",
      "authorize_url": "https://github.com/login/oauth/authorize",
      "token_url": "https://github.com/login/oauth/access_token",
      "userinfo_url": "https://api.github.com/user",
      "scopes": ["user:email"],
      "subject_field": "id",
      "attribute_mapping": {
        "username": "login",
        "display_name": "name",
        "email": "email"
      }
    },
    {
      "name": "keycloak",
      "type": "oidc",
      "issuer": "https://keycloak.example.com/realms/master",
      "client_id": "xxx",
      "client_secret": "***",
      "scopes": ["openid", "profile", "email"],
      "subject_field": "sub"
    }
  ]
}
```
