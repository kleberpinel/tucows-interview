package repository

import (
	"database/sql"
	"real-estate-manager/backend/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(user *models.User) error {
	query := `
        INSERT INTO users (username, password, email, created_at, updated_at) 
        VALUES (?, ?, ?, NOW(), NOW())
    `

	result, err := r.db.Exec(query, user.Username, user.Password, user.Email)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Fix: Convert int64 to uint properly
	user.ID = uint(id)
	return nil
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
	query := `
        SELECT id, username, password, email, created_at, updated_at 
        FROM users 
        WHERE id = ?
    `

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	query := `
        SELECT id, username, password, email, created_at, updated_at 
        FROM users 
        WHERE username = ?
    `

	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) Update(user *models.User) error {
	query := `
        UPDATE users 
        SET username = ?, password = ?, email = ?, updated_at = NOW() 
        WHERE id = ?
    `

	_, err := r.db.Exec(query, user.Username, user.Password, user.Email, user.ID)
	return err
}

func (r *userRepository) Delete(id uint) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}