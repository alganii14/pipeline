package controllers

import (
	"pipeline-backend/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type RFMTController struct {
	DB *gorm.DB
}

func NewRFMTController(db *gorm.DB) *RFMTController {
	return &RFMTController{DB: db}
}

// GetAll - Get all RFMTs with pagination and filters
func (c *RFMTController) GetAll(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	search := ctx.Query("search", "")
	pn := ctx.Query("pn", "")

	offset := (page - 1) * limit

	var rfmts []models.RFMT
	var total int64

	query := c.DB.Model(&models.RFMT{}).Preload("Pipeline").Preload("UkerRelation")

	// Filter by PN if provided
	if pn != "" {
		query = query.Where("pn = ?", pn)
	}

	// Search filter
	if search != "" {
		query = query.Where("pn LIKE ? OR nama_lengkap LIKE ? OR jg LIKE ? OR kanca LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&rfmts).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch RFMTs"})
	}

	return ctx.JSON(fiber.Map{
		"data":  rfmts,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetByID - Get single RFMT by ID
func (c *RFMTController) GetByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var rfmt models.RFMT
	if err := c.DB.Preload("Pipeline").Preload("UkerRelation").First(&rfmt, id).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "RFMT not found"})
	}

	return ctx.JSON(rfmt)
}

// Create - Create new RFMT
func (c *RFMTController) Create(ctx *fiber.Ctx) error {
	rfmt := new(models.RFMT)

	if err := ctx.BodyParser(rfmt); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if rfmt.PN == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "PN is required"})
	}

	if err := c.DB.Create(&rfmt).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to create RFMT"})
	}

	// Load pipeline and uker relation if exists
	c.DB.Preload("Pipeline").Preload("UkerRelation").First(&rfmt, rfmt.ID)

	return ctx.Status(201).JSON(rfmt)
}

// Update - Update existing RFMT
func (c *RFMTController) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var rfmt models.RFMT
	if err := c.DB.First(&rfmt, id).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "RFMT not found"})
	}

	if err := ctx.BodyParser(&rfmt); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if rfmt.PN == "" {
		return ctx.Status(400).JSON(fiber.Map{"error": "PN is required"})
	}

	if err := c.DB.Save(&rfmt).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to update RFMT"})
	}

	// Load pipeline and uker relation if exists
	c.DB.Preload("Pipeline").Preload("UkerRelation").First(&rfmt, rfmt.ID)

	return ctx.JSON(rfmt)
}

// Delete - Soft delete RFMT
func (c *RFMTController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")

	var rfmt models.RFMT
	if err := c.DB.First(&rfmt, id).Error; err != nil {
		return ctx.Status(404).JSON(fiber.Map{"error": "RFMT not found"})
	}

	if err := c.DB.Delete(&rfmt).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to delete RFMT"})
	}

	return ctx.JSON(fiber.Map{"message": "RFMT deleted successfully"})
}

// GetByPipelinePN - Get all RFMTs for a specific pipeline by PN
func (c *RFMTController) GetByPipelinePN(ctx *fiber.Ctx) error {
	pn := ctx.Params("pn")

	var rfmts []models.RFMT
	if err := c.DB.Where("pn = ?", pn).Order("created_at DESC").Find(&rfmts).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to fetch RFMTs"})
	}

	return ctx.JSON(rfmts)
}

// SearchUkers - Search ukers for selection in RFMT form
func (c *RFMTController) SearchUkers(ctx *fiber.Ctx) error {
	search := ctx.Query("search", "")

	var ukers []models.Uker
	query := c.DB.Model(&models.Uker{}).Where("ACTIVE = ?", "Y")

	if search != "" {
		query = query.Where("kode_uker LIKE ? OR nama_uker LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Limit(50).Order("kode_uker ASC").Find(&ukers).Error; err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to search ukers"})
	}

	return ctx.JSON(ukers)
}
