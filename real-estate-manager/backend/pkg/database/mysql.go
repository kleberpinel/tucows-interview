package database

import (
    "database/sql"
    "fmt"

    _ "github.com/go-sql-driver/mysql"
)

type Config struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
}

func NewMySQLConnection(config Config) (*sql.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.User,
        config.Password,
        config.Host,
        config.Port,
        config.DBName,
    )

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Test the connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    // Set connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)

    return db, nil
}