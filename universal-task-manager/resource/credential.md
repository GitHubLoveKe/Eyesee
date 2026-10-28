# M4 凭据管理

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：M4 | 职责：凭据的加密存储、引用分发、版本管理与过期管理

---

## 一、概述

凭据（Credential）是实例访问被监控资源所需的认证信息，包括用户名密码、SSH 密钥、SNMP Community、Bearer Token、TLS 证书等。凭据以加密形式存储在网区级别，实例通过 `credential_id` 引用凭据，自身不存储任何敏感数据。

本模块承担以下核心职责：

- 多类型凭据的创建、查询、更新与删除
- AES 加密存储凭据数据，确保敏感信息安全
- 引用分发机制：实例通过 ID 引用，不存储实际凭据
- 版本管理：凭据更新时自动递增版本号
- 过期管理：支持设置过期时间，到期前告警提醒

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 凭据 CRUD | 创建、查询、更新、删除凭据 |
| 加密存储 | AES-256-GCM 加密凭据数据 |
| 类型管理 | 支持 basic_auth / ssh_key / snmp / bearer / tls 五种类型 |
| 版本管理 | 凭据更新时自动递增 version 字段 |
| 过期管理 | 设置 expires_at，到期前告警提醒 |
| 引用关系 | 提供凭据被实例引用的查询能力 |
| 连通性验证 | 使用凭据测试目标实例的认证有效性 |

### 本模块不负责

| 职责 | 归属模块 |
|------|----------|
| 实例的具体管理 | M3 实例管理 |
| 网区管理 | M1 网区与网段管理 |
| 采集任务配置中的凭据使用 | T1 采集任务管理 |
| 加密密钥管理（KMS） | 基础设施层 / 部署配置 |

---

## 三、功能清单

| # | 功能 | 说明 | 优先级 |
|---|------|------|--------|
| F1 | 创建凭据 | 指定类型、名称、凭据数据，加密存储 | P1 |
| F2 | 查询凭据列表 | 按网区、类型查询，不返回敏感数据 | P1 |
| F3 | 查询凭据详情 | 返回基本信息（不含解密数据），含引用统计 | P1 |
| F4 | 更新凭据 | 更新凭据数据，自动递增版本号 | P1 |
| F5 | 删除凭据 | 需确保无实例引用 | P1 |
| F6 | 凭据类型定义 | basic_auth / ssh_key / snmp / bearer / tls | P1 |
| F7 | AES 加密存储 | 使用 AES-256-GCM 加密 encrypted_data | P1 |
| F8 | 版本管理 | 更新凭据时自动递增 version 字段 | P2 |
| F9 | 过期管理 | 设置 expires_at，到期前告警提醒 | P2 |
| F10 | 引用查询 | 查询凭据被哪些实例引用 | P1 |
| F11 | 连通性测试 | 使用凭据测试目标实例的认证 | P2 |
| F12 | 凭据轮转 | 支持批量更新同一类型的凭据 | P3 |

---

## 四、核心数据模型

### 4.1 credential 表

```sql
CREATE TABLE credential (
    credential_id   VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL COMMENT '所属网区ID',
    name            VARCHAR(128) NOT NULL COMMENT '凭据名称',
    cred_type       VARCHAR(32) NOT NULL COMMENT 'basic_auth/ssh_key/snmp/bearer/tls',
    encrypted_data  TEXT NOT NULL COMMENT 'AES-256-GCM加密后的凭据数据',
    version         INT DEFAULT 1 COMMENT '凭据版本号',
    expires_at      TIMESTAMP NULL COMMENT '过期时间',
    description     TEXT COMMENT '凭据描述',
    created_by      VARCHAR(64) COMMENT '创建人',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zone(zone_id)
);

CREATE INDEX idx_cred_zone ON credential(zone_id);
CREATE INDEX idx_cred_type ON credential(cred_type);
CREATE INDEX idx_cred_expires ON credential(expires_at);
```

### 4.2 凭据类型与数据结构

每种凭据类型的 `encrypted_data` 解密后的 JSON 结构如下：

#### basic_auth（用户名密码）

```json
{
  "username": "monitor",
  "password": "***"
}
```

#### ssh_key（SSH 密钥）

```json
{
  "username": "root",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
  "passphrase": ""
}
```

#### snmp（SNMP 社区名）

```json
{
  "version": "v2c",
  "community": "public",
  "security_level": "noAuthNoPriv",
  "auth_protocol": "",
  "auth_password": "",
  "priv_protocol": "",
  "priv_password": ""
}
```

#### bearer（Bearer Token）

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### tls（TLS 证书）

```json
{
  "ca_cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "client_cert": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
  "client_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
  "insecure_skip_verify": false
}
```

### 4.3 关联关系

```
zone (1) ──< (N) credential
                  │
                  └──< (N) instance (通过 credential_id 引用)
                         │
                         └── 采集任务使用凭据进行认证
```

---

## 五、接口与交互

### 5.1 上游依赖

| 依赖模块 | 依赖内容 | 交互方式 |
|----------|----------|----------|
| M1 网区与网段 | zone_id 校验、网区存在性确认 | 内部 API 调用 |
| 基础设施层 | AES 加密密钥管理（KMS 或配置文件） | 配置加载 |

### 5.2 下游提供

| 消费模块 | 提供内容 | 交互方式 |
|----------|----------|----------|
| M3 实例管理 | 凭据引用关系、凭据有效性状态 | 查询 API |
| T1 采集任务 | 凭据解密数据（用于生成采集器认证配置） | 内部 API 调用 |
| 采集器适配器 | 凭据数据（注入到采集器配置模板中） | 内部 API 调用 |

### 5.3 对外 API

#### 5.3.1 创建凭据

```
POST /api/v1/zones/{zone_id}/credentials
```

**请求体（basic_auth 示例）：**

```json
{
  "credential_id": "cred-mysql-001",
  "name": "MySQL 生产凭据",
  "cred_type": "basic_auth",
  "description": "MySQL 生产环境监控账号认证",
  "data": {
    "username": "monitor",
    "password": "***"
  },
  "expires_at": "2027-10-27T00:00:00Z"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "credential_id": "cred-mysql-001",
    "zone_id": "zone-east-1",
    "name": "MySQL 生产凭据",
    "cred_type": "basic_auth",
    "version": 1,
    "expires_at": "2027-10-27T00:00:00Z",
    "referenced_by": 0,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

**注意：** 请求体中的 `data` 字段为明文凭据数据，服务端加密后存储为 `encrypted_data`。响应中不返回凭据明文。

#### 5.3.2 查询凭据列表

```
GET /api/v1/zones/{zone_id}/credentials
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cred_type | string | 否 | 按凭据类型过滤 |
| keyword | string | 否 | 按名称/描述搜索 |
| expired | boolean | 否 | 过滤是否已过期 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "credential_id": "cred-mysql-001",
        "name": "MySQL 生产凭据",
        "cred_type": "basic_auth",
        "version": 2,
        "expires_at": "2027-10-27T00:00:00Z",
        "is_expired": false,
        "referenced_by": 5,
        "description": "MySQL 生产环境监控账号认证",
        "created_at": "2026-10-27T10:00:00Z",
        "updated_at": "2026-10-27T15:00:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询凭据详情

```
GET /api/v1/credentials/{credential_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "credential_id": "cred-mysql-001",
    "zone_id": "zone-east-1",
    "zone_name": "华东区",
    "name": "MySQL 生产凭据",
    "cred_type": "basic_auth",
    "description": "MySQL 生产环境监控账号认证",
    "version": 2,
    "expires_at": "2027-10-27T00:00:00Z",
    "is_expired": false,
    "days_until_expiry": 365,
    "references": [
      {
        "instance_id": "inst-mysql-prod-001",
        "instance_name": "MySQL 生产主库",
        "host": "10.0.1.5"
      },
      {
        "instance_id": "inst-mysql-prod-002",
        "instance_name": "MySQL 生产从库",
        "host": "10.0.1.6"
      }
    ],
    "referenced_by": 2,
    "created_by": "user-001",
    "created_at": "2026-10-27T10:00:00Z",
    "updated_at": "2026-10-27T15:00:00Z"
  }
}
```

**注意：** 详情接口不返回解密后的凭据数据（`data` 字段），仅返回元信息与引用关系。

#### 5.3.4 更新凭据

```
PUT /api/v1/credentials/{credential_id}
```

**请求体：**

```json
{
  "name": "MySQL 生产凭据（更新）",
  "description": "更新后的描述",
  "data": {
    "username": "monitor_v2",
    "password": "***"
  },
  "expires_at": "2028-10-27T00:00:00Z"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "credential_id": "cred-mysql-001",
    "version": 3,
    "updated_at": "2026-10-27T16:00:00Z"
  }
}
```

**说明：** 更新凭据数据时，`version` 字段自动递增。所有引用该凭据的实例在下次采集时将使用新凭据。

#### 5.3.5 删除凭据

```
DELETE /api/v1/credentials/{credential_id}
```

**前置校验：**
- 凭据未被任何实例引用（referenced_by = 0）
- 若存在引用，需先解除实例关联或返回错误提示

**错误响应（存在引用时）：**

```json
{
  "code": 42201,
  "message": "凭据正在被 2 个实例引用，无法删除",
  "details": {
    "referenced_instances": ["inst-mysql-prod-001", "inst-mysql-prod-002"]
  }
}
```

#### 5.3.6 查询凭据引用

```
GET /api/v1/credentials/{credential_id}/references
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "credential_id": "cred-mysql-001",
    "references": [
      {
        "instance_id": "inst-mysql-prod-001",
        "instance_name": "MySQL 生产主库",
        "zone_id": "zone-east-1",
        "host": "10.0.1.5",
        "status": "active"
      }
    ],
    "total": 1
  }
}
```

#### 5.3.7 连通性测试

```
POST /api/v1/credentials/{credential_id}/test
```

**请求体：**

```json
{
  "target_host": "10.0.1.5",
  "target_port": 3306,
  "timeout": "10s"
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "test_id": "test-cred-001",
    "status": "passed",
    "latency_ms": 15,
    "details": {
      "connectivity": "ok",
      "authentication": "ok"
    },
    "tested_at": "2026-10-27T10:10:00Z"
  }
}
```

#### 5.3.8 查询即将过期的凭据

```
GET /api/v1/credentials/expiring
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| within_days | int | 否 | 查询多少天内过期的凭据，默认 30 |
| zone_id | string | 否 | 限定网区范围 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "credential_id": "cred-mysql-001",
        "name": "MySQL 生产凭据",
        "cred_type": "basic_auth",
        "zone_id": "zone-east-1",
        "zone_name": "华东区",
        "expires_at": "2026-11-15T00:00:00Z",
        "days_until_expiry": 19,
        "referenced_by": 5
      }
    ],
    "total": 1
  }
}
```

---

## 六、设计决策

### 6.1 加密方案：AES-256-GCM

**决策：** 使用 AES-256-GCM 对称加密算法保护凭据数据。

**理由：**
- AES-256 是当前业界公认的安全对称加密标准
- GCM 模式同时提供加密与完整性校验（AEAD），防止数据篡改
- 加密密钥通过环境变量或 KMS 注入，不存储在数据库中
- Go/Java/Python 等主流语言均有标准库支持

### 6.2 凭据与实例的引用关系设计

**决策：** 实例通过 `credential_id` 外键引用凭据，凭据不存储实例列表（反向查询通过索引实现）。

**理由：**
- 引用关系单向（实例 → 凭据），避免数据冗余
- 凭据更新时，所有引用实例自动使用新凭据，无需逐一更新
- 查询凭据的引用列表通过 `instance.credential_id` 索引实现，性能可控
- 删除凭据前检查引用关系，防止悬挂引用

### 6.3 版本管理策略

**决策：** 凭据每次更新自动递增 `version` 字段，不保留历史版本。

**理由：**
- 版本号用于追踪凭据变更频率，辅助安全审计
- 不保留历史版本避免加密数据膨胀和安全风险
- 采集任务可通过版本号判断是否需要重新加载凭据（缓存失效策略）
- 如需回滚，用户需手动重新创建凭据

### 6.4 过期管理机制

**决策：** 凭据支持设置 `expires_at` 过期时间，平台通过定时任务检查并告警。

**理由：**
- 安全合规要求：定期轮转凭据是最佳安全实践
- 过期前 30/7/1 天触发告警通知，提醒管理员更新
- 过期后凭据不自动失效（避免突然中断采集），但标记为 `is_expired: true`
- 管理员可选择在过期后继续使用或强制禁用

### 6.5 API 安全设计

**决策：** 凭据 API 遵循最小暴露原则——列表和详情接口不返回解密数据，仅创建/更新接口接收明文。

**理由：**
- 凭据是高度敏感数据，传输和存储链路需最小化暴露面
- 只有采集器适配器内部可调用解密接口，不对用户开放
- API 响应中的 `data` 字段永远为 null 或不包含，避免日志泄露
- 审计日志记录凭据的访问与变更操作

### 6.6 凭据类型扩展性

**决策：** 通过 `cred_type` 字段区分凭据类型，每种类型有独立的 JSON 数据结构。

**理由：**
- 五种内置类型（basic_auth / ssh_key / snmp / bearer / tls）覆盖 95% 的监控场景
- JSON 结构灵活，不同类型可定义不同的字段
- 未来新增凭据类型只需扩展 `数据结构和 API 校验逻辑
- 前端根据 `cred_type` 动态渲染不同的表单字段
