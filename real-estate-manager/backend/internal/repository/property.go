package repository

import (
	"context"
	"database/sql"
	"errors"
	"real-estate-manager/backend/internal/models"
)

type PropertyRepository interface {
	Create(ctx context.Context, property *models.Property) error
	GetByID(ctx context.Context, id int) (*models.Property, error)
	Update(ctx context.Context, property *models.Property) error
	Delete(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]models.Property, error)
}

type propertyRepository struct {
	db *sql.DB
}

func NewPropertyRepository(db *sql.DB) PropertyRepository {
	return &propertyRepository{db: db}
}

func (r *propertyRepository) Create(ctx context.Context, property *models.Property) error {
	query := `INSERT INTO properties (name, location, price, description, photos, external_id, mls_number, 
		property_type, bedrooms, bathrooms, square_feet, lot_size, year_built) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.ExecContext(ctx, query, 
		property.Name, property.Location, property.Price, property.Description, property.Photos,
		property.ExternalID, property.MLSNumber, property.PropertyType,
		property.Bedrooms, property.Bathrooms, property.SquareFeet, property.LotSize, property.YearBuilt)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	property.ID = int(id)
	return nil
}

func (r *propertyRepository) GetByID(ctx context.Context, id int) (*models.Property, error) {
	query := `SELECT id, name, location, price, description, photos, external_id, mls_number, 
		property_type, bedrooms, bathrooms, square_feet, lot_size, year_built, created_at, updated_at 
		FROM properties WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var property models.Property
	if err := row.Scan(&property.ID, &property.Name, &property.Location, &property.Price, 
		&property.Description, &property.Photos, &property.ExternalID, &property.MLSNumber,
		&property.PropertyType, &property.Bedrooms, &property.Bathrooms, &property.SquareFeet,
		&property.LotSize, &property.YearBuilt, &property.CreatedAt, &property.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &property, nil
}

func (r *propertyRepository) Update(ctx context.Context, property *models.Property) error {
	query := `UPDATE properties SET name = ?, location = ?, price = ?, description = ?, photos = ?, 
		external_id = ?, mls_number = ?, property_type = ?, bedrooms = ?, bathrooms = ?, 
		square_feet = ?, lot_size = ?, year_built = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, 
		property.Name, property.Location, property.Price, property.Description, property.Photos,
		property.ExternalID, property.MLSNumber, property.PropertyType,
		property.Bedrooms, property.Bathrooms, property.SquareFeet, property.LotSize, 
		property.YearBuilt, property.ID)
	return err
}

func (r *propertyRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM properties WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *propertyRepository) GetAll(ctx context.Context) ([]models.Property, error) {
	query := `SELECT id, name, location, price, description, photos, external_id, mls_number, 
		property_type, bedrooms, bathrooms, square_feet, lot_size, year_built, created_at, updated_at 
		FROM properties ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []models.Property
	for rows.Next() {
		var property models.Property
		if err := rows.Scan(&property.ID, &property.Name, &property.Location, &property.Price,
			&property.Description, &property.Photos, &property.ExternalID, &property.MLSNumber,
			&property.PropertyType, &property.Bedrooms, &property.Bathrooms, &property.SquareFeet,
			&property.LotSize, &property.YearBuilt, &property.CreatedAt, &property.UpdatedAt); err != nil {
			return nil, err
		}
		properties = append(properties, property)
	}
	return properties, nil
}