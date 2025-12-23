-- PostgreSQL Stress Test Tuning Script
-- Run this as postgres superuser

-- Check current settings
SELECT name, setting, unit, context 
FROM pg_settings 
WHERE name IN ('max_connections', 'shared_buffers', 'work_mem', 'effective_cache_size', 'wal_buffers');

-- To change max_connections (requires restart):
-- ALTER SYSTEM SET max_connections = 500;

-- These can be changed without restart:
ALTER SYSTEM SET work_mem = '16MB';
ALTER SYSTEM SET maintenance_work_mem = '128MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET checkpoint_completion_target = 0.9;

-- Reload config (doesn't apply max_connections)
SELECT pg_reload_conf();

-- For max_connections change, you need to:
-- 1. Stop PostgreSQL service
-- 2. Edit postgresql.conf: max_connections = 500
-- 3. Edit postgresql.conf: shared_buffers = 256MB
-- 4. Start PostgreSQL service

-- Verify changes
SELECT name, setting, unit 
FROM pg_settings 
WHERE name IN ('max_connections', 'shared_buffers', 'work_mem');
