package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"real-estate-manager/backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewPropertyRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock database: %v", err)
	}
	defer db.Close()

	repo := NewPropertyRepository(db)
	if repo == nil {
		t.Error("NewPropertyRepository() returned nil")
	}
}

func TestPropertyRepository_Create(t *testing.T) {
	tests := []struct {
		name          string
		property      *models.Property
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		errorMessage  string
		expectedID    int
	}{
		{
			name: "successful property creation",
			property: &models.Property{
				Name:     "Beautiful House",
				Location: "123 Main St, New York, NY",
				Price:    500000.00,
				Description: models.NullString{
					NullString: sql.NullString{
						String: "A beautiful 3-bedroom house",
						Valid:  true,
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO properties").
					WithArgs("Beautiful House", "123 Main St, New York, NY", 500000.00, 
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: false,
			expectedID:    1,
		},
		{
			name: "database error during insert",
			property: &models.Property{
				Name:     "Test House",
				Location: "456 Oak St",
				Price:    300000.00,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO properties").
					WillReturnError(errors.New("database connection failed"))
			},
			expectedError: true,
			errorMessage:  "database connection failed",
		},
		{
			name: "error getting last insert id",
			property: &models.Property{
				Name:     "Test House",
				Location: "456 Oak St",
				Price:    300000.00,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO properties").
					WillReturnResult(sqlmock.NewErrorResult(errors.New("last insert id error")))
			},
			expectedError: true,
			errorMessage:  "last insert id error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating mock database: %v", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			repo := NewPropertyRepository(db)
			err = repo.Create(context.Background(), tt.property)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if tt.property.ID != tt.expectedID {
					t.Errorf("Expected ID %d, got %d", tt.expectedID, tt.property.ID)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPropertyRepository_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		id             int
		setupMock      func(sqlmock.Sqlmock)
		expectedProp   *models.Property
		expectedError  bool
		errorMessage   string
	}{
		{
			name: "successful property retrieval",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "location", "price", "description", "photos", 
					"external_id", "mls_number", "property_type", "bedrooms", "bathrooms",
					"square_feet", "lot_size", "year_built", "created_at", "updated_at",
				}).AddRow(
					1, "Beautiful House", "123 Main St", 500000.00, 
					models.NullString{NullString: sql.NullString{String: "Beautiful house", Valid: true}},
					models.PhotoList{}, 
					models.NullString{}, models.NullString{}, models.NullString{},
					models.NullInt32{}, models.NullInt32{}, models.NullInt32{},
					models.NullString{}, models.NullInt32{},
					time.Now(), time.Now(),
				)
				mock.ExpectQuery("SELECT (.+) FROM properties WHERE id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedProp: &models.Property{
				ID:       1,
				Name:     "Beautiful House",
				Location: "123 Main St",
				Price:    500000.00,
			},
			expectedError: false,
		},
		{
			name: "property not found",
			id:   999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM properties WHERE id = ?").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedProp:  nil,
			expectedError: false,
		},
		{
			name: "database error",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM properties WHERE id = ?").
					WithArgs(1).
					WillReturnError(errors.New("database connection error"))
			},
			expectedProp:  nil,
			expectedError: true,
			errorMessage:  "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating mock database: %v", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			repo := NewPropertyRepository(db)
			prop, err := repo.GetByID(context.Background(), tt.id)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if tt.expectedProp == nil {
				if prop != nil {
					t.Error("Expected nil property but got one")
				}
			} else {
				if prop == nil {
					t.Error("Expected property but got nil")
				} else {
					if prop.ID != tt.expectedProp.ID {
						t.Errorf("Expected ID %d, got %d", tt.expectedProp.ID, prop.ID)
					}
					if prop.Name != tt.expectedProp.Name {
						t.Errorf("Expected Name %s, got %s", tt.expectedProp.Name, prop.Name)
					}
					if prop.Location != tt.expectedProp.Location {
						t.Errorf("Expected Location %s, got %s", tt.expectedProp.Location, prop.Location)
					}
					if prop.Price != tt.expectedProp.Price {
						t.Errorf("Expected Price %f, got %f", tt.expectedProp.Price, prop.Price)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPropertyRepository_Update(t *testing.T) {
	tests := []struct {
		name          string
		property      *models.Property
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful property update",
			property: &models.Property{
				ID:       1,
				Name:     "Updated House",
				Location: "456 Oak St, Boston, MA",
				Price:    750000.00,
				Description: models.NullString{
					NullString: sql.NullString{
						String: "An updated beautiful house",
						Valid:  true,
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE properties SET").
					WithArgs("Updated House", "456 Oak St, Boston, MA", 750000.00,
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: false,
		},
		{
			name: "database error during update",
			property: &models.Property{
				ID:       1,
				Name:     "Test House",
				Location: "123 Main St",
				Price:    500000.00,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE properties SET").
					WillReturnError(errors.New("update failed"))
			},
			expectedError: true,
			errorMessage:  "update failed",
		},
		{
			name: "property not found for update",
			property: &models.Property{
				ID:       999,
				Name:     "Non-existent House",
				Location: "Nowhere",
				Price:    100000.00,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE properties SET").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating mock database: %v", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			repo := NewPropertyRepository(db)
			err = repo.Update(context.Background(), tt.property)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPropertyRepository_Delete(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful property deletion",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM properties WHERE id = ?").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name: "database error during deletion",
			id:   1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM properties WHERE id = ?").
					WithArgs(1).
					WillReturnError(errors.New("delete operation failed"))
			},
			expectedError: true,
			errorMessage:  "delete operation failed",
		},
		{
			name: "property not found for deletion",
			id:   999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM properties WHERE id = ?").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating mock database: %v", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			repo := NewPropertyRepository(db)
			err = repo.Delete(context.Background(), tt.id)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPropertyRepository_GetAll(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(sqlmock.Sqlmock)
		expectedProps  []models.Property
		expectedError  bool
		errorMessage   string
	}{
		{
			name: "successful retrieval with multiple properties",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "location", "price", "description", "photos",
					"external_id", "mls_number", "property_type", "bedrooms", "bathrooms",
					"square_feet", "lot_size", "year_built", "created_at", "updated_at",
				}).AddRow(
					1, "House 1", "Location 1", 500000.00,
					models.NullString{}, models.PhotoList{},
					models.NullString{}, models.NullString{}, models.NullString{},
					models.NullInt32{}, models.NullInt32{}, models.NullInt32{},
					models.NullString{}, models.NullInt32{},
					time.Now(), time.Now(),
				).AddRow(
					2, "House 2", "Location 2", 750000.00,
					models.NullString{}, models.PhotoList{},
					models.NullString{}, models.NullString{}, models.NullString{},
					models.NullInt32{}, models.NullInt32{}, models.NullInt32{},
					models.NullString{}, models.NullInt32{},
					time.Now(), time.Now(),
				)
				mock.ExpectQuery("SELECT (.+) FROM properties ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedProps: []models.Property{
				{
					ID:       1,
					Name:     "House 1",
					Location: "Location 1",
					Price:    500000.00,
				},
				{
					ID:       2,
					Name:     "House 2",
					Location: "Location 2",
					Price:    750000.00,
				},
			},
			expectedError: false,
		},
		{
			name: "successful retrieval with empty list",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "location", "price", "description", "photos",
					"external_id", "mls_number", "property_type", "bedrooms", "bathrooms",
					"square_feet", "lot_size", "year_built", "created_at", "updated_at",
				})
				mock.ExpectQuery("SELECT (.+) FROM properties ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedProps: nil,
			expectedError: false,
		},
		{
			name: "database error during query",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM properties ORDER BY created_at DESC").
					WillReturnError(errors.New("database connection error"))
			},
			expectedProps: nil,
			expectedError: true,
			errorMessage:  "database connection error",
		},
		{
			name: "scan error during row processing",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "location", "price", "description", "photos",
					"external_id", "mls_number", "property_type", "bedrooms", "bathrooms",
					"square_feet", "lot_size", "year_built", "created_at", "updated_at",
				}).AddRow(
					"invalid_id", "House 1", "Location 1", 500000.00,
					models.NullString{}, models.PhotoList{},
					models.NullString{}, models.NullString{}, models.NullString{},
					models.NullInt32{}, models.NullInt32{}, models.NullInt32{},
					models.NullString{}, models.NullInt32{},
					time.Now(), time.Now(),
				)
				mock.ExpectQuery("SELECT (.+) FROM properties ORDER BY created_at DESC").
					WillReturnRows(rows)
			},
			expectedProps: nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating mock database: %v", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			repo := NewPropertyRepository(db)
			props, err := repo.GetAll(context.Background())

			if tt.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if tt.expectedProps == nil {
				if props != nil {
					t.Error("Expected nil properties but got some")
				}
			} else {
				if props == nil {
					t.Error("Expected properties but got nil")
				} else {
					if len(props) != len(tt.expectedProps) {
						t.Errorf("Expected %d properties, got %d", len(tt.expectedProps), len(props))
					}
					for i, expectedProp := range tt.expectedProps {
						if i < len(props) {
							if props[i].ID != expectedProp.ID {
								t.Errorf("Expected ID %d at index %d, got %d", expectedProp.ID, i, props[i].ID)
							}
							if props[i].Name != expectedProp.Name {
								t.Errorf("Expected Name %s at index %d, got %s", expectedProp.Name, i, props[i].Name)
							}
							if props[i].Location != expectedProp.Location {
								t.Errorf("Expected Location %s at index %d, got %s", expectedProp.Location, i, props[i].Location)
							}
							if props[i].Price != expectedProp.Price {
								t.Errorf("Expected Price %f at index %d, got %f", expectedProp.Price, i, props[i].Price)
							}
						}
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}
