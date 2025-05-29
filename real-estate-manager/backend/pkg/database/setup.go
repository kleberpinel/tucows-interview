package database

import (
	"database/sql"
	"fmt"
)

// CreateDatabaseIfNotExists creates the database if it doesn't exist
func CreateDatabaseIfNotExists(config Config) error {
    // Connect without specifying database name
    tempConfig := config
    tempConfig.DBName = ""
    
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
        tempConfig.User,
        tempConfig.Password,
        tempConfig.Host,
        tempConfig.Port,
    )

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return fmt.Errorf("failed to connect to MySQL server: %w", err)
    }
    defer db.Close()

    // Create database if it doesn't exist
    _, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.DBName))
    if err != nil {
        return fmt.Errorf("failed to create database: %w", err)
    }

    return nil
}