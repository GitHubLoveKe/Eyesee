# DCS 项目进度跟踪

> 多网区监控采集管理平台

## 项目信息

- **分支**: feature/phase1-core-skeleton
- **创建日期**: 2026-10-28
- **技术栈**: Go 1.24 + Vue 3 + PostgreSQL 17 + Redis 8

## 基础设施

- [x] PostgreSQL 17 (Podman: dcs-postgres, localhost:5432, user=dcs, password=dcs123, db=dcs)
- [x] Redis 8 (Podman: dcs-redis, localhost:6379)
- [x] 镜像代理: 15dczmo6htbpugrx1i.xuanyuan.run
- [x] 数据卷: dcs-pgdata, dcs-redisdata

## Phase 1 模块进度

| 模块 | 名称 | 后端 | 前端 | 测试 | 备注 |
|------|------|------|------|------|------|
| B1 | 业务系统管理 | ✅ | ✅ | ✅ | CRUD+成员管理 |
| M1 | 网区与网段管理 | ✅ | ✅ | ✅ | CIDR重叠检测 |
| M2 | 资源类型树 | ✅ | ✅ | ✅ | 树形结构CRUD |
| M3 | 实例管理 | ✅ | ✅ | ✅ | 动态过滤+关联 |
| M5 | 数据存储管理 | ✅ | ✅ | ✅ | 健康检测 |
| T1 | 采集任务管理 | ✅ | ✅ | ✅ | 启用/禁用切换 |
| I1 | 采集器管理 | ✅ | ✅ | ✅ | 心跳上报 |
| I3 | 状态缓存 | ✅ | ✅ | ✅ | Redis聚合查询 |

## Phase 2 模块进度

| 模块 | 名称 | 后端 | 前端 | 测试 | 备注 |
|------|------|------|------|------|------|
| I2 | 用户管理 | ✅ | ✅ | ✅ | CRUD+角色管理 |
| A3 | 通知渠道 | ✅ | ✅ | ✅ | 多渠道+连通测试 |
| A2 | 告警组 | ✅ | ✅ | ✅ | 成员管理+升级链 |
| A1 | 告警策略 | ✅ | ✅ | ✅ | 作用域+实例关联 |
| T5 | 定时任务 | ✅ | ✅ | ✅ | 启禁用+执行历史 |

## Phase 3 模块进度

| 模块 | 名称 | 后端 | 前端 | 测试 | 备注 |
|------|------|------|------|------|------|
| T2 | RC任务管理 | ✅ | ✅ | ✅ | 任务CRUD+规则子资源 |
| T3 | 指标管理 | ✅ | ✅ | ✅ | 指标定义+常用查询 |
| M4 | 凭据管理 | ✅ | ✅ | ✅ | 加密存储+引用检查 |
| T4 | 服务发现 | ✅ | - | ✅ | Prometheus http_sd格式 |
| I4 | RC Worker | ✅ | - | ✅ | 执行器管理+心跳 |

## Phase 4 模块进度

| 模块 | 名称 | 后端 | 前端 | 测试 | 备注 |
|------|------|------|------|------|------|
| A4 | 告警收敛 | ✅ | ✅ | ✅ | 事件落库+认领+静默+统计 |
| I2+ | 用户认证增强 | ✅ | ✅ | ✅ | 通知目标管理+密码修改 |
| I1+ | 协调器集成 | ✅ | ✅ | ✅ | 区协调器CRUD+心跳 |
| B1+ | 业务概览 | ✅ | - | ✅ | 按业务系统聚合仪表盘 |
| T4+ | 服务发现配置 | ✅ | - | ✅ | SD配置管理(区别于T4只读) |
| M2+ | 资源类型指标 | ✅ | - | ✅ | 类型-指标关联管理 |

## Phase 5 模块进度

| 模块 | 名称 | 后端 | 前端 | 测试 | 备注 |
|------|------|------|------|------|------|
| A5 | JWT认证 | ✅ | ✅ | ✅ | 登录/刷新/中间件保护 |
| T6 | 任务类型管理 | ✅ | ✅ | ✅ | CRUD+内置类型保护 |
| M3+ | 网段管理 | ✅ | ✅ | ✅ | Zone级联CRUD+CIDR重叠检测 |
| B1++ | 业务系统增强 | ✅ | ✅ | ✅ | 暂停/归档/成员角色/我的系统 |
| M3++ | 实例连通测试 | ✅ | - | ✅ | TCP拨测+SSRF防护 |
| M5++ | 数据源健康检查 | ✅ | - | ✅ | HTTP探测+重定向校验 |
| T5++ | 定时任务手动触发 | ✅ | - | ✅ | 立即执行+执行记录 |
| T1++ | 采集任务批量操作 | ✅ | - | ✅ | 批量启用/禁用/删除 |
| M3+++ | 实例批量导入导出 | ✅ | - | ✅ | JSON批量操作 |

## 后端架构

```
server/
├── cmd/server/main.go          # 入口
├── internal/
│   ├── config/                  # 配置加载
│   ├── server/                  # HTTP服务器
│   ├── middleware/              # 中间件(auth/log/trace)
│   ├── model/                   # 领域模型
│   ├── handler/                 # API处理器(按模块分文件)
│   ├── service/                 # 业务逻辑(按模块分文件)
│   └── repository/
│       ├── interfaces.go        # 仓储接口定义
│       └── postgres/            # PostgreSQL实现
├── pkg/
│   ├── response/                # 统一响应
│   ├── errors/                  # 错误码
│   └── otelutil/                # OpenTelemetry工具
└── sql/postgres/                # PostgreSQL DDL迁移
    ├── 001_init.sql             # Phase 1 DDL
    ├── 002_phase2.sql           # Phase 2 DDL
    ├── 003_phase3.sql           # Phase 3 DDL
    ├── 004_phase4.sql           # Phase 4 DDL
    └── 005_phase5.sql           # Phase 5 DDL
```

## 前端架构

```
web/
├── src/
│   ├── api/                     # API调用层
│   ├── views/                   # 页面视图(按模块分目录)
│   │   ├── dashboard/           # 概览
│   │   ├── business/            # B1
│   │   ├── zone/                # M1
│   │   ├── resource/            # M2
│   │   ├── instance/            # M3
│   │   ├── datasink/            # M5
│   │   ├── task/                # T1
│   │   ├── collector/           # I1
│   │   ├── user/                # I2
│   │   ├── notification/        # A3
│   │   ├── alert-group/         # A2
│   │   ├── alert-strategy/      # A1
│   │   ├── scheduled-task/      # T5
│   │   ├── rc-task/             # T2
│   │   ├── metric/              # T3
│   │   ├── credential/          # M4
│   │   ├── alert-event/         # A4
│   │   ├── coordinator/         # I1+
│   │   ├── login/               # A5
│   │   └── task-type/           # T6
│   ├── router/                  # 路由配置
│   ├── App.vue
│   └── main.js
├── package.json
└── vite.config.js
```

## 设计决策记录

1. **SQL层拆分**: repository接口定义在internal/repository/interfaces.go，PostgreSQL实现在internal/repository/postgres/，未来可扩展MySQL/达梦等实现
2. **信创兼容**: DDL使用标准SQL+PostgreSQL方言，避免使用PG独有特性（如LATERAL、jsonb操作符等），便于后续迁移到达梦/人大金仓
3. **APM/Trace预留**: 使用OpenTelemetry标准，当前默认no-op实现，可通过配置切换到真实exporter
4. **日志**: 使用Go标准库slog，支持JSON格式输出，便于后续接入ELK/Loki

## 启动命令

```bash
# 基础设施
bash deploy/podman/deploy.sh

# 数据库迁移
cat server/sql/postgres/001_init.sql | podman exec -i dcs-postgres psql -U dcs -d dcs
cat server/sql/postgres/002_phase2.sql | podman exec -i dcs-postgres psql -U dcs -d dcs
cat server/sql/postgres/003_phase3.sql | podman exec -i dcs-postgres psql -U dcs -d dcs
cat server/sql/postgres/004_phase4.sql | podman exec -i dcs-postgres psql -U dcs -d dcs
cat server/sql/postgres/005_phase5.sql | podman exec -i dcs-postgres psql -U dcs -d dcs

# 后端
cd server && go run cmd/server/main.go

# 前端
cd web && npm run dev
```

## Git提交记录

- `c0d3b14` 协调器v1 (main分支)
- `ddce5e7` Phase 1: DCS 核心骨架完成 — Go后端 + Vue前端 + 8模块全链路 (v0.1.0-phase1)
- `2ceefe6` Phase 2: 告警层 + 用户管理 + 定时任务 — 5模块全链路 (v0.2.0-phase2)
- `55b09c4` Phase 3: RC任务 + 指标管理 + 凭据管理 + 服务发现 — 4模块全链路 (v0.3.0-phase3)
- `38dc502` Phase 4: 告警收敛 + 用户认证增强 + 协调器集成 + 业务概览 — 6模块全链路 (v0.4.0-phase4)
- Phase 5: JWT认证 + 任务类型 + 网段管理 + 业务增强 + 批量操作 — 9模块全链路 (v0.5.0-phase5)
