-- DCS Phase 2 Schema Migration
-- Target: PostgreSQL 17 (compatible with 达梦/人大金仓 via standard SQL)
-- Date: 2026-10-28
-- Modules: A1(AlertStrategy), A2(AlertGroup), A3(NotificationChannel), T5(ScheduledTask)

-- ============================================================
-- A3: Notification Channel
-- ============================================================

CREATE TABLE IF NOT EXISTS notification_channel (
    channel_id      VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    channel_type    VARCHAR(32) NOT NULL,
    config          TEXT NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    last_test_at    TIMESTAMP,
    last_test_result VARCHAR(32),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_nc_type ON notification_channel(channel_type);

-- ============================================================
-- A2: Alert Group
-- ============================================================

CREATE TABLE IF NOT EXISTS alert_group (
    group_id        VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    parent_id       VARCHAR(64) REFERENCES alert_group(group_id) ON DELETE SET NULL,
    escalation_minutes INT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS alert_group_member (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    group_id        VARCHAR(64) NOT NULL REFERENCES alert_group(group_id) ON DELETE CASCADE,
    user_id         VARCHAR(64) NOT NULL REFERENCES sys_user(user_id) ON DELETE CASCADE,
    role            VARCHAR(32) NOT NULL DEFAULT 'member',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_agm_group_user UNIQUE (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_agm_group ON alert_group_member(group_id);
CREATE INDEX IF NOT EXISTS idx_agm_user ON alert_group_member(user_id);

-- ============================================================
-- A1: Alert Strategy
-- ============================================================

CREATE TABLE IF NOT EXISTS alert_strategy (
    strategy_id     VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL,
    description     TEXT,
    metric_id       VARCHAR(64) NOT NULL,
    expr_template   TEXT NOT NULL,
    duration        VARCHAR(32) NOT NULL DEFAULT '5m',
    severity        VARCHAR(32) NOT NULL DEFAULT 'warning',
    scope_type      VARCHAR(32) NOT NULL DEFAULT 'instance',
    scope_value     TEXT,
    alert_group_id  VARCHAR(64) REFERENCES alert_group(group_id) ON DELETE SET NULL,
    channel_id      VARCHAR(64) REFERENCES notification_channel(channel_id) ON DELETE SET NULL,
    labels          TEXT,
    annotations     TEXT,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    status          VARCHAR(32) NOT NULL DEFAULT 'draft',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_as_metric ON alert_strategy(metric_id);
CREATE INDEX IF NOT EXISTS idx_as_group ON alert_strategy(alert_group_id);
CREATE INDEX IF NOT EXISTS idx_as_severity ON alert_strategy(severity);
CREATE INDEX IF NOT EXISTS idx_as_status ON alert_strategy(status);

CREATE TABLE IF NOT EXISTS alert_strategy_instance (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    strategy_id     VARCHAR(64) NOT NULL REFERENCES alert_strategy(strategy_id) ON DELETE CASCADE,
    instance_id     VARCHAR(64) NOT NULL REFERENCES instance(instance_id) ON DELETE CASCADE,
    zone_id         VARCHAR(64) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_asi_strategy_instance UNIQUE (strategy_id, instance_id)
);

CREATE INDEX IF NOT EXISTS idx_asi_zone ON alert_strategy_instance(zone_id);
CREATE INDEX IF NOT EXISTS idx_asi_instance ON alert_strategy_instance(instance_id);
