-- Phase 5: JWT认证 + Task Type管理 + 网段管理 + 业务系统增强 + 批量操作
-- 日期: 2026-10-28

-- 1. zone表增加system_id列，支持业务系统资源隔离
ALTER TABLE zone ADD COLUMN system_id VARCHAR(64);
ALTER TABLE zone ADD CONSTRAINT fk_zone_business_system
    FOREIGN KEY (system_id) REFERENCES business_system(system_id) ON DELETE SET NULL;
CREATE INDEX idx_zone_system ON zone(system_id);

-- 2. business_system表增加status字段（支持暂停/归档）
-- 检查是否已存在
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'business_system' AND column_name = 'status'
    ) THEN
        ALTER TABLE business_system ADD COLUMN status VARCHAR(32) DEFAULT 'active';
    END IF;
END $$;
