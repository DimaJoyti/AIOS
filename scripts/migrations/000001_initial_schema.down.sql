-- AIOS Initial Schema Rollback
-- This migration removes the foundational database structure

-- Drop triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON aios.users;
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON aios.user_preferences;
DROP TRIGGER IF EXISTS update_models_updated_at ON ai_models.models;
DROP TRIGGER IF EXISTS update_file_metadata_updated_at ON aios.file_metadata;
DROP TRIGGER IF EXISTS update_workspaces_updated_at ON aios.workspaces;
DROP TRIGGER IF EXISTS update_applications_updated_at ON aios.applications;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS aios.applications;
DROP TABLE IF EXISTS aios.workspaces;
DROP TABLE IF EXISTS aios.file_metadata;
DROP TABLE IF EXISTS audit.activity_logs;
DROP TABLE IF EXISTS system_metrics.health_checks;
DROP TABLE IF EXISTS system_metrics.performance_logs;
DROP TABLE IF EXISTS ai_models.usage_logs;
DROP TABLE IF EXISTS ai_models.models;
DROP TABLE IF EXISTS aios.user_preferences;
DROP TABLE IF EXISTS aios.sessions;
DROP TABLE IF EXISTS aios.users;

-- Drop schemas
DROP SCHEMA IF EXISTS audit CASCADE;
DROP SCHEMA IF EXISTS system_metrics CASCADE;
DROP SCHEMA IF EXISTS ai_models CASCADE;
DROP SCHEMA IF EXISTS aios CASCADE;

-- Drop extensions (be careful with this in production)
-- DROP EXTENSION IF EXISTS "pg_stat_statements";
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
