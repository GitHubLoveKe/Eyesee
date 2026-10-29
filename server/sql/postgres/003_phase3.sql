-- DCS Phase 3 Schema Migration
-- Target: PostgreSQL 17 (compatible with 达梦/人大金仓 via standard SQL)
-- Date: 2026-10-28
-- Modules: T2(RCTask), T3(MetricDefinition), M4(Credential)

-- ============================================================
-- T3: Metric Definition
-- ============================================================

CREATE TABLE IF NOT EXISTS metric_definition (
    metric_id       VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL,
    category        VARCHAR(32) NOT NULL DEFAULT 'metric',
    metric_type     VARCHAR(32),
    help            TEXT,
    unit            VARCHAR(64),
    promql_template TEXT,
    variables       TEXT,
    labels          TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_md_category ON metric_definition(category);
CREATE INDEX IF NOT EXISTS idx_md_name ON metric_definition(name);

CREATE TABLE IF NOT EXISTS common_query (
    query_id        VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(256) NOT NULL,
    description     TEXT,
    metric_id       VARCHAR(64) NOT NULL REFERENCES metric_definition(metric_id) ON DELETE CASCADE,
    promql          TEXT NOT NULL,
    variables       TEXT,
    category        VARCHAR(64),
    is_builtin      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cq_metric ON common_query(metric_id);

-- ============================================================
-- T2: RC Task (Rule Collection)
-- ============================================================

CREATE TABLE IF NOT EXISTS rc_task (
    task_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id),
    name            VARCHAR(256) NOT NULL,
    description     TEXT,
    strategy_id     VARCHAR(64) REFERENCES alert_strategy(strategy_id) ON DELETE SET NULL,
    query_address   VARCHAR(256),
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    status          VARCHAR(32) NOT NULL DEFAULT 'draft',
    rule_group      VARCHAR(128),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rct_zone ON rc_task(zone_id);
CREATE INDEX IF NOT EXISTS idx_rct_status ON rc_task(status);

CREATE TABLE IF NOT EXISTS rc_rule (
    rule_id         VARCHAR(64) PRIMARY KEY,
    task_id         VARCHAR(64) NOT NULL REFERENCES rc_task(task_id) ON DELETE CASCADE,
    module          VARCHAR(32) NOT NULL DEFAULT 'metrics',
    alert_name      VARCHAR(256) NOT NULL,
    expr            TEXT NOT NULL,
    duration        VARCHAR(32) NOT NULL DEFAULT '5m',
    severity        VARCHAR(32) NOT NULL DEFAULT 'warning',
    labels          TEXT,
    annotations     TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rr_task ON rc_rule(task_id);
CREATE INDEX IF NOT EXISTS idx_rr_module ON rc_rule(module);

CREATE TABLE IF NOT EXISTS rc_worker (
    worker_id       VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id),
    name            VARCHAR(128) NOT NULL,
    address         VARCHAR(256) NOT NULL,
    query_address   VARCHAR(256) NOT NULL,
    capacity        INT NOT NULL DEFAULT 500,
    owned_rules     INT NOT NULL DEFAULT 0,
    status          VARCHAR(32) NOT NULL DEFAULT 'unknown',
    last_heartbeat  TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rcw_zone ON rc_worker(zone_id);
CREATE INDEX IF NOT EXISTS idx_rcw_status ON rc_worker(status);

-- ============================================================
-- M4: Credential
-- ============================================================

CREATE TABLE IF NOT EXISTS credential (
    credential_id   VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id),
    name            VARCHAR(128) NOT NULL,
    cred_type       VARCHAR(32) NOT NULL,
    encrypted_data  TEXT NOT NULL,
    version         INT NOT NULL DEFAULT 1,
    expires_at      TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cred_zone ON credential(zone_id);
CREATE INDEX IF NOT EXISTS idx_cred_type ON credential(cred_type);
