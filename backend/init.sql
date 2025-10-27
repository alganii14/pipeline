-- Create database
CREATE DATABASE IF NOT EXISTS pipeline_db;
USE pipeline_db;

-- Create pipelines table
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

-- Insert seed data (10 rows)
INSERT INTO pipelines (pn, nama_rmft, kode_uker, kc, prod, no_rek, dup, nama, tgl, strategy, segment, pipeline, proyeksi) VALUES
('PN001', 'RMFT Jakarta', 'UK001', 'KC Jakarta Pusat', 'PROD001', 'REK001', 'DUP001', 'Budi Santoso', '2025-01-15', 'Kolaborasi Perusahaan Anak', 'KONSUMER', 'Pipeline A', 1500000.50),
('PN002', 'RMFT Bandung', 'UK002', 'KC Bandung Utara', 'PROD002', 'REK002', 'DUP002', 'Siti Nurhaliza', '2025-02-20', 'Optimalisasi Business Cluster', 'Merchant', 'Pipeline B', 2300000.75),
('PN003', 'RMFT Surabaya', 'UK003', 'KC Surabaya Timur', 'PROD003', 'REK003', 'DUP003', 'Ahmad Dahlan', '2025-03-10', 'Optimalisasi Digital Channel', 'Mikro', 'Pipeline C', 1800000.00),
('PN004', 'RMFT Medan', 'UK004', 'KC Medan Selatan', 'PROD004', 'REK004', 'DUP004', 'Dewi Lestari', '2025-04-05', 'Optimalisasi Nasabah Prio, BOD, BOC', 'Prioritas', 'Pipeline A', 3200000.25),
('PN005', 'RMFT Makassar', 'UK005', 'KC Makassar Barat', 'PROD005', 'REK005', 'DUP005', 'Rizki Pratama', '2025-05-18', 'Penguatan Produk & Fungsi RM', 'RITEL BADAN USAHA', 'Pipeline B', 2700000.50),
('PN006', 'RMFT Palembang', 'UK006', 'KC Palembang Tengah', 'PROD006', 'REK006', 'DUP006', 'Linda Wijaya', '2025-06-22', 'Peningkatan Payroll Berkualitas', 'RITEL INDIVIDU', 'Pipeline C', 1950000.00),
('PN007', 'RMFT Denpasar', 'UK007', 'KC Denpasar Utara', 'PROD007', 'REK007', 'DUP007', 'Eko Prasetyo', '2025-07-14', 'Reaktivasi Rek Dormant', 'RITEL NON INDIVIDU', 'Pipeline A', 2850000.75),
('PN008', 'RMFT Semarang', 'UK008', 'KC Semarang Barat', 'PROD008', 'REK008', 'DUP008', 'Maya Safitri', '2025-08-09', 'Rekening Trx Debitur', 'Ritel Perusahaan', 'Pipeline B', 2100000.00),
('PN009', 'RMFT Yogyakarta', 'UK009', 'KC Yogyakarta Selatan', 'PROD009', 'REK009', 'DUP009', 'Bambang Susilo', '2025-09-25', 'Kolaborasi Perusahaan Anak', 'SME', 'Pipeline C', 1650000.50),
('PN010', 'RMFT Malang', 'UK010', 'KC Malang Timur', 'PROD010', 'REK010', 'DUP010', 'Sari Indah', '2025-10-30', 'Optimalisasi Business Cluster', 'WEALTH', 'Pipeline A', 3500000.00);
