-- Optimization script for existing Pipeline database
-- Run this script in MySQL to add indexes for better performance

USE pipeline_db;

-- Add indexes for better query performance
ALTER TABLE pipelines ADD INDEX idx_strategy (strategy);
ALTER TABLE pipelines ADD INDEX idx_segment (segment);
ALTER TABLE pipelines ADD INDEX idx_pipeline (pipeline);
ALTER TABLE pipelines ADD INDEX idx_proyeksi (proyeksi);
ALTER TABLE pipelines ADD INDEX idx_created_at (created_at);

-- Composite indexes for common query patterns
ALTER TABLE pipelines ADD INDEX idx_strategy_proyeksi (strategy, proyeksi);
ALTER TABLE pipelines ADD INDEX idx_segment_proyeksi (segment, proyeksi);

-- Optimize table after adding indexes
OPTIMIZE TABLE pipelines;

-- Show table indexes
SHOW INDEX FROM pipelines;
