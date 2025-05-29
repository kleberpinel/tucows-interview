-- Remove additional columns from properties table
ALTER TABLE properties 
DROP INDEX idx_external_id,
DROP INDEX idx_mls_number,
DROP COLUMN external_id,
DROP COLUMN mls_number,
DROP COLUMN property_type,
DROP COLUMN bedrooms,
DROP COLUMN bathrooms,
DROP COLUMN square_feet,
DROP COLUMN lot_size,
DROP COLUMN year_built;
