-- AIOS Knowledge Management Schema
-- This migration creates the database structure for knowledge management and RAG capabilities

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS vector;

-- Create knowledge management schema
CREATE SCHEMA IF NOT EXISTS knowledge;

-- Set search path
SET search_path TO knowledge, aios, public;

-- Knowledge bases table
CREATE TABLE IF NOT EXISTS knowledge.knowledge_bases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID, -- References aios.users(id) if users table exists
    config JSONB DEFAULT '{}',
    stats JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, owner_id)
);

-- Documents table
CREATE TABLE IF NOT EXISTS knowledge.documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge.knowledge_bases(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(100) NOT NULL DEFAULT 'text',
    language VARCHAR(10) DEFAULT 'en',
    source VARCHAR(500),
    url TEXT,
    author VARCHAR(255),
    file_path TEXT,
    file_size BIGINT,
    file_hash VARCHAR(64),
    mime_type VARCHAR(255),
    tags TEXT[],
    categories TEXT[],
    metadata JSONB DEFAULT '{}',
    processing_status VARCHAR(50) DEFAULT 'pending',
    processing_error TEXT,
    version INTEGER DEFAULT 1,
    parent_id UUID REFERENCES knowledge.documents(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE
);

-- Document chunks table for RAG
CREATE TABLE IF NOT EXISTS knowledge.document_chunks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES knowledge.documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    content_length INTEGER NOT NULL,
    chunk_type VARCHAR(50) DEFAULT 'text',
    start_position INTEGER,
    end_position INTEGER,
    metadata JSONB DEFAULT '{}',
    embedding vector(1536), -- OpenAI ada-002 embedding dimension
    embedding_model VARCHAR(100) DEFAULT 'text-embedding-ada-002',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(document_id, chunk_index)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_knowledge_bases_name ON knowledge.knowledge_bases(name);
CREATE INDEX IF NOT EXISTS idx_knowledge_bases_owner_id ON knowledge.knowledge_bases(owner_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_bases_status ON knowledge.knowledge_bases(status);

CREATE INDEX IF NOT EXISTS idx_documents_knowledge_base_id ON knowledge.documents(knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_documents_content_type ON knowledge.documents(content_type);
CREATE INDEX IF NOT EXISTS idx_documents_processing_status ON knowledge.documents(processing_status);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON knowledge.documents(created_at);

CREATE INDEX IF NOT EXISTS idx_document_chunks_document_id ON knowledge.document_chunks(document_id);
CREATE INDEX IF NOT EXISTS idx_document_chunks_chunk_index ON knowledge.document_chunks(chunk_index);

-- Create vector similarity index for embeddings
CREATE INDEX IF NOT EXISTS document_chunks_embedding_idx ON knowledge.document_chunks 
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_crawl_jobs_knowledge_base_id ON knowledge.crawl_jobs(knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_crawl_jobs_status ON knowledge.crawl_jobs(status);
CREATE INDEX IF NOT EXISTS idx_crawl_jobs_created_at ON knowledge.crawl_jobs(created_at);

CREATE INDEX IF NOT EXISTS idx_crawled_pages_crawl_job_id ON knowledge.crawled_pages(crawl_job_id);
CREATE INDEX IF NOT EXISTS idx_crawled_pages_url ON knowledge.crawled_pages(url);
CREATE INDEX IF NOT EXISTS idx_crawled_pages_crawled_at ON knowledge.crawled_pages(crawled_at);

CREATE INDEX IF NOT EXISTS idx_entities_knowledge_base_id ON knowledge.entities(knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_entities_type ON knowledge.entities(type);
CREATE INDEX IF NOT EXISTS idx_entities_name ON knowledge.entities(name);

CREATE INDEX IF NOT EXISTS idx_entity_relationships_knowledge_base_id ON knowledge.entity_relationships(knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_entity_relationships_source_entity_id ON knowledge.entity_relationships(source_entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_relationships_target_entity_id ON knowledge.entity_relationships(target_entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_relationships_relationship_type ON knowledge.entity_relationships(relationship_type);

CREATE INDEX IF NOT EXISTS idx_search_queries_knowledge_base_id ON knowledge.search_queries(knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_search_queries_user_id ON knowledge.search_queries(user_id);
CREATE INDEX IF NOT EXISTS idx_search_queries_query_type ON knowledge.search_queries(query_type);
CREATE INDEX IF NOT EXISTS idx_search_queries_created_at ON knowledge.search_queries(created_at);

CREATE INDEX IF NOT EXISTS idx_search_cache_query_hash ON knowledge.search_cache(query_hash);
CREATE INDEX IF NOT EXISTS idx_search_cache_knowledge_base_id ON knowledge.search_cache(knowledge_base_id);
CREATE INDEX IF NOT EXISTS idx_search_cache_expires_at ON knowledge.search_cache(expires_at);

-- Crawl jobs table
CREATE TABLE IF NOT EXISTS knowledge.crawl_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge.knowledge_bases(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    max_pages INTEGER DEFAULT 100,
    max_depth INTEGER DEFAULT 3,
    follow_links BOOLEAN DEFAULT true,
    pages_found INTEGER DEFAULT 0,
    pages_processed INTEGER DEFAULT 0,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
);

-- Crawled pages table
CREATE TABLE IF NOT EXISTS knowledge.crawled_pages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    crawl_job_id UUID NOT NULL REFERENCES knowledge.crawl_jobs(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    title VARCHAR(500),
    content TEXT,
    content_length INTEGER,
    status_code INTEGER,
    content_type VARCHAR(255),
    links TEXT[],
    depth INTEGER DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    crawled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(crawl_job_id, url)
);

-- Knowledge entities table (for knowledge graph)
CREATE TABLE IF NOT EXISTS knowledge.entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge.knowledge_bases(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    description TEXT,
    properties JSONB DEFAULT '{}',
    aliases TEXT[],
    embedding vector(1536),
    confidence DECIMAL(3,2) DEFAULT 0.0,
    source VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(knowledge_base_id, name, type)
);

-- Entity relationships table
CREATE TABLE IF NOT EXISTS knowledge.entity_relationships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge.knowledge_bases(id) ON DELETE CASCADE,
    source_entity_id UUID NOT NULL REFERENCES knowledge.entities(id) ON DELETE CASCADE,
    target_entity_id UUID NOT NULL REFERENCES knowledge.entities(id) ON DELETE CASCADE,
    relationship_type VARCHAR(100) NOT NULL,
    properties JSONB DEFAULT '{}',
    confidence DECIMAL(3,2) DEFAULT 0.0,
    source VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(source_entity_id, target_entity_id, relationship_type)
);

-- Search queries table (for analytics and caching)
CREATE TABLE IF NOT EXISTS knowledge.search_queries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge.knowledge_bases(id) ON DELETE CASCADE,
    user_id UUID, -- References aios.users(id) if users table exists
    query_text TEXT NOT NULL,
    query_type VARCHAR(50) DEFAULT 'semantic', -- semantic, keyword, hybrid
    query_embedding vector(1536),
    results_count INTEGER DEFAULT 0,
    response_time_ms INTEGER,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
);

-- Search results cache table
CREATE TABLE IF NOT EXISTS knowledge.search_cache (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    query_hash VARCHAR(64) NOT NULL,
    knowledge_base_id UUID NOT NULL REFERENCES knowledge.knowledge_bases(id) ON DELETE CASCADE,
    query_text TEXT NOT NULL,
    results JSONB NOT NULL,
    hit_count INTEGER DEFAULT 1,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(query_hash, knowledge_base_id)
);

-- Create updated_at triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to tables
CREATE TRIGGER update_knowledge_bases_updated_at BEFORE UPDATE ON knowledge.knowledge_bases FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON knowledge.documents FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_crawl_jobs_updated_at BEFORE UPDATE ON knowledge.crawl_jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entities_updated_at BEFORE UPDATE ON knowledge.entities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_search_cache_updated_at BEFORE UPDATE ON knowledge.search_cache FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_documents_full_text ON knowledge.documents USING gin(to_tsvector('english', title || ' ' || content));
CREATE INDEX IF NOT EXISTS idx_document_chunks_full_text ON knowledge.document_chunks USING gin(to_tsvector('english', content));
CREATE INDEX IF NOT EXISTS idx_entities_full_text ON knowledge.entities USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- Insert default knowledge base (without user dependency)
INSERT INTO knowledge.knowledge_bases (name, description, owner_id, config) 
VALUES (
    'Default Knowledge Base',
    'Default knowledge base for AIOS',
    uuid_generate_v4(), -- Generate a default UUID for owner
    '{"embedding_model": "text-embedding-ada-002", "chunk_size": 1000, "chunk_overlap": 200, "indexing_enabled": true}'::jsonb
)
ON CONFLICT (name, owner_id) DO NOTHING;
