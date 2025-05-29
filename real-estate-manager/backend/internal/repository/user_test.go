package repository

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"real-estate-manager/backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name          string
		user          *models.User
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		errorMessage  string
		expectedID    uint
	}{
		{
			name: "successful user creation",
			user: &models.User{
				Username: "testuser",
				Password: "hashedpassword",
				Email:    "test@example.com",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("testuser", "hashedpassword", "test@example.com").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: false,
			expectedID:    1,
		},
		{
			name: "database error during insert",
			user: &models.User{
				Username: "testuser",
				Password: "hashedpassword",
				Email:    "test@example.com",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("testuser", "hashedpassword", "test@example.com").
					WillReturnError(errors.New("database connection failed"))
			},
			expectedError: true,
			errorMessage:  "database connection failed",
		},
		{
			name: "error getting last insert id",
			user: &models.User{
				Username: "testuser",
				Password: "hashedpassword",
				Email:    "test@example.com",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("testuser", "hashedpassword", "test@example.com").
					WillReturnResult(sqlmock.NewErrorResult(errors.New("last insert id error")))
			},
			expectedError: true,
			errorMessage:  "last insert id error",
		},
		{
			name: "duplicate username constraint violation",
			user: &models.User{
				Username: "existinguser",
				Password: "hashedpassword",
				Email:    "existing@example.com",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("existinguser", "hashedpassword", "existing@example.com").
					WillReturnError(errors.New("UNIQUE constraint failed: users.username"))
			},
			expectedError: true,
			errorMessage:  "UNIQUE constraint failed: users.username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			userRepo := NewUserRepository(db)
			err = userRepo.Create(tt.user)

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
				if tt.user.ID != tt.expectedID {
					t.Errorf("expected user ID %d, got %d", tt.expectedID, tt.user.ID)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		userID        uint
		setupMock     func(sqlmock.Sqlmock)
		expectedUser  *models.User
		expectedError bool
		errorMessage  string
	}{
		{
			name:   "successful user retrieval",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "email", "created_at", "updated_at"}).
					AddRow(1, "testuser", "hashedpassword", "test@example.com", now, now)
				mock.ExpectQuery("SELECT id, username, password, email, created_at, updated_at FROM users WHERE id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedUser: &models.User{
				ID:        1,
				Username:  "testuser",
				Password:  "hashedpassword",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password, email, created_at, updated_at FROM users WHERE id = ?").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: true,
			errorMessage:  "sql: no rows in result set",
		},
		{
			name:   "database error",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password, email, created_at, updated_at FROM users WHERE id = ?").
					WithArgs(1).
					WillReturnError(errors.New("database connection failed"))
			},
			expectedUser:  nil,
			expectedError: true,
			errorMessage:  "database connection failed",
		},
		{
			name:   "scan error",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "email", "created_at", "updated_at"}).
					AddRow("invalid_id", "testuser", "hashedpassword", "test@example.com", now, now)
				mock.ExpectQuery("SELECT id, username, password, email, created_at, updated_at FROM users WHERE id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedUser:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			userRepo := NewUserRepository(db)
			user, err := userRepo.GetByID(tt.userID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.expectedUser != nil && user != nil {
					if user.ID != tt.expectedUser.ID || user.Username != tt.expectedUser.Username ||
						user.Email != tt.expectedUser.Email {
						t.Errorf("expected user %+v, got %+v", tt.expectedUser, user)
					}
				} else if tt.expectedUser != user {
					t.Errorf("expected user %+v, got %+v", tt.expectedUser, user)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		username      string
		setupMock     func(sqlmock.Sqlmock)
		expectedUser  *models.User
		expectedError bool
		errorMessage  string
	}{
		{
			name:     "successful user retrieval by username",
			username: "testuser",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "email", "created_at", "updated_at"}).
					AddRow(1, "testuser", "hashedpassword", "test@example.com", now, now)
				mock.ExpectQuery("SELECT id, username, password, email, created_at, updated_at FROM users WHERE username = ?").
					WithArgs("testuser").
					WillReturnRows(rows)
			},
			expectedUser: &models.User{
				ID:        1,
				Username:  "testuser",
				Password:  "hashedpassword",
				Email:     "test@example.com",
				CreatedAt: now,
				UpdatedAt: now,
			},
			expectedError: false,
		},
		{
			name:     "user not found by username",
			username: "nonexistent",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password, email, created_at, updated_at FROM users WHERE username = ?").
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: true,
			errorMessage:  "sql: no rows in result set",
		},
		{
			name:     "database error during username query",
			username: "testuser",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, password, email, created_at, updated_at FROM users WHERE username = ?").
					WithArgs("testuser").
					WillReturnError(errors.New("database connection failed"))
			},
			expectedUser:  nil,
			expectedError: true,
			errorMessage:  "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			userRepo := NewUserRepository(db)
			user, err := userRepo.GetByUsername(tt.username)

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
				if tt.expectedUser != nil && user != nil {
					if user.ID != tt.expectedUser.ID || user.Username != tt.expectedUser.Username ||
						user.Email != tt.expectedUser.Email {
						t.Errorf("expected user %+v, got %+v", tt.expectedUser, user)
					}
				} else if tt.expectedUser != user {
					t.Errorf("expected user %+v, got %+v", tt.expectedUser, user)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	tests := []struct {
		name          string
		user          *models.User
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful user update",
			user: &models.User{
				ID:       1,
				Username: "updateduser",
				Password: "newhashed",
				Email:    "updated@example.com",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users\s+SET username = \?, password = \?, email = \?, updated_at = NOW\(\)\s+WHERE id = \?`).
					WithArgs("updateduser", "newhashed", "updated@example.com", uint(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name: "database error during update",
			user: &models.User{
				ID:       1,
				Username: "updateduser",
				Password: "newhashed",
				Email:    "updated@example.com",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users\s+SET username = \?, password = \?, email = \?, updated_at = NOW\(\)\s+WHERE id = \?`).
					WithArgs("updateduser", "newhashed", "updated@example.com", uint(1)).
					WillReturnError(errors.New("database connection failed"))
			},
			expectedError: true,
			errorMessage:  "database connection failed",
		},
		{
			name: "user not found for update",
			user: &models.User{
				ID:       999,
				Username: "updateduser",
				Password: "newhashed",
				Email:    "updated@example.com",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users\s+SET username = \?, password = \?, email = \?, updated_at = NOW\(\)\s+WHERE id = \?`).
					WithArgs("updateduser", "newhashed", "updated@example.com", uint(999)).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
			expectedError: false, // Update doesn't return error for 0 affected rows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			userRepo := NewUserRepository(db)
			err = userRepo.Update(tt.user)

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

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint
		setupMock     func(sqlmock.Sqlmock)
		expectedError bool
		errorMessage  string
	}{
		{
			name:   "successful user deletion",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id = ?").
					WithArgs(uint(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: false,
		},
		{
			name:   "database error during deletion",
			userID: 1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id = ?").
					WithArgs(uint(1)).
					WillReturnError(errors.New("database connection failed"))
			},
			expectedError: true,
			errorMessage:  "database connection failed",
		},
		{
			name:   "user not found for deletion",
			userID: 999,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM users WHERE id = ?").
					WithArgs(uint(999)).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected
			},
			expectedError: false, // Delete doesn't return error for 0 affected rows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.setupMock(mock)

			userRepo := NewUserRepository(db)
			err = userRepo.Delete(tt.userID)

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

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestNewUserRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userRepo := NewUserRepository(db)

	if userRepo == nil {
		t.Errorf("expected UserRepository instance, got nil")
		return
	}

	// Verify that the repository implements the interface
	var _ UserRepository = userRepo
}
