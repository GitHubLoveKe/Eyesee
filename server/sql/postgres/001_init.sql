-- DCS Phase 1 Schema Migration
-- Target: PostgreSQL 17 (compatible with 达梦/人大金仓 via standard SQL)
-- Date: 2026-10-28

-- ============================================================
-- B1: Business System
-- ============================================================

CREATE TABLE IF NOT EXISTS business_system (
    system_id       VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS business_system_member (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    system_id       VARCHAR(64) NOT NULL REFERENCES business_system(system_id) ON DELETE CASCADE,
    user_id         VARCHAR(64) NOT NULL,
    role            VARCHAR(32) NOT NULL DEFAULT 'viewer',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_bsm_system_user UNIQUE (system_id, user_id)
);

CREATE TABLE IF NOT EXISTS business_system_zone (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    system_id       VARCHAR(64) NOT NULL REFERENCES business_system(system_id) ON DELETE CASCADE,
    zone_id         VARCHAR(64) NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_bsz_system_zone UNIQUE (system_id, zone_id)
);

-- ============================================================
-- M1: Zone & Network Segment
-- ============================================================

CREATE TABLE IF NOT EXISTS zone (
    zone_id         VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    collector_type  VARCHAR(32) NOT NULL DEFAULT 'vmagent',
    remote_write_url VARCHAR(512),
    default_sink_id VARCHAR(64),
    config          TEXT,
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS network_segment (
    segment_id      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id) ON DELETE CASCADE,
    cidr            VARCHAR(45) NOT NULL,
    name            VARCHAR(128),
    description     TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_ns_zone_cidr UNIQUE (zone_id, cidr)
);

CREATE INDEX IF NOT EXISTS idx_ns_cidr ON network_segment(cidr);

-- ============================================================
-- M2: Resource Type Tree
-- ============================================================

CREATE TABLE IF NOT EXISTS resource_type (
    type_id         VARCHAR(64) PRIMARY KEY,
    parent_id       VARCHAR(64) REFERENCES resource_type(type_id) ON DELETE SET NULL,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    icon            VARCHAR(64),
    sort_order      INT NOT NULL DEFAULT 0,
    attributes      TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rt_parent ON resource_type(parent_id);

-- ============================================================
-- M3: Instance
-- ============================================================

CREATE TABLE IF NOT EXISTS instance (
    instance_id     VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id),
    resource_type_id VARCHAR(64) NOT NULL REFERENCES resource_type(type_id),
    business_system_id VARCHAR(64) REFERENCES business_system(system_id),
    name            VARCHAR(256) NOT NULL,
    host            VARCHAR(256) NOT NULL,
    port            INT,
    credential_id   VARCHAR(64),
    labels          TEXT,
    extra           TEXT,
    status          VARCHAR(32) NOT NULL DEFAULT 'pending',
    last_test_at    TIMESTAMP,
    last_test_result VARCHAR(32),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_inst_zone ON instance(zone_id);
CREATE INDEX IF NOT EXISTS idx_inst_type ON instance(resource_type_id);
CREATE INDEX IF NOT EXISTS idx_inst_host ON instance(host);
CREATE INDEX IF NOT EXISTS idx_inst_status ON instance(status);
CREATE INDEX IF NOT EXISTS idx_inst_biz ON instance(business_system_id);

-- ============================================================
-- M5: Data Sink
-- ============================================================

CREATE TABLE IF NOT EXISTS data_sink (
    sink_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id),
    name            VARCHAR(128) NOT NULL,
    sink_type       VARCHAR(32) NOT NULL,
    protocol        VARCHAR(32) NOT NULL,
    config          TEXT NOT NULL,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    last_check_at   TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sink_zone ON data_sink(zone_id);
CREATE INDEX IF NOT EXISTS idx_sink_type ON data_sink(sink_type);

-- Add foreign key for zone.default_sink_id after data_sink table exists
ALTER TABLE zone ADD CONSTRAINT fk_zone_default_sink
    FOREIGN KEY (default_sink_id) REFERENCES data_sink(sink_id) ON DELETE SET NULL;

-- ============================================================
-- T1: Collection Task
-- ============================================================

CREATE TABLE IF NOT EXISTS task_type (
    type_id         VARCHAR(64) PRIMARY KEY,
    name            VARCHAR(128) NOT NULL,
    description     TEXT,
    source_type     VARCHAR(32) NOT NULL,
    collector       VARCHAR(32) NOT NULL,
    form_schema     TEXT NOT NULL,
    detail_panels   TEXT,
    declared_metrics TEXT,
    config_template TEXT NOT NULL,
    is_builtin      BOOLEAN NOT NULL DEFAULT FALSE,
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS collection_task (
    task_id         VARCHAR(64) PRIMARY KEY,
    instance_id     VARCHAR(64) NOT NULL REFERENCES instance(instance_id) ON DELETE CASCADE,
    type_id         VARCHAR(64) NOT NULL REFERENCES task_type(type_id),
    module          VARCHAR(32) NOT NULL DEFAULT 'metrics',
    node_id         VARCHAR(64),
    data_sink_id    VARCHAR(64) REFERENCES data_sink(sink_id),
    name            VARCHAR(256) NOT NULL,
    config          TEXT NOT NULL,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    status          VARCHAR(32) NOT NULL DEFAULT 'pending',
    last_scrape_at  TIMESTAMP,
    last_error      TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ct_instance ON collection_task(instance_id);
CREATE INDEX IF NOT EXISTS idx_ct_type ON collection_task(type_id);
CREATE INDEX IF NOT EXISTS idx_ct_module ON collection_task(module);
CREATE INDEX IF NOT EXISTS idx_ct_node ON collection_task(node_id);
CREATE INDEX IF NOT EXISTS idx_ct_status ON collection_task(status);

-- ============================================================
-- I1: Collector Node
-- ============================================================

CREATE TABLE IF NOT EXISTS collector_node (
    node_id         VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id),
    name            VARCHAR(128) NOT NULL,
    node_type       VARCHAR(32) NOT NULL,
    address         VARCHAR(256) NOT NULL,
    health_endpoint VARCHAR(256) NOT NULL DEFAULT '/-/healthy',
    labels          TEXT,
    status          VARCHAR(32) NOT NULL DEFAULT 'unknown',
    task_count      INT NOT NULL DEFAULT 0,
    last_heartbeat  TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cn_zone ON collector_node(zone_id);
CREATE INDEX IF NOT EXISTS idx_cn_status ON collector_node(status);

-- ============================================================
-- I2: User (Phase 1 basic, Phase 4 full)
-- ============================================================

CREATE TABLE IF NOT EXISTS sys_user (
    user_id         VARCHAR(64) PRIMARY KEY,
    username        VARCHAR(128) NOT NULL UNIQUE,
    display_name    VARCHAR(256),
    email           VARCHAR(256),
    phone           VARCHAR(32),
    role            VARCHAR(32) NOT NULL DEFAULT 'viewer',
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    last_login_at   TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_auth_source (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL REFERENCES sys_user(user_id) ON DELETE CASCADE,
    source_type     VARCHAR(32) NOT NULL,
    external_id     VARCHAR(256),
    password_hash   VARCHAR(256),
    extra           TEXT,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_uas_user_source UNIQUE (user_id, source_type)
);

-- ============================================================
-- T5: Scheduled Task
-- ============================================================

CREATE TABLE IF NOT EXISTS scheduled_task (
    task_id         VARCHAR(64) PRIMARY KEY,
    business_system_id VARCHAR(64) REFERENCES business_system(system_id),
    name            VARCHAR(256) NOT NULL,
    task_type       VARCHAR(32) NOT NULL,
    cron_expr       VARCHAR(128) NOT NULL,
    config          TEXT,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    last_run_at     TIMESTAMP,
    next_run_at     TIMESTAMP,
    status          VARCHAR(32) NOT NULL DEFAULT 'idle',
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_st_type ON scheduled_task(task_type);
CREATE INDEX IF NOT EXISTS idx_st_status ON scheduled_task(status);
CREATE INDEX IF NOT EXISTS idx_st_next_run ON scheduled_task(next_run_at);

CREATE TABLE IF NOT EXISTS scheduled_task_execution (
    execution_id    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id         VARCHAR(64) NOT NULL REFERENCES scheduled_task(task_id) ON DELETE CASCADE,
    started_at      TIMESTAMP NOT NULL,
    completed_at    TIMESTAMP,
    duration_ms     INT,
    status          VARCHAR(32) NOT NULL,
    error_message   TEXT
);

CREATE INDEX IF NOT EXISTS idx_ste_task ON scheduled_task_execution(task_id);
CREATE INDEX IF NOT EXISTS idx_ste_time ON scheduled_task_execution(started_at);
