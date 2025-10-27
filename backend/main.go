package main

import (
	"log"
	"os"
	"pipeline-backend/config"
	"pipeline-backend/models"
	"pipeline-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to database
	config.ConnectDatabase()

	// Auto migrate database schema
	db := config.GetDB()
	err := db.AutoMigrate(&models.Pipeline{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("✅ Database migration completed!")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 1024, // 1GB for large CSV files
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Setup routes
	routes.SetupRoutes(app)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on http://localhost:%s\n", port)
	log.Printf("📊 API Documentation: http://localhost:%s/api/health\n", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
