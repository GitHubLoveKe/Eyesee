# A2 告警组模块

> 版本：v1.0 | 日期：2026-10-27
> 模块编号：A2 | 职责：管理告警分组与升级链，维护组成员关系，提供告警从策略到通知接收人的路由能力。

---

## 一、概述

告警组模块是告警层的路由中枢。它定义了告警的分组结构和升级链，将告警策略（AlertStrategy）产生的告警事件路由到正确的通知接收人。每个告警策略通过 `alert_group_id` 关联到一个告警组，告警组内包含若干成员（member / manager），成员通过通知渠道（A3 模块）接收告警通知。

升级链机制（`parent_id` + `escalation_minutes`）确保告警在规定时间未被响应时自动升级到上级组，避免关键告警被遗漏。

---

## 二、职责边界

### 本模块负责

| 职责 | 说明 |
|------|------|
| 告警组 CRUD | 创建、查询、更新、删除告警组 |
| 升级链管理 | 维护 parent_id 指向的父子关系与 escalation_minutes 升级超时 |
| 组成员管理 | 添加/移除成员，区分 member 与 manager 角色 |
| 告警路由查询 | 根据策略 ID 解析出通知接收人列表 |
| 升级超时检测 | 检测未响应的告警是否达到升级条件 |

### 本模块不负责

| 职责 | 归属 |
|------|------|
| 告警策略的定义与维护 | A1 告警策略模块 |
| 通知的实际投递 | A3 通知渠道模块（NotificationDispatcher） |
| 告警事件的生成与收敛 | A4 告警收敛模块（AlertEngine） |
| 用户账号与认证管理 | 基础设施层（User / UserAuthSource） |
| 用户通知目标配置 | 基础设施层（UserNotificationTarget） |

---

## 三、功能清单

| 功能编号 | 功能名称 | 优先级 | 说明 |
|----------|----------|--------|------|
| A2-F01 | 告警组创建 | P0 | 创建告警组，可指定父组建立升级链 |
| A2-F02 | 告警组查询 | P0 | 支持按名称、父组等条件查询，返回组树结构 |
| A2-F03 | 告警组更新 | P0 | 修改组信息，包括调整升级链关系 |
| A2-F04 | 告警组删除 | P0 | 删除告警组，需检查策略引用与子组依赖 |
| A2-F05 | 添加组成员 | P0 | 将用户添加为组成员，指定角色（member/manager） |
| A2-F06 | 移除组成员 | P0 | 从组中移除用户 |
| A2-F07 | 成员角色变更 | P1 | 变更成员角色（member ↔ manager） |
| A2-F08 | 升级链查询 | P0 | 查询组的升级链路（向上追溯至根组） |
| A2-F09 | 升级超时检测 | P1 | 定时检查未响应告警是否达到 escalation_minutes |
| A2-F10 | 路由解析 | P0 | 根据策略 ID 解析完整的通知接收人列表（含升级链） |
| A2-F11 | 组树展示 | P2 | 以树形结构展示所有告警组的层级关系 |
| A2-F12 | 组成员批量操作 | P2 | 批量添加/移除组成员 |

---

## 四、核心数据模型

### 4.1 alert_group（告警组表）

> 引用 data-model.md 4.3 节定义。

```sql
CREATE TABLE alert_group (
    group_id        VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    parent_id       VARCHAR(64) COMMENT '父组ID，形成升级链',
    escalation_minutes INT COMMENT '升级超时时间（分钟），告警在此时间内未被认领则升级到父组',
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES alert_group(group_id)
);
```

**关键字段说明：**

| 字段 | 说明 |
|------|------|
| `parent_id` | 指向父告警组的 group_id，形成链式升级结构。为 NULL 表示顶层组（最终升级目标） |
| `escalation_minutes` | 升级超时时间。当告警分发到本组后，在指定分钟内无人认领（Acknowledge），则自动升级到 parent_id 指向的父组 |

**升级链示例：**

```
[L1 值班组] (escalation_minutes=15)
    ↑ parent_id
[L2 技术主管组] (escalation_minutes=30)
    ↑ parent_id
[L3 运维总监组] (escalation_minutes=NULL, 顶层)
```

### 4.2 alert_group_member（告警组成员表）

> 引用 data-model.md 4.4 节定义。

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

**角色说明：**

| 角色 | 权限 | 说明 |
|------|------|------|
| `member` | 接收告警通知 | 普通成员，接收组内告警通知，可认领告警 |
| `manager` | 接收通知 + 管理组 | 组管理员，除 member 权限外，可管理组成员、调整组配置 |

### 4.3 关联模型引用

| 表名 | 引用章节 | 用途 |
|------|----------|------|
| `alert_strategy` | data-model.md 4.1 | 策略通过 alert_group_id 关联告警组 |
| `user` | data-model.md 5.2 | 组成员引用的用户实体 |
| `user_notification_target` | data-model.md 5.4 | 用户的通知目标（邮箱、手机、IM 账号等） |
| `alert_event` | data-model.md 4.6 | 告警事件，路由结果决定通知哪些人 |

---

## 五、接口与交互

### 5.1 上游依赖

| 上游模块 | 依赖接口 | 用途 |
|----------|----------|------|
| A1 告警策略模块 | alert_strategy.alert_group_id | 策略关联告警组，路由查询的入口 |
| 基础设施层 | User 查询 | 获取组成员的用户信息 |
| 基础设施层 | UserNotificationTarget 查询 | 获取成员的通知目标（邮箱、手机等） |

### 5.2 下游提供

| 下游模块 | 提供接口 | 用途 |
|----------|----------|------|
| A3 通知渠道模块 | 通知接收人列表 | 根据告警事件解析出需要通知的用户及其通知目标 |
| A4 告警收敛模块 | 升级链查询 | 告警认领/升级时查询升级链路 |
| 前端展示 | 组树结构 | 展示告警组层级关系与成员分布 |

### 5.3 对外 API

#### 5.3.1 创建告警组

```
POST /api/v1/alert-groups
```

**请求体：**

```json
{
  "name": "DBA值班组",
  "description": "数据库管理员值班组，负责MySQL/Redis等数据库告警的一线响应",
  "parent_id": "group-tech-lead",
  "escalation_minutes": 15
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "group_id": "group-dba",
    "name": "DBA值班组",
    "parent_id": "group-tech-lead",
    "escalation_minutes": 15,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.2 查询告警组列表

```
GET /api/v1/alert-groups
```

**查询参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| parent_id | string | 否 | 按父组过滤 |
| keyword | string | 否 | 按名称/描述模糊搜索 |
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页数量 |

**响应：**

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "group_id": "group-dba",
        "name": "DBA值班组",
        "description": "数据库管理员值班组",
        "parent_id": "group-tech-lead",
        "escalation_minutes": 15,
        "member_count": 5,
        "strategy_count": 8,
        "created_at": "2026-10-27T10:00:00Z"
      }
    ],
    "total": 6,
    "page": 1,
    "page_size": 20
  }
}
```

#### 5.3.3 查询单个告警组

```
GET /api/v1/alert-groups/{group_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "group_id": "group-dba",
    "name": "DBA值班组",
    "description": "数据库管理员值班组",
    "parent_id": "group-tech-lead",
    "escalation_minutes": 15,
    "members": [
      { "user_id": "user-001", "username": "张三", "role": "manager", "display_name": "张三" },
      { "user_id": "user-002", "username": "李四", "role": "member", "display_name": "李四" }
    ],
    "escalation_chain": [
      { "group_id": "group-dba", "name": "DBA值班组", "escalation_minutes": 15 },
      { "group_id": "group-tech-lead", "name": "技术主管组", "escalation_minutes": 30 },
      { "group_id": "group-ops-director", "name": "运维总监组", "escalation_minutes": null }
    ],
    "strategy_count": 8,
    "created_at": "2026-10-27T10:00:00Z"
  }
}
```

#### 5.3.4 更新告警组

```
PUT /api/v1/alert-groups/{group_id}
```

**请求体：** 同创建，字段可选。

**校验规则：**

- `parent_id` 不能指向自身。
- `parent_id` 不能形成循环引用（A→B→C→A）。
- 修改 `escalation_minutes` 不影响已在处理中的告警。

#### 5.3.5 删除告警组

```
DELETE /api/v1/alert-groups/{group_id}
```

**前置校验：**

1. 检查是否有告警策略引用该组（`alert_strategy.alert_group_id`）。若有，拒绝删除或提示先解绑。
2. 检查是否有子组以该组为 parent_id。若有，拒绝删除或提示先调整子组的 parent_id。

#### 5.3.6 组成员管理

```
# 添加成员
POST /api/v1/alert-groups/{group_id}/members
{
  "user_id": "user-003",
  "role": "member"
}

# 移除成员
DELETE /api/v1/alert-groups/{group_id}/members/{user_id}

# 变更成员角色
PUT /api/v1/alert-groups/{group_id}/members/{user_id}
{
  "role": "manager"
}

# 查询组成员列表
GET /api/v1/alert-groups/{group_id}/members

# 批量添加成员
POST /api/v1/alert-groups/{group_id}/members/batch
{
  "members": [
    { "user_id": "user-004", "role": "member" },
    { "user_id": "user-005", "role": "member" }
  ]
}
```

#### 5.3.7 查询升级链

```
GET /api/v1/alert-groups/{group_id}/escalation-chain
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "chain": [
      {
        "level": 0,
        "group_id": "group-dba",
        "name": "DBA值班组",
        "escalation_minutes": 15
      },
      {
        "level": 1,
        "group_id": "group-tech-lead",
        "name": "技术主管组",
        "escalation_minutes": 30
      },
      {
        "level": 2,
        "group_id": "group-ops-director",
        "name": "运维总监组",
        "escalation_minutes": null
      }
    ]
  }
}
```

#### 5.3.8 查询告警组树

```
GET /api/v1/alert-groups/tree
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "tree": [
      {
        "group_id": "group-ops-director",
        "name": "运维总监组",
        "children": [
          {
            "group_id": "group-tech-lead",
            "name": "技术主管组",
            "children": [
              {
                "group_id": "group-dba",
                "name": "DBA值班组",
                "children": []
              },
              {
                "group_id": "group-network",
                "name": "网络值班组",
                "children": []
              }
            ]
          }
        ]
      }
    ]
  }
}
```

#### 5.3.9 查询策略关联的告警组

```
GET /api/v1/alert-groups/by-strategy/{strategy_id}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "group_id": "group-dba",
    "name": "DBA值班组",
    "strategy_count": 8
  }
}
```

#### 5.3.10 路由解析（内部接口 / 管理接口）

```
POST /api/v1/alert-groups/resolve-recipients
```

**请求体：**

```json
{
  "strategy_id": "strategy-001",
  "include_escalation": true
}
```

**响应：**

```json
{
  "code": 0,
  "data": {
    "primary": {
      "group_id": "group-dba",
      "name": "DBA值班组",
      "recipients": [
        {
          "user_id": "user-001",
          "username": "张三",
          "targets": [
            { "channel_type": "feishu", "target_value": "zhangsan@company.com", "is_primary": true },
            { "channel_type": "email", "target_value": "zhangsan@company.com", "is_primary": false }
          ]
        }
      ]
    },
    "escalation": [
      {
        "group_id": "group-tech-lead",
        "name": "技术主管组",
        "escalation_minutes": 30,
        "recipients": [...]
      }
    ]
  }
}
```

---

## 六、设计决策

### 6.1 升级链数据结构

**决策：** 使用 `parent_id` 单指针形成链式升级结构，而非多叉树或矩阵。

**理由：**

- 告警升级的本质是"一线 → 二线 → 三线"的线性升级路径，链式结构语义清晰。
- 单指针设计简单，查询升级链路只需沿 parent_id 向上遍历，复杂度为 O(depth)，通常 depth < 5。
- 避免了多叉树场景下的并行升级歧义（多个子组同时升级到一个父组时，父组只处理一次）。

**升级链约束：**

- 不允许循环引用：创建/更新时通过遍历检测环路。
- 链深度建议不超过 5 级，过深的升级链会导致通知延迟和职责模糊。
- 顶层组的 `parent_id` 为 NULL，`escalation_minutes` 也为 NULL（不再向上升级）。

### 6.2 升级超时机制

**决策：** 升级超时从告警被分发到当前组时开始计时，以"认领"（Acknowledge）作为停止条件。

**流程：**

```
告警触发
    │
    ▼
分发到 L1 组（DBA值班组），开始计时 15 分钟
    │
    ├── 15 分钟内有人认领 → 停止计时，不再升级
    │
    └── 15 分钟无人认领
            │
            ▼
        升级到 L2 组（技术主管组），开始计时 30 分钟
            │
            ├── 30 分钟内有人认领 → 停止计时
            │
            └── 30 分钟无人认领
                    │
                    ▼
                升级到 L3 组（运维总监组），不再升级
```

**实现方式：**

- 定时任务（如每 1 分钟）扫描 `starts_at + escalation_minutes < now()` 且未认领的告警。
- 升级时记录升级日志，包括升级时间、源组、目标组。
- 升级通知发送到父组成员，并标注"升级告警"标签。

### 6.3 成员角色设计

**决策：** 区分 `member` 和 `manager` 两种角色，而非引入更复杂的 RBAC。

**理由：**

- 告警组的角色需求简单：谁能接收通知（member）、谁能管理组（manager）。
- 避免过度设计，两种角色覆盖了 95% 的组织场景。
- manager 可以管理组成员和组配置，但不能修改关联的告警策略（策略属于 A1 模块的职责）。

### 6.4 路由解析策略

**决策：** 路由解析从策略出发，沿 策略 → 告警组 → 成员 → 通知目标 的链路完成。

**流程：**

```
AlertStrategy
    │
    │ alert_group_id
    ▼
AlertGroup (primary)
    │
    │ alert_group_member
    ▼
User[] (members)
    │
    │ user_notification_target
    ▼
NotificationTarget[] (email, phone, feishu_id, ...)
    │
    │ 若超时未认领，沿升级链向上
    ▼
AlertGroup (parent) → User[] → NotificationTarget[]
```

**理由：**

- 路由链路与数据模型对齐，查询逻辑直观。
- 升级链的路由解析是延迟执行的（只在超时后才解析父组的成员），避免不必要的数据查询。
- 通知目标由用户自行配置（UserNotificationTarget），告警组只负责"通知谁"，不负责"怎么通知"。

### 6.5 删除保护策略

**决策：** 告警组被策略引用或被作为父组时，禁止直接删除。

**理由：**

- 防止误删导致告警路由中断。
- 强制用户先解绑依赖关系，确保数据一致性。
- 提供明确的错误提示，告知用户哪些策略或子组依赖该组。
