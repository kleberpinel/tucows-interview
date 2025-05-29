package main

import (
	"database/sql"
	"log"
	"os"

	"real-estate-manager/backend/internal/handlers"
	"real-estate-manager/backend/internal/middleware"
	"real-estate-manager/backend/internal/repository"
	"real-estate-manager/backend/internal/services"
	"real-estate-manager/backend/pkg/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	loadEnvironment()
	validateJWTSecret()
	
	db := initializeDatabase()
	defer db.Close()

	repositories := initializeRepositories(db)
	services := initializeServices(repositories)
	handlers := initializeHandlers(repositories, services)

	router := setupRouter(handlers, services.AuthService)
	startServer(router)
}

func loadEnvironment() {
	// Load .env file in development
	if gin.Mode() != gin.ReleaseMode {
		if err := godotenv.Load(".env.dev"); err != nil {
			log.Println("No .env.dev file found, using environment variables")
		}
	}
}

func validateJWTSecret() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("Warning: JWT_SECRET not set, using default (insecure for production)")
	} else if len(jwtSecret) < 32 {
		log.Println("Warning: JWT_SECRET should be at least 32 characters long")
	}
}

func initializeDatabase() *sql.DB {
	// Database configuration from environment variables
	dbConfig := database.NewConfigFromEnv()

	// Create database if it doesn't exist
	if err := database.CreateDatabaseIfNotExists(dbConfig); err != nil {
		log.Fatal("Failed to create database:", err)
	}

	// Initialize database connection
	db, err := database.NewMySQLConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.RunMigrations(db, "./migrations"); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	return db
}

type Repositories struct {
	UserRepo     repository.UserRepository
	PropertyRepo repository.PropertyRepository
}

func initializeRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		UserRepo:     repository.NewUserRepository(db),
		PropertyRepo: repository.NewPropertyRepository(db),
	}
}

type Services struct {
	AuthService       *services.AuthService
	PropertyService   *services.PropertyService
	SimplyRETSService *services.SimplyRETSService
}

func initializeServices(repos *Repositories) *Services {
	return &Services{
		AuthService:       services.NewAuthService(repos.UserRepo),
		PropertyService:   services.NewPropertyService(repos.PropertyRepo),
		SimplyRETSService: services.NewSimplyRETSService(repos.PropertyRepo),
	}
}

type Handlers struct {
	AuthHandler       *handlers.AuthHandler
	PropertyHandler   *handlers.PropertyHandler
	SimplyRETSHandler *handlers.SimplyRETSHandler
}

func initializeHandlers(repos *Repositories, services *Services) *Handlers {
	return &Handlers{
		AuthHandler:       handlers.NewAuthHandler(repos.UserRepo),
		PropertyHandler:   handlers.NewPropertyHandler(services.PropertyService),
		SimplyRETSHandler: handlers.NewSimplyRETSHandler(services.SimplyRETSService),
	}
}

func setupRouter(handlers *Handlers, authService *services.AuthService) *gin.Engine {
	r := gin.Default()

	// CORS middleware for frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Static file serving for images
	r.Static("/images", "./uploads/images")

	setupAPIRoutes(r, handlers, authService)

	return r
}

func setupAPIRoutes(r *gin.Engine, handlers *Handlers, authService *services.AuthService) {
	api := r.Group("/api")
	{
		// Authentication routes
		api.POST("/register", handlers.AuthHandler.Register)
		api.POST("/login", handlers.AuthHandler.Login)

		// SimplyRETS integration routes (protected)
		simplyrets := api.Group("/simplyrets")
		simplyrets.Use(middleware.AuthMiddleware(authService))
		{
			simplyrets.POST("/process", handlers.SimplyRETSHandler.StartProcessing)
			simplyrets.GET("/jobs/:jobId/status", handlers.SimplyRETSHandler.GetJobStatus)
			simplyrets.DELETE("/jobs/:jobId", handlers.SimplyRETSHandler.CancelJob)
			simplyrets.GET("/health", handlers.SimplyRETSHandler.HealthCheck)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/properties", handlers.PropertyHandler.GetProperties)
			protected.GET("/properties/:id", handlers.PropertyHandler.GetProperty)
			protected.POST("/properties", handlers.PropertyHandler.CreateProperty)
			protected.PUT("/properties/:id", handlers.PropertyHandler.UpdateProperty)
			protected.DELETE("/properties/:id", handlers.PropertyHandler.DeleteProperty)
		}
	}
}

func startServer(router *gin.Engine) {
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}