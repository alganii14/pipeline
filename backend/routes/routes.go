package routes

import (
	"pipeline-backend/controllers"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App) {
	// API group
	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Pipeline API is running",
		})
	})

	// Dashboard stats
	api.Get("/stats", controllers.GetStats)

	// Pipeline routes
	pipelines := api.Group("/pipelines")
	pipelines.Get("/", controllers.GetPipelines)
	pipelines.Post("/", controllers.CreatePipeline)
	pipelines.Post("/import", controllers.ImportCSV)
	pipelines.Get("/import/progress", controllers.GetImportProgress) // Get import progress
	pipelines.Delete("/all", controllers.DeleteAllPipelines)         // Delete all pipelines - MUST be before /:id
	pipelines.Get("/:id", controllers.GetPipeline)
	pipelines.Put("/:id", controllers.UpdatePipeline)
	pipelines.Delete("/:id", controllers.DeletePipeline)
}
