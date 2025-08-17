-- AIOS Initial Database Schema
-- This migration creates the foundational database structure for AIOS

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create schemas for organization
CREATE SCHEMA IF NOT EXISTS aios;
CREATE SCHEMA IF NOT EXISTS ai_models;
CREATE SCHEMA IF NOT EXISTS system_metrics;
CREATE SCHEMA IF NOT EXISTS audit;

-- Set default search path
SET search_path TO aios, public;

-- Users table
CREATE TABLE IF NOT EXISTS aios.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT true,
    is_admin BOOLEAN DEFAULT false,
    email_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Sessions table for authentication
CREATE TABLE IF NOT EXISTS aios.sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES aios.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    refresh_expires_at TIMESTAMP WITH TIME ZONE,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User preferences
CREATE TABLE IF NOT EXISTS aios.user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES aios.users(id) ON DELETE CASCADE,
    theme VARCHAR(50) DEFAULT 'dark',
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(100) DEFAULT 'UTC',
    ai_assistant_enabled BOOLEAN DEFAULT true,
    voice_control_enabled BOOLEAN DEFAULT false,
    notifications_enabled BOOLEAN DEFAULT true,
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);

-- AI Models registry
CREATE TABLE IF NOT EXISTS ai_models.models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL, -- llm, cv, optimization, etc.
    provider VARCHAR(100) NOT NULL, -- ollama, openai, local, etc.
    model_path TEXT,
    config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    size_bytes BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, version, type)
);

-- AI Model usage tracking
CREATE TABLE IF NOT EXISTS ai_models.usage_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    model_id UUID NOT NULL REFERENCES ai_models.models(id) ON DELETE CASCADE,
    user_id UUID REFERENCES aios.users(id) ON DELETE SET NULL,
    request_type VARCHAR(100) NOT NULL,
    input_tokens INTEGER,
    output_tokens INTEGER,
    processing_time_ms INTEGER,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- System metrics
CREATE TABLE IF NOT EXISTS system_metrics.performance_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    service_name VARCHAR(100) NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,6) NOT NULL,
    unit VARCHAR(50),
    tags JSONB DEFAULT '{}',
    INDEX (timestamp, service_name, metric_name)
);

-- System health checks
CREATE TABLE IF NOT EXISTS system_metrics.health_checks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name VARCHAR(100) NOT NULL,
    check_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL, -- healthy, unhealthy, warning
    response_time_ms INTEGER,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX (checked_at, service_name, status)
);

-- Audit logs
CREATE TABLE IF NOT EXISTS audit.activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES aios.users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX (created_at, user_id, action)
);

-- File system AI metadata
CREATE TABLE IF NOT EXISTS aios.file_metadata (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    file_path TEXT NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(255),
    ai_tags TEXT[],
    ai_description TEXT,
    ai_confidence DECIMAL(3,2),
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(file_path)
);

-- Desktop workspaces
CREATE TABLE IF NOT EXISTS aios.workspaces (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES aios.users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    layout JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Desktop applications
CREATE TABLE IF NOT EXISTS aios.applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    executable_path TEXT NOT NULL,
    icon_path TEXT,
    category VARCHAR(100),
    ai_enhanced BOOLEAN DEFAULT false,
    auto_launch BOOLEAN DEFAULT false,
    config JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_username ON aios.users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON aios.users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON aios.users(is_active);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON aios.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON aios.sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON aios.sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON aios.sessions(is_active);

CREATE INDEX IF NOT EXISTS idx_models_type ON ai_models.models(type);
CREATE INDEX IF NOT EXISTS idx_models_provider ON ai_models.models(provider);
CREATE INDEX IF NOT EXISTS idx_models_active ON ai_models.models(is_active);

CREATE INDEX IF NOT EXISTS idx_usage_logs_model_id ON ai_models.usage_logs(model_id);
CREATE INDEX IF NOT EXISTS idx_usage_logs_user_id ON ai_models.usage_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_logs_created_at ON ai_models.usage_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_file_metadata_path ON aios.file_metadata(file_path);
CREATE INDEX IF NOT EXISTS idx_file_metadata_hash ON aios.file_metadata(file_hash);
CREATE INDEX IF NOT EXISTS idx_file_metadata_tags ON aios.file_metadata USING GIN(ai_tags);

CREATE INDEX IF NOT EXISTS idx_workspaces_user_id ON aios.workspaces(user_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_active ON aios.workspaces(is_active);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON aios.users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_preferences_updated_at BEFORE UPDATE ON aios.user_preferences FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_models_updated_at BEFORE UPDATE ON ai_models.models FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_file_metadata_updated_at BEFORE UPDATE ON aios.file_metadata FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_workspaces_updated_at BEFORE UPDATE ON aios.workspaces FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON aios.applications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default data
INSERT INTO aios.users (username, email, password_hash, first_name, last_name, is_admin) 
VALUES ('admin', 'admin@aios.dev', '$2a$10$rQZ8ZqNQzqNQzqNQzqNQzOeKqNQzqNQzqNQzqNQzqNQzqNQzqNQzq', 'AIOS', 'Administrator', true)
ON CONFLICT (username) DO NOTHING;

-- Insert default AI models
INSERT INTO ai_models.models (name, version, type, provider, is_default) VALUES
('llama2', '7b', 'llm', 'ollama', true),
('codellama', '7b', 'llm', 'ollama', false),
('mistral', '7b', 'llm', 'ollama', false)
ON CONFLICT (name, version, type) DO NOTHING;

-- Insert default applications
INSERT INTO aios.applications (name, executable_path, category, ai_enhanced) VALUES
('Terminal', '/usr/bin/gnome-terminal', 'System', true),
('File Manager', '/usr/bin/nautilus', 'System', true),
('Text Editor', '/usr/bin/gedit', 'Development', true),
('Web Browser', '/usr/bin/firefox', 'Internet', true)
ON CONFLICT (name) DO NOTHING;
