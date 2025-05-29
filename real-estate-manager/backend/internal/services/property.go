package services

import (
	"context"
	"errors"
	"real-estate-manager/backend/internal/models"
	"real-estate-manager/backend/internal/repository"
)

type PropertyService struct {
	repo repository.PropertyRepository
}

func NewPropertyService(repo repository.PropertyRepository) *PropertyService {
	return &PropertyService{repo: repo}
}

func (s *PropertyService) CreateProperty(ctx context.Context, property *models.Property) error {
	if err := validateProperty(property); err != nil {
		return err
	}
	return s.repo.Create(ctx, property)
}

func (s *PropertyService) GetProperty(ctx context.Context, id int) (*models.Property, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PropertyService) UpdateProperty(ctx context.Context, property *models.Property) error {
	if err := validateProperty(property); err != nil {
		return err
	}
	return s.repo.Update(ctx, property)
}

func (s *PropertyService) DeleteProperty(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *PropertyService) GetAllProperties(ctx context.Context) ([]models.Property, error) {
	return s.repo.GetAll(ctx)
}

func validateProperty(property *models.Property) error {
	if property == nil || property.Name == "" || property.Location == "" || property.Price <= 0 {
		return errors.New("invalid property data")
	}
	return nil
}