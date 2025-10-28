package controllers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"pipeline-backend/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DI319ImportController struct {
	DB *gorm.DB
}

func NewDI319ImportController(db *gorm.DB) *DI319ImportController {
	return &DI319ImportController{DB: db}
}

var (
	di319ImportProgress     = 0
	di319ImportTotal        = 0
	di319ImportStatus       = "idle"
	di319ImportMessage      = ""
	di319ImportMutex        sync.Mutex
	di319PipelineCreated    = 0
	di319FilteredPercentage = 0.0
)

// ImportCSV - Import DI319 data and auto-filter to pipelines
func (c *DI319ImportController) ImportCSV(ctx *fiber.Ctx) error {
	// Check if import is already running
	di319ImportMutex.Lock()
	if di319ImportStatus == "processing" {
		di319ImportMutex.Unlock()
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Import already in progress",
		})
	}
	di319ImportStatus = "processing"
	di319ImportProgress = 0
	di319ImportTotal = 0
	di319ImportMessage = "Starting import..."
	di319PipelineCreated = 0
	di319FilteredPercentage = 0.0
	di319ImportMutex.Unlock()

	// Get file from request
	file, err := ctx.FormFile("file")
	if err != nil {
		di319ImportMutex.Lock()
		di319ImportStatus = "error"
		di319ImportMessage = "No file uploaded"
		di319ImportMutex.Unlock()
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Validate file type
	if !strings.HasSuffix(file.Filename, ".csv") {
		di319ImportMutex.Lock()
		di319ImportStatus = "error"
		di319ImportMessage = "File must be CSV"
		di319ImportMutex.Unlock()
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File must be CSV",
		})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		di319ImportMutex.Lock()
		di319ImportStatus = "error"
		di319ImportMessage = "Failed to open file"
		di319ImportMutex.Unlock()
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	// Parse CSV
	reader := csv.NewReader(src)
	reader.Comma = ';' // Semicolon delimiter

	// Read header
	header, err := reader.Read()
	if err != nil {
		di319ImportMutex.Lock()
		di319ImportStatus = "error"
		di319ImportMessage = "Failed to read CSV header"
		di319ImportMutex.Unlock()
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to read CSV header",
		})
	}

	log.Println("CSV Header:", header)

	// Process in background
	go func() {
		var di319Records []models.DI319
		var filteredCount int // Count records that meet ≥50% drop criteria
		lineNumber := 1

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("Error reading line %d: %v", lineNumber, err)
				lineNumber++
				continue
			}

			lineNumber++

			// Parse DI319 record - expecting 12 fields (excluding auto-generated ID)
			if len(record) < 12 {
				log.Printf("Skipping line %d: insufficient fields (got %d, expected 12)", lineNumber, len(record))
				continue
			}

			// Parse fields based on expected structure
			di319 := models.DI319{}

			// Periode (index 0)
			if periodeDate, err := time.Parse("2006-01-02", strings.TrimSpace(record[0])); err == nil {
				di319.Periode = periodeDate
			} else {
				log.Printf("Line %d: Invalid periode format: %s", lineNumber, record[0])
				continue
			}

			// Main Branch (index 1)
			di319.MainBranch = strings.TrimSpace(record[1])

			// Branch / kode_uker (index 2)
			di319.Branch = strings.TrimSpace(record[2])

			// CIF (index 3)
			di319.CIF = strings.TrimSpace(record[3])

			// NoRek (index 4)
			di319.NoRek = strings.TrimSpace(record[4])

			// Type (index 5)
			di319.Type = strings.TrimSpace(record[5])

			// Nama (index 6)
			di319.Nama = strings.TrimSpace(record[6])

			// PN Pengelola (index 7)
			di319.PNPengelola = strings.TrimSpace(record[7])

			// Balance (index 8)
			balanceStr := strings.ReplaceAll(strings.TrimSpace(record[8]), ",", "")
			if balance, err := strconv.ParseInt(balanceStr, 10, 64); err == nil {
				di319.Balance = balance
			} else {
				log.Printf("Line %d: Invalid balance: %s", lineNumber, record[8])
				continue
			}

			// Aval Balance (index 9)
			di319.AvalBalance = strings.TrimSpace(record[9])

			// Avg Balance - nullable (index 10)
			avgBalanceStr := strings.TrimSpace(record[10])
			if avgBalanceStr != "" && avgBalanceStr != "0" {
				di319.AvgBalance = &avgBalanceStr
			}

			// Open Date (index 11)
			if openDate, err := time.Parse("2006-01-02", strings.TrimSpace(record[11])); err == nil {
				di319.OpenDate = openDate
			} else {
				log.Printf("Line %d: Invalid open_date format: %s", lineNumber, record[11])
				continue
			}

			// ONLY save to DI319 if balance drop >= 50%
			if shouldCreatePipeline(di319) {
				di319Records = append(di319Records, di319)
				filteredCount++
			}

			// Batch insert every 1000 records
			if len(di319Records) >= 1000 {
				if err := c.DB.Create(&di319Records).Error; err != nil {
					log.Printf("Error inserting DI319 batch: %v", err)
				}

				di319ImportMutex.Lock()
				di319ImportProgress += len(di319Records)
				di319ImportMutex.Unlock()

				di319Records = []models.DI319{}
			}
		}

		// Insert remaining records
		if len(di319Records) > 0 {
			if err := c.DB.Create(&di319Records).Error; err != nil {
				log.Printf("Error inserting final DI319 batch: %v", err)
			}
			di319ImportMutex.Lock()
			di319ImportProgress += len(di319Records)
			di319ImportMutex.Unlock()
		}

		// Calculate filter percentage
		di319ImportMutex.Lock()
		di319ImportTotal = lineNumber - 1 // Total lines processed (excluding header)
		di319PipelineCreated = filteredCount // Records that met ≥50% criteria
		if di319ImportTotal > 0 {
			di319FilteredPercentage = (float64(di319ImportProgress) / float64(di319ImportTotal)) * 100
		}
		di319ImportStatus = "completed"
		di319ImportMessage = fmt.Sprintf("Import completed! Processed %d records, saved %d records with ≥50%% drop (%.2f%% filtered)",
			di319ImportTotal, di319ImportProgress, di319FilteredPercentage)
		di319ImportMutex.Unlock()

		log.Println(di319ImportMessage)
	}()

	return ctx.JSON(fiber.Map{
		"message": "Import started in background",
	})
}

// GetImportProgress - Get import progress
func (c *DI319ImportController) GetImportProgress(ctx *fiber.Ctx) error {
	di319ImportMutex.Lock()
	defer di319ImportMutex.Unlock()

	return ctx.JSON(fiber.Map{
		"status":            di319ImportStatus,
		"imported_rows":     di319ImportProgress, // Records saved (≥50% drop)
		"total_rows":        di319ImportTotal,    // Total records processed
		"message":           di319ImportMessage,
		"filtered_records":  di319ImportProgress, // Same as imported (only ≥50% saved)
		"filter_percentage": di319FilteredPercentage,
	})
}

// shouldCreatePipeline - Check if DI319 record should create pipeline entry
// Logic: Balance dropped 50% or more compared to average balance
func shouldCreatePipeline(di319 models.DI319) bool {
	// Try to parse avg_balance
	if di319.AvgBalance == nil || *di319.AvgBalance == "" {
		return false
	}

	avgBalanceStr := strings.ReplaceAll(*di319.AvgBalance, ",", "")
	avgBalance, err := strconv.ParseInt(avgBalanceStr, 10, 64)
	if err != nil {
		return false
	}

	// If avg_balance is 0, cannot calculate percentage
	if avgBalance == 0 {
		return false
	}

	// Calculate percentage drop
	// Formula: ((avg_balance - current_balance) / avg_balance) * 100
	dropPercentage := float64(avgBalance-di319.Balance) / float64(avgBalance) * 100

	// Return true if drop is 50% or more
	return dropPercentage >= 50.0
}

// DeleteAllDI319 - Delete all DI319 records
func (c *DI319ImportController) DeleteAll(ctx *fiber.Ctx) error {
	if err := c.DB.Exec("DELETE FROM di319").Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "All DI319 records deleted successfully",
	})
}

// GetAll - Get all DI319 records with pagination
func (c *DI319ImportController) GetAll(ctx *fiber.Ctx) error {
	var records []models.DI319
	var total int64

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size", "10"))
	offset := (page - 1) * pageSize

	search := ctx.Query("search", "")

	query := c.DB.Model(&models.DI319{})

	if search != "" {
		query = query.Where("branch LIKE ? OR nama LIKE ? OR norek LIKE ? OR cif LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	if err := query.Offset(offset).Limit(pageSize).Order("periode DESC, id DESC").Find(&records).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return ctx.JSON(fiber.Map{
		"data": records,
		"pagination": fiber.Map{
			"total_records": total,
			"total_pages":   totalPages,
			"current_page":  page,
			"page_size":     pageSize,
		},
	})
}
