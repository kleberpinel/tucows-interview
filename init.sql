-- This file can be used for initial database setup if needed
-- The migrations will handle table creation

-- Create database and set default charset
CREATE DATABASE IF NOT EXISTS real_estate_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE real_estate_db;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create properties table with all new fields
CREATE TABLE IF NOT EXISTS properties (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    description TEXT,
    photos JSON DEFAULT NULL,
    external_id VARCHAR(255) DEFAULT NULL,
    mls_number VARCHAR(255) DEFAULT NULL,
    property_type VARCHAR(100) DEFAULT NULL,
    bedrooms INT DEFAULT NULL,
    bathrooms INT DEFAULT NULL,
    square_feet INT DEFAULT NULL,
    lot_size VARCHAR(100) DEFAULT NULL,
    year_built INT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_external_id (external_id),
    INDEX idx_mls_number (mls_number)
);