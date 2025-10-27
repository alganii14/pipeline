package controllers

import (
	"math"
	"pipeline-backend/config"
	"pipeline-backend/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// GetPipelines returns all pipelines with pagination and filtering
func GetPipelines(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	search := c.Query("search", "")
	strategy := c.Query("strategy", "")
	segment := c.Query("segment", "")

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	db := config.GetDB()
	var pipelines []models.Pipeline
	var total int64

	query := db.Model(&models.Pipeline{})

	// Apply filters
	if search != "" {
		query = query.Where("pn LIKE ? OR nama LIKE ? OR kode_uker LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if strategy != "" {
		query = query.Where("strategy = ?", strategy)
	}
	if segment != "" {
		query = query.Where("segment = ?", segment)
	}

	// Get total count
	query.Count(&total)

	// Calculate offset
	offset := (page - 1) * perPage

	// Get paginated data
	result := query.Offset(offset).Limit(perPage).Order("created_at DESC").Find(&pipelines)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch pipelines",
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	response := models.PipelineResponse{
		Data:       pipelines,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}

	return c.JSON(response)
}

// GetPipeline returns a single pipeline by ID
func GetPipeline(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.GetDB()

	var pipeline models.Pipeline
	result := db.First(&pipeline, id)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Pipeline not found",
		})
	}

	return c.JSON(pipeline)
}

// CreatePipeline creates a new pipeline
func CreatePipeline(c *fiber.Ctx) error {
	pipeline := new(models.Pipeline)

	if err := c.BodyParser(pipeline); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	db := config.GetDB()
	result := db.Create(&pipeline)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create pipeline",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Pipeline created successfully",
		"data":    pipeline,
	})
}

// UpdatePipeline updates an existing pipeline
func UpdatePipeline(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.GetDB()

	var pipeline models.Pipeline
	result := db.First(&pipeline, id)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Pipeline not found",
		})
	}

	if err := c.BodyParser(&pipeline); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	db.Save(&pipeline)

	return c.JSON(fiber.Map{
		"message": "Pipeline updated successfully",
		"data":    pipeline,
	})
}

// DeletePipeline deletes a pipeline (soft delete)
func DeletePipeline(c *fiber.Ctx) error {
	id := c.Params("id")
	db := config.GetDB()

	var pipeline models.Pipeline
	result := db.First(&pipeline, id)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Pipeline not found",
		})
	}

	db.Delete(&pipeline)

	return c.JSON(fiber.Map{
		"message": "Pipeline deleted successfully",
	})
}

// DeleteAllPipelines deletes all pipelines (TRUNCATE)
func DeleteAllPipelines(c *fiber.Ctx) error {
	db := config.GetDB()

	// Count before delete
	var totalBefore int64
	db.Model(&models.Pipeline{}).Count(&totalBefore)

	// Hard delete all records (including soft deleted)
	result := db.Unscoped().Where("1 = 1").Delete(&models.Pipeline{})

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete all pipelines",
		})
	}

	// Reset auto increment
	db.Exec("ALTER TABLE pipelines AUTO_INCREMENT = 1")

	return c.JSON(fiber.Map{
		"message":       "All pipelines deleted successfully",
		"deleted_count": totalBefore,
	})
}

// GetStats returns statistics for dashboard
func GetStats(c *fiber.Ctx) error {
	db := config.GetDB()

	// Use a single query to get all stats at once for better performance
	type StatsResult struct {
		Total         int64
		TotalProyeksi float64
	}

	var result StatsResult
	db.Model(&models.Pipeline{}).
		Select("COUNT(*) as total, COALESCE(SUM(proyeksi), 0) as total_proyeksi").
		Scan(&result)

	// Get stats by strategy - optimized with index
	var strategyStats []struct {
		Strategy      string  `json:"strategy"`
		Count         int64   `json:"count"`
		TotalProyeksi float64 `json:"total_proyeksi"`
	}
	db.Model(&models.Pipeline{}).
		Select("strategy, COUNT(*) as count, COALESCE(SUM(proyeksi), 0) as total_proyeksi").
		Group("strategy").
		Scan(&strategyStats)

	// Get stats by segment - optimized with index
	var segmentStats []struct {
		Segment       string  `json:"segment"`
		Count         int64   `json:"count"`
		TotalProyeksi float64 `json:"total_proyeksi"`
	}
	db.Model(&models.Pipeline{}).
		Select("segment, COUNT(*) as count, COALESCE(SUM(proyeksi), 0) as total_proyeksi").
		Group("segment").
		Scan(&segmentStats)

	return c.JSON(fiber.Map{
		"total_pipelines": result.Total,
		"total_proyeksi":  result.TotalProyeksi,
		"strategy_stats":  strategyStats,
		"segment_stats":   segmentStats,
	})
}
