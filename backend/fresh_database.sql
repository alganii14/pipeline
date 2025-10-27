-- Drop database jika sudah ada
DROP DATABASE IF EXISTS pipeline_db;

-- Create database baru
CREATE DATABASE pipeline_db;
USE pipeline_db;

-- Create pipelines table (kosong, tanpa seed data)
CREATE TABLE IF NOT EXISTS pipelines (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    pn VARCHAR(255),
    nama_rmft VARCHAR(255),
    kode_uker VARCHAR(255),
    kc VARCHAR(255),
    prod VARCHAR(255),
    no_rek VARCHAR(255),
    dup VARCHAR(255),
    nama VARCHAR(255),
    tgl VARCHAR(255),
    strategy VARCHAR(255),
    segment VARCHAR(255),
    pipeline VARCHAR(255),
    proyeksi DOUBLE DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    INDEX idx_pn (pn),
    INDEX idx_kode_uker (kode_uker),
    INDEX idx_strategy (strategy),
    INDEX idx_segment (segment)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Database siap, tabel kosong (0 data)
