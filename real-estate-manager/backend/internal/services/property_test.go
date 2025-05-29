package services

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"real-estate-manager/backend/internal/mocks"
	"real-estate-manager/backend/internal/models"

	"go.uber.org/mock/gomock"
)

func TestNewPropertyService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPropertyRepository(ctrl)
	service := NewPropertyService(mockRepo)

	if service == nil {
		t.Error("NewPropertyService() returned nil")
	}
	if service.repo != mockRepo {
		t.Error("NewPropertyService() did not set repository correctly")
	}
}

func TestPropertyService_CreateProperty(t *testing.T) {
	tests := []struct {
		name        string
		property    *models.Property
		setupMock   func(mock *mocks.MockPropertyRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful creation with valid property",
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
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectError: false,
		},
		{
			name:     "validation error - nil property",
			property: nil,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "validation error - empty name",
			property: &models.Property{
				Name:     "",
				Location: "123 Main St, New York, NY",
				Price:    500000.00,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "validation error - empty location",
			property: &models.Property{
				Name:     "Beautiful House",
				Location: "",
				Price:    500000.00,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "validation error - zero price",
			property: &models.Property{
				Name:     "Beautiful House",
				Location: "123 Main St, New York, NY",
				Price:    0,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "validation error - negative price",
			property: &models.Property{
				Name:     "Beautiful House",
				Location: "123 Main St, New York, NY",
				Price:    -100.00,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "repository error",
			property: &models.Property{
				Name:     "Beautiful House",
				Location: "123 Main St, New York, NY",
				Price:    500000.00,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			expectError: true,
			errorMsg:    "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			tt.setupMock(mockRepo)

			service := NewPropertyService(mockRepo)
			err := service.CreateProperty(context.Background(), tt.property)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPropertyService_GetProperty(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		setupMock     func(mock *mocks.MockPropertyRepository)
		expectedProp  *models.Property
		expectError   bool
		errorMsg      string
	}{
		{
			name: "successful retrieval",
			id:   1,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				prop := &models.Property{
					ID:       1,
					Name:     "Beautiful House",
					Location: "123 Main St, New York, NY",
					Price:    500000.00,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				mock.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(prop, nil).
					Times(1)
			},
			expectedProp: &models.Property{
				ID:       1,
				Name:     "Beautiful House",
				Location: "123 Main St, New York, NY",
				Price:    500000.00,
			},
			expectError: false,
		},
		{
			name: "property not found",
			id:   999,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, errors.New("property not found")).
					Times(1)
			},
			expectedProp: nil,
			expectError:  true,
			errorMsg:     "property not found",
		},
		{
			name: "repository error",
			id:   1,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, errors.New("database connection error")).
					Times(1)
			},
			expectedProp: nil,
			expectError:  true,
			errorMsg:     "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			tt.setupMock(mockRepo)

			service := NewPropertyService(mockRepo)
			prop, err := service.GetProperty(context.Background(), tt.id)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				if prop != nil {
					t.Error("Expected nil property on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
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
		})
	}
}

func TestPropertyService_UpdateProperty(t *testing.T) {
	tests := []struct {
		name        string
		property    *models.Property
		setupMock   func(mock *mocks.MockPropertyRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful update with valid property",
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
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectError: false,
		},
		{
			name:     "validation error - nil property",
			property: nil,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "validation error - empty name",
			property: &models.Property{
				ID:       1,
				Name:     "",
				Location: "456 Oak St, Boston, MA",
				Price:    750000.00,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "validation error - empty location",
			property: &models.Property{
				ID:       1,
				Name:     "Updated House",
				Location: "",
				Price:    750000.00,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "validation error - zero price",
			property: &models.Property{
				ID:       1,
				Name:     "Updated House",
				Location: "456 Oak St, Boston, MA",
				Price:    0,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				// No repository call expected
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "repository error",
			property: &models.Property{
				ID:       1,
				Name:     "Updated House",
				Location: "456 Oak St, Boston, MA",
				Price:    750000.00,
			},
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(errors.New("update failed")).
					Times(1)
			},
			expectError: true,
			errorMsg:    "update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			tt.setupMock(mockRepo)

			service := NewPropertyService(mockRepo)
			err := service.UpdateProperty(context.Background(), tt.property)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPropertyService_DeleteProperty(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		setupMock   func(mock *mocks.MockPropertyRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful deletion",
			id:   1,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Delete(gomock.Any(), 1).
					Return(nil).
					Times(1)
			},
			expectError: false,
		},
		{
			name: "property not found",
			id:   999,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Delete(gomock.Any(), 999).
					Return(errors.New("property not found")).
					Times(1)
			},
			expectError: true,
			errorMsg:    "property not found",
		},
		{
			name: "repository error",
			id:   1,
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					Delete(gomock.Any(), 1).
					Return(errors.New("delete operation failed")).
					Times(1)
			},
			expectError: true,
			errorMsg:    "delete operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			tt.setupMock(mockRepo)

			service := NewPropertyService(mockRepo)
			err := service.DeleteProperty(context.Background(), tt.id)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestPropertyService_GetAllProperties(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(mock *mocks.MockPropertyRepository)
		expectedProps  []models.Property
		expectError    bool
		errorMsg       string
	}{
		{
			name: "successful retrieval with multiple properties",
			setupMock: func(mock *mocks.MockPropertyRepository) {
				props := []models.Property{
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
				}
				mock.EXPECT().
					GetAll(gomock.Any()).
					Return(props, nil).
					Times(1)
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
			expectError: false,
		},
		{
			name: "successful retrieval with empty list",
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					GetAll(gomock.Any()).
					Return([]models.Property{}, nil).
					Times(1)
			},
			expectedProps: []models.Property{},
			expectError:   false,
		},
		{
			name: "repository error",
			setupMock: func(mock *mocks.MockPropertyRepository) {
				mock.EXPECT().
					GetAll(gomock.Any()).
					Return(nil, errors.New("database connection error")).
					Times(1)
			},
			expectedProps: nil,
			expectError:   true,
			errorMsg:      "database connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockPropertyRepository(ctrl)
			tt.setupMock(mockRepo)

			service := NewPropertyService(mockRepo)
			props, err := service.GetAllProperties(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
				if props != nil {
					t.Error("Expected nil properties on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
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
		})
	}
}

func TestValidateProperty(t *testing.T) {
	tests := []struct {
		name        string
		property    *models.Property
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid property",
			property: &models.Property{
				Name:     "Valid House",
				Location: "123 Main St",
				Price:    100000.00,
			},
			expectError: false,
		},
		{
			name:        "nil property",
			property:    nil,
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "empty name",
			property: &models.Property{
				Name:     "",
				Location: "123 Main St",
				Price:    100000.00,
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "empty location",
			property: &models.Property{
				Name:     "Valid House",
				Location: "",
				Price:    100000.00,
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "zero price",
			property: &models.Property{
				Name:     "Valid House",
				Location: "123 Main St",
				Price:    0,
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
		{
			name: "negative price",
			property: &models.Property{
				Name:     "Valid House",
				Location: "123 Main St",
				Price:    -1000.00,
			},
			expectError: true,
			errorMsg:    "invalid property data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProperty(tt.property)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
