package routes

import (
	"pipeline-backend/config"
	"pipeline-backend/controllers"
	"pipeline-backend/middleware"

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

	// Get database instance
	db := config.GetDB()

	// Auth routes (Public - tidak perlu JWT)
	authController := controllers.NewAuthController(db)
	auth := api.Group("/auth")
	auth.Post("/login", authController.Login)
	auth.Post("/register", authController.Register)

	// Protected routes - require JWT token
	protected := api.Group("/", middleware.JWTMiddleware())
	protected.Get("/profile", authController.GetProfile)
	protected.Post("/change-password", authController.ChangePassword)

	// Dashboard stats (protected)
	protected.Get("/stats", controllers.GetStats)

	// Initialize controllers
	di319Controller := controllers.NewDI319ImportController(db)

	// Pipeline routes (Protected - Import now uses DI319 logic with >= 50% filter)
	pipelines := protected.Group("/pipelines")
	pipelines.Get("/", controllers.GetPipelines)
	pipelines.Get("/search-rfmts", controllers.SearchRFMTs) // Must be before /:id
	pipelines.Post("/", controllers.CreatePipeline)
	pipelines.Post("/import", di319Controller.ImportCSV)                 // Import DI319 with auto-filter to pipelines
	pipelines.Get("/import/progress", di319Controller.GetImportProgress) // Get import progress
	pipelines.Delete("/all", controllers.DeleteAllPipelines)             // Delete all pipelines - MUST be before /:id
	pipelines.Get("/:id", controllers.GetPipeline)
	pipelines.Put("/:id", controllers.UpdatePipeline)
	pipelines.Delete("/:id", controllers.DeletePipeline)

	// RFMT routes (Protected)
	rfmtController := controllers.NewRFMTController(db)
	rfmts := protected.Group("/rfmts")
	rfmts.Get("/", rfmtController.GetAll)
	rfmts.Get("/search-ukers", rfmtController.SearchUkers) // Must be before /:id
	rfmts.Get("/:id", rfmtController.GetByID)
	rfmts.Post("/", rfmtController.Create)
	rfmts.Put("/:id", rfmtController.Update)
	rfmts.Delete("/:id", rfmtController.Delete)
	rfmts.Get("/pipeline/:pn", rfmtController.GetByPipelinePN)
	rfmts.Post("/import", rfmtController.ImportCSV)
	rfmts.Get("/import/progress", rfmtController.GetImportProgress)
	rfmts.Delete("/all", rfmtController.DeleteAll)

	// Uker routes (Protected)
	ukerController := controllers.NewUkerController(db)
	ukers := protected.Group("/ukers")
	ukers.Get("/", ukerController.GetAll)
	ukers.Get("/regions", ukerController.GetRegions)            // Must be before /:id
	ukers.Get("/types", ukerController.GetUkerTypes)            // Must be before /:id
	ukers.Get("/kode/:kode_uker", ukerController.GetByKodeUker) // Get by kode_uker
	ukers.Get("/:id", ukerController.GetByID)
	ukers.Post("/", ukerController.Create)
	ukers.Put("/:id", ukerController.Update)
	ukers.Delete("/:id", ukerController.Delete)

	// Product Type routes (Protected)
	productTypeController := controllers.NewProductTypeController(db)
	productTypes := protected.Group("/product-types")
	productTypes.Get("/", productTypeController.GetAll)
	productTypes.Get("/kode/:kode_product", productTypeController.GetByKodeProduct) // Must be before /:id
	productTypes.Get("/:id", productTypeController.GetByID)
	productTypes.Post("/", productTypeController.Create)
	productTypes.Put("/:id", productTypeController.Update)
	productTypes.Delete("/:id", productTypeController.Delete)

	// DI319 Import routes (Protected - same as pipeline import)
	di319 := protected.Group("/di319")
	di319.Get("/", di319Controller.GetAll)
	di319.Post("/import", di319Controller.ImportCSV)
	di319.Get("/import/progress", di319Controller.GetImportProgress)
	di319.Delete("/all", di319Controller.DeleteAll)
}
