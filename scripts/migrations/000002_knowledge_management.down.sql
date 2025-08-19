-- AIOS Knowledge Management Schema Rollback
-- This migration removes the knowledge management database structure

-- Drop triggers first
DROP TRIGGER IF EXISTS update_knowledge_bases_updated_at ON knowledge.knowledge_bases;
DROP TRIGGER IF EXISTS update_documents_updated_at ON knowledge.documents;
DROP TRIGGER IF EXISTS update_crawl_jobs_updated_at ON knowledge.crawl_jobs;
DROP TRIGGER IF EXISTS update_entities_updated_at ON knowledge.entities;
DROP TRIGGER IF EXISTS update_search_cache_updated_at ON knowledge.search_cache;

-- Drop indexes
DROP INDEX IF EXISTS knowledge.idx_documents_full_text;
DROP INDEX IF EXISTS knowledge.idx_document_chunks_full_text;
DROP INDEX IF EXISTS knowledge.idx_entities_full_text;
DROP INDEX IF EXISTS knowledge.document_chunks_embedding_idx;

-- Drop performance indexes
DROP INDEX IF EXISTS knowledge.idx_knowledge_bases_name;
DROP INDEX IF EXISTS knowledge.idx_knowledge_bases_owner_id;
DROP INDEX IF EXISTS knowledge.idx_knowledge_bases_status;
DROP INDEX IF EXISTS knowledge.idx_documents_knowledge_base_id;
DROP INDEX IF EXISTS knowledge.idx_documents_content_type;
DROP INDEX IF EXISTS knowledge.idx_documents_processing_status;
DROP INDEX IF EXISTS knowledge.idx_documents_created_at;
DROP INDEX IF EXISTS knowledge.idx_document_chunks_document_id;
DROP INDEX IF EXISTS knowledge.idx_document_chunks_chunk_index;
DROP INDEX IF EXISTS knowledge.idx_crawl_jobs_knowledge_base_id;
DROP INDEX IF EXISTS knowledge.idx_crawl_jobs_status;
DROP INDEX IF EXISTS knowledge.idx_crawl_jobs_created_at;
DROP INDEX IF EXISTS knowledge.idx_crawled_pages_crawl_job_id;
DROP INDEX IF EXISTS knowledge.idx_crawled_pages_url;
DROP INDEX IF EXISTS knowledge.idx_crawled_pages_crawled_at;
DROP INDEX IF EXISTS knowledge.idx_entities_knowledge_base_id;
DROP INDEX IF EXISTS knowledge.idx_entities_type;
DROP INDEX IF EXISTS knowledge.idx_entities_name;
DROP INDEX IF EXISTS knowledge.idx_entity_relationships_knowledge_base_id;
DROP INDEX IF EXISTS knowledge.idx_entity_relationships_source_entity_id;
DROP INDEX IF EXISTS knowledge.idx_entity_relationships_target_entity_id;
DROP INDEX IF EXISTS knowledge.idx_entity_relationships_relationship_type;
DROP INDEX IF EXISTS knowledge.idx_search_queries_knowledge_base_id;
DROP INDEX IF EXISTS knowledge.idx_search_queries_user_id;
DROP INDEX IF EXISTS knowledge.idx_search_queries_query_type;
DROP INDEX IF EXISTS knowledge.idx_search_queries_created_at;
DROP INDEX IF EXISTS knowledge.idx_search_cache_query_hash;
DROP INDEX IF EXISTS knowledge.idx_search_cache_knowledge_base_id;
DROP INDEX IF EXISTS knowledge.idx_search_cache_expires_at;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS knowledge.search_cache;
DROP TABLE IF EXISTS knowledge.search_queries;
DROP TABLE IF EXISTS knowledge.entity_relationships;
DROP TABLE IF EXISTS knowledge.entities;
DROP TABLE IF EXISTS knowledge.crawled_pages;
DROP TABLE IF EXISTS knowledge.crawl_jobs;
DROP TABLE IF EXISTS knowledge.document_chunks;
DROP TABLE IF EXISTS knowledge.documents;
DROP TABLE IF EXISTS knowledge.knowledge_bases;

-- Drop schema
DROP SCHEMA IF EXISTS knowledge CASCADE;

-- Drop vector extension if no other tables use it
-- Note: Only drop if you're sure no other parts of the system use vector types
-- DROP EXTENSION IF EXISTS vector;
