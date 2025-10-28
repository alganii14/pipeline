-- Optimization for Dashboard Performance
-- Run this script to speed up dashboard queries

USE pipeline_db;

-- Drop existing indexes if any (to recreate with better structure)
DROP INDEX IF EXISTS idx_strategy ON pipelines;
DROP INDEX IF EXISTS idx_segment ON pipelines;
DROP INDEX IF EXISTS idx_pipeline ON pipelines;
DROP INDEX IF EXISTS idx_proyeksi ON pipelines;
DROP INDEX IF EXISTS idx_created_at ON pipelines;

-- Create optimized indexes for dashboard queries
-- These indexes will speed up COUNT, SUM, and GROUP BY operations

-- Index for strategy statistics (COUNT and SUM grouped by strategy)
CREATE INDEX idx_strategy_stats ON pipelines(deleted_at, strategy, proyeksi);

-- Index for segment statistics (COUNT and SUM grouped by segment)
CREATE INDEX idx_segment_stats ON pipelines(deleted_at, segment, proyeksi);

-- Index for general queries (pagination, filtering)
CREATE INDEX idx_general ON pipelines(deleted_at, created_at);

-- Composite index for total count and sum (covers deleted_at and proyeksi)
CREATE INDEX idx_total_stats ON pipelines(deleted_at, proyeksi);

-- Analyze table to update statistics
ANALYZE TABLE pipelines;

-- Show all indexes
SHOW INDEX FROM pipelines;

-- Show table status
SHOW TABLE STATUS LIKE 'pipelines';
