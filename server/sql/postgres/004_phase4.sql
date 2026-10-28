-- DCS Phase 4 Schema Migration
-- Target: PostgreSQL 17 (compatible with 达梦/人大金仓 via standard SQL)
-- Date: 2026-10-28
-- Modules: A4(AlertConvergence), I2-Enh(UserAuth), I1-Enh(Coordinator), B1-Enh(Overview)

-- ============================================================
-- A4: Alert Convergence
-- ============================================================

CREATE TABLE IF NOT EXISTS alert_event (
    event_id        BIGSERIAL PRIMARY KEY,
    alert_name      VARCHAR(256) NOT NULL,
    zone_id         VARCHAR(64),
    instance_id     VARCHAR(64),
    severity        VARCHAR(32),
    status          VARCHAR(32),
    labels          TEXT,
    annotations     TEXT,
    starts_at       TIMESTAMP,
    ends_at         TIMESTAMP,
    fingerprint     VARCHAR(128),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ae_zone ON alert_event(zone_id);
CREATE INDEX IF NOT EXISTS idx_ae_status ON alert_event(status);
CREATE INDEX IF NOT EXISTS idx_ae_time ON alert_event(starts_at);
CREATE INDEX IF NOT EXISTS idx_ae_fingerprint ON alert_event(fingerprint);
CREATE INDEX IF NOT EXISTS idx_ae_severity ON alert_event(severity);

CREATE TABLE IF NOT EXISTS alert_acknowledgement (
    id              BIGSERIAL PRIMARY KEY,
    event_id        BIGINT NOT NULL REFERENCES alert_event(event_id) ON DELETE CASCADE,
    user_id         VARCHAR(64) NOT NULL,
    note            TEXT,
    acknowledged_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ack_event ON alert_acknowledgement(event_id);
CREATE INDEX IF NOT EXISTS idx_ack_user ON alert_acknowledgement(user_id);

CREATE TABLE IF NOT EXISTS alert_silence (
    silence_id      VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64),
    matchers        TEXT NOT NULL,
    starts_at       TIMESTAMP NOT NULL,
    ends_at         TIMESTAMP NOT NULL,
    duration        VARCHAR(32),
    reason          TEXT,
    created_by      VARCHAR(64) NOT NULL,
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    engine_silence_id VARCHAR(128),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_silence_zone ON alert_silence(zone_id);
CREATE INDEX IF NOT EXISTS idx_silence_status ON alert_silence(status);
CREATE INDEX IF NOT EXISTS idx_silence_time ON alert_silence(starts_at, ends_at);

-- ============================================================
-- I2-Enh: User Notification Target
-- ============================================================

CREATE TABLE IF NOT EXISTS user_notification_target (
    id              BIGSERIAL PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL,
    channel_type    VARCHAR(32) NOT NULL,
    target_value    VARCHAR(256) NOT NULL,
    is_primary      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_unt_user ON user_notification_target(user_id);

-- ============================================================
-- I1-Enh: Zone Coordinator
-- ============================================================

CREATE TABLE IF NOT EXISTS zone_coordinator (
    coordinator_id  VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id),
    name            VARCHAR(128) NOT NULL,
    address         VARCHAR(256) NOT NULL,
    gossip_addr     VARCHAR(256) NOT NULL,
    role            VARCHAR(32) NOT NULL DEFAULT 'leader',
    managed_nodes   INT NOT NULL DEFAULT 0,
    status          VARCHAR(32) NOT NULL DEFAULT 'unknown',
    last_heartbeat  TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_zc_zone ON zone_coordinator(zone_id);
CREATE INDEX IF NOT EXISTS idx_zc_status ON zone_coordinator(status);

-- ============================================================
-- T4-Enh: Service Discovery Configuration
-- ============================================================

CREATE TABLE IF NOT EXISTS service_discovery (
    discovery_id    VARCHAR(64) PRIMARY KEY,
    zone_id         VARCHAR(64) NOT NULL REFERENCES zone(zone_id),
    name            VARCHAR(128) NOT NULL,
    protocol        VARCHAR(32) NOT NULL,
    config          TEXT NOT NULL,
    refresh_interval VARCHAR(32) NOT NULL DEFAULT '30s',
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    status          VARCHAR(32) NOT NULL DEFAULT 'active',
    last_refresh_at TIMESTAMP,
    discovered_count INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sd_zone ON service_discovery(zone_id);
CREATE INDEX IF NOT EXISTS idx_sd_protocol ON service_discovery(protocol);

-- ============================================================
-- M2-Enh: Resource Type Metric Inheritance
-- ============================================================

CREATE TABLE IF NOT EXISTS resource_type_metric (
    id              BIGSERIAL PRIMARY KEY,
    type_id         VARCHAR(64) NOT NULL REFERENCES resource_type(type_id) ON DELETE CASCADE,
    metric_id       VARCHAR(64) NOT NULL REFERENCES metric_definition(metric_id) ON DELETE CASCADE,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (type_id, metric_id)
);

CREATE INDEX IF NOT EXISTS idx_rtm_type ON resource_type_metric(type_id);
