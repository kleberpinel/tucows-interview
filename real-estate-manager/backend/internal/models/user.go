package models

import "time"

type User struct {
    ID        uint      `json:"id" db:"id"`
    Username  string    `json:"username" db:"username"`
    Password  string    `json:"password,omitempty" db:"password"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}