-- Drop pipelines table from database
-- Execute this SQL in your MySQL/MariaDB database

-- First, drop foreign key constraint from rfmts table
ALTER TABLE rfmts DROP FOREIGN KEY fk_rfmts_pipeline;

-- Drop pipeline_id column from rfmts table
ALTER TABLE rfmts DROP COLUMN pipeline_id;

-- Now drop the pipelines table
DROP TABLE IF EXISTS pipelines;

-- Done! The pipelines table has been removed from the database.
