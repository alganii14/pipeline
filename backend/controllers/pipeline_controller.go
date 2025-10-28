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

// GetStats returns statistics for dashboard (optimized version)
func GetStats(c *fiber.Ctx) error {
	db := config.GetDB()

	// Use raw SQL for maximum performance
	type StatsResult struct {
		Total         int64   `json:"total"`
		TotalProyeksi float64 `json:"total_proyeksi"`
	}

	type GroupStats struct {
		Name          string  `json:"name"`
		Count         int64   `json:"count"`
		TotalProyeksi float64 `json:"total_proyeksi"`
	}

	var result StatsResult
	var strategyStats []GroupStats
	var segmentStats []GroupStats

	// Execute all queries in parallel using goroutines
	done := make(chan bool, 3)

	// Query 1: Total count and sum
	go func() {
		db.Raw("SELECT COUNT(*) as total, COALESCE(SUM(proyeksi), 0) as total_proyeksi FROM pipelines WHERE deleted_at IS NULL").
			Scan(&result)
		done <- true
	}()

	// Query 2: Strategy stats
	go func() {
		db.Raw("SELECT strategy as name, COUNT(*) as count, COALESCE(SUM(proyeksi), 0) as total_proyeksi FROM pipelines WHERE deleted_at IS NULL GROUP BY strategy").
			Scan(&strategyStats)
		done <- true
	}()

	// Query 3: Segment stats
	go func() {
		db.Raw("SELECT segment as name, COUNT(*) as count, COALESCE(SUM(proyeksi), 0) as total_proyeksi FROM pipelines WHERE deleted_at IS NULL GROUP BY segment").
			Scan(&segmentStats)
		done <- true
	}()

	// Wait for all queries to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Transform strategy stats to match frontend format
	var formattedStrategyStats []struct {
		Strategy      string  `json:"strategy"`
		Count         int64   `json:"count"`
		TotalProyeksi float64 `json:"total_proyeksi"`
	}
	for _, s := range strategyStats {
		formattedStrategyStats = append(formattedStrategyStats, struct {
			Strategy      string  `json:"strategy"`
			Count         int64   `json:"count"`
			TotalProyeksi float64 `json:"total_proyeksi"`
		}{
			Strategy:      s.Name,
			Count:         s.Count,
			TotalProyeksi: s.TotalProyeksi,
		})
	}

	// Transform segment stats to match frontend format
	var formattedSegmentStats []struct {
		Segment       string  `json:"segment"`
		Count         int64   `json:"count"`
		TotalProyeksi float64 `json:"total_proyeksi"`
	}
	for _, s := range segmentStats {
		formattedSegmentStats = append(formattedSegmentStats, struct {
			Segment       string  `json:"segment"`
			Count         int64   `json:"count"`
			TotalProyeksi float64 `json:"total_proyeksi"`
		}{
			Segment:       s.Name,
			Count:         s.Count,
			TotalProyeksi: s.TotalProyeksi,
		})
	}

	return c.JSON(fiber.Map{
		"total_pipelines": result.Total,
		"total_proyeksi":  result.TotalProyeksi,
		"strategy_stats":  formattedStrategyStats,
		"segment_stats":   formattedSegmentStats,
	})
}

// SearchRFMTs - Search RFMTs by PN or Nama for Pipeline RMFT selection
func SearchRFMTs(c *fiber.Ctx) error {
	search := c.Query("search", "")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	db := config.GetDB()
	var rfmts []models.RFMT
	query := db.Model(&models.RFMT{})

	if search != "" {
		query = query.Where("pn LIKE ? OR nama_lengkap LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Limit(limit).Order("pn ASC").Find(&rfmts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to search RFMTs"})
	}

	return c.JSON(rfmts)
}
