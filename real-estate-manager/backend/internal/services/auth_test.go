package services

import (
	"errors"
	"os"
	"testing"
	"time"

	"real-estate-manager/backend/internal/mocks"
	"real-estate-manager/backend/internal/models"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	// Set up test JWT secret
	os.Setenv("JWT_SECRET", "test_secret_key_for_testing_purposes")
	defer os.Unsetenv("JWT_SECRET")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	tests := []struct {
		name           string
		user           models.User
		setupMock      func()
		expectedError  bool
		errorMessage   string
	}{
		{
			name: "successful registration",
			user: models.User{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			setupMock: func() {
				// User doesn't exist
				mockUserRepo.EXPECT().
					GetByUsername("testuser").
					Return(nil, errors.New("user not found"))
				
				// Create user successfully
				mockUserRepo.EXPECT().
					Create(gomock.Any()).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "user already exists",
			user: models.User{
				Username: "existinguser",
				Password: "password123",
				Email:    "existing@example.com",
			},
			setupMock: func() {
				existingUser := &models.User{
					ID:       1,
					Username: "existinguser",
					Email:    "existing@example.com",
				}
				mockUserRepo.EXPECT().
					GetByUsername("existinguser").
					Return(existingUser, nil)
			},
			expectedError: true,
			errorMessage:  "user already exists",
		},
		{
			name: "repository create error",
			user: models.User{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			setupMock: func() {
				// User doesn't exist
				mockUserRepo.EXPECT().
					GetByUsername("testuser").
					Return(nil, errors.New("user not found"))
				
				// Create user fails
				mockUserRepo.EXPECT().
					Create(gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedError: true,
			errorMessage:  "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			authService := NewAuthService(mockUserRepo)
			err := authService.Register(tt.user)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMessage {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	// Set up test JWT secret
	os.Setenv("JWT_SECRET", "test_secret_key_for_testing_purposes")
	defer os.Unsetenv("JWT_SECRET")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	// Create a hashed password for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		username      string
		password      string
		setupMock     func()
		expectedError bool
		errorMessage  string
		expectToken   bool
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "password123",
			setupMock: func() {
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Password: string(hashedPassword),
					Email:    "test@example.com",
				}
				mockUserRepo.EXPECT().
					GetByUsername("testuser").
					Return(user, nil)
			},
			expectedError: false,
			expectToken:   true,
		},
		{
			name:     "user not found",
			username: "nonexistent",
			password: "password123",
			setupMock: func() {
				mockUserRepo.EXPECT().
					GetByUsername("nonexistent").
					Return(nil, errors.New("user not found"))
			},
			expectedError: true,
			errorMessage:  "invalid credentials",
		},
		{
			name:     "invalid password",
			username: "testuser",
			password: "wrongpassword",
			setupMock: func() {
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Password: string(hashedPassword),
					Email:    "test@example.com",
				}
				mockUserRepo.EXPECT().
					GetByUsername("testuser").
					Return(user, nil)
			},
			expectedError: true,
			errorMessage:  "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			
			authService := NewAuthService(mockUserRepo)
			token, err := authService.Login(tt.username, tt.password)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMessage {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.expectToken && token == "" {
					t.Errorf("expected token but got empty string")
				}
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	// Set up test JWT secret
	testSecret := "test_secret_key_for_testing_purposes"
	os.Setenv("JWT_SECRET", testSecret)
	defer os.Unsetenv("JWT_SECRET")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	authService := NewAuthService(mockUserRepo)

	// Create a valid token for testing
	validClaims := jwt.MapClaims{
		"user_id":  uint(1),
		"username": "testuser",
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	}
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims)
	validTokenString, _ := validToken.SignedString([]byte(testSecret))

	// Create an expired token for testing
	expiredClaims := jwt.MapClaims{
		"user_id":  uint(1),
		"username": "testuser",
		"exp":      time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat":      time.Now().Add(-time.Hour * 2).Unix(),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte(testSecret))

	// Create a token with wrong signing method
	wrongMethodToken := jwt.NewWithClaims(jwt.SigningMethodRS256, validClaims)
	wrongMethodTokenString, _ := wrongMethodToken.SignedString([]byte(testSecret))

	tests := []struct {
		name          string
		tokenString   string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "valid token",
			tokenString:   validTokenString,
			expectedError: false,
		},
		{
			name:          "expired token",
			tokenString:   expiredTokenString,
			expectedError: true,
			errorMessage:  "invalid token",
		},
		{
			name:          "invalid token string",
			tokenString:   "invalid.token.string",
			expectedError: true,
			errorMessage:  "invalid token",
		},
		{
			name:          "empty token string",
			tokenString:   "",
			expectedError: true,
			errorMessage:  "invalid token",
		},
		{
			name:          "wrong signing method",
			tokenString:   wrongMethodTokenString,
			expectedError: true,
			errorMessage:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authService.ValidateToken(tt.tokenString)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMessage {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if claims == nil {
					t.Errorf("expected claims but got nil")
				} else {
					// Verify claims content
					if (*claims)["username"] != "testuser" {
						t.Errorf("expected username 'testuser', got '%v'", (*claims)["username"])
					}
					if (*claims)["user_id"] != float64(1) { // JSON numbers are float64
						t.Errorf("expected user_id 1, got '%v'", (*claims)["user_id"])
					}
				}
			}
		})
	}
}

func TestNewAuthService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	tests := []struct {
		name          string
		setupEnv      func()
		cleanupEnv    func()
		expectedSecret string
	}{
		{
			name: "with JWT_SECRET environment variable",
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "custom_secret_key")
			},
			cleanupEnv: func() {
				os.Unsetenv("JWT_SECRET")
			},
			expectedSecret: "custom_secret_key",
		},
		{
			name: "without JWT_SECRET environment variable",
			setupEnv: func() {
				os.Unsetenv("JWT_SECRET")
			},
			cleanupEnv: func() {},
			expectedSecret: "your_default_secret_key_change_this_in_production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			authService := NewAuthService(mockUserRepo)

			if authService == nil {
				t.Errorf("expected AuthService instance, got nil")
				return
			}

			if authService.userRepo != mockUserRepo {
				t.Errorf("expected userRepo to be set correctly")
			}

			if string(authService.jwtSecret) != tt.expectedSecret {
				t.Errorf("expected jwtSecret '%s', got '%s'", tt.expectedSecret, string(authService.jwtSecret))
			}
		})
	}
}
