-- Add additional columns to properties table
ALTER TABLE properties 
ADD COLUMN external_id VARCHAR(255) DEFAULT NULL,
ADD COLUMN mls_number VARCHAR(255) DEFAULT NULL,
ADD COLUMN property_type VARCHAR(100) DEFAULT NULL,
ADD COLUMN bedrooms INT DEFAULT NULL,
ADD COLUMN bathrooms INT DEFAULT NULL,
ADD COLUMN square_feet INT DEFAULT NULL,
ADD COLUMN lot_size VARCHAR(100) DEFAULT NULL,
ADD COLUMN year_built INT DEFAULT NULL,
ADD INDEX idx_external_id (external_id),
ADD INDEX idx_mls_number (mls_number);
