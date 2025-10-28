package controllers

import (
	"encoding/csv"
	"log"
	"pipeline-backend/models"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Import progress tracking for RFMT
var (
	rfmtImportProgress = struct {
		sync.RWMutex
		total     int64
		processed int64
		failed    int64
		status    string
		startTime time.Time
	}{}
)

// RFMTImportCSV handles CSV file upload and imports RFMT data
func (c *RFMTController) ImportCSV(ctx *fiber.Ctx) error {
	// Get uploaded file
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	// Open file
	fileHandle, err := file.Open()
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer fileHandle.Close()

	// Create CSV reader
	reader := csv.NewReader(fileHandle)
	reader.Comma = ';' // CSV uses semicolon as delimiter
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read header
	header, err := reader.Read()
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Failed to read CSV header"})
	}

	log.Printf("📋 CSV Header: %v", header)

	// Reset progress
	rfmtImportProgress.Lock()
	rfmtImportProgress.total = 0
	rfmtImportProgress.processed = 0
	rfmtImportProgress.failed = 0
	rfmtImportProgress.status = "processing"
	rfmtImportProgress.startTime = time.Now()
	rfmtImportProgress.Unlock()

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{"error": "Failed to read CSV data"})
	}

	totalRecords := int64(len(records))
	rfmtImportProgress.Lock()
	rfmtImportProgress.total = totalRecords
	rfmtImportProgress.Unlock()

	log.Printf("📊 Total records to import: %d", totalRecords)

	// Start import in background
	go c.processRFMTImport(records)

	return ctx.JSON(fiber.Map{
		"message": "Import started",
		"total":   totalRecords,
	})
}

// processRFMTImport processes the import in background using goroutines
func (c *RFMTController) processRFMTImport(records [][]string) {
	startTime := time.Now()
	numWorkers := 8
	batchSize := 4000

	log.Printf("🚀 Starting RFMT import with %d workers, batch size: %d", numWorkers, batchSize)

	// Channel for work distribution
	jobs := make(chan []models.RFMT, numWorkers*2)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for batch := range jobs {
				if err := c.insertRFMTBatch(batch); err != nil {
					log.Printf("❌ Worker %d: Batch insert failed: %v", workerID, err)
					atomic.AddInt64(&rfmtImportProgress.failed, int64(len(batch)))
				} else {
					atomic.AddInt64(&rfmtImportProgress.processed, int64(len(batch)))
					log.Printf("✅ Worker %d: Inserted %d records", workerID, len(batch))
				}
			}
		}(i)
	}

	// Prepare batches
	batch := make([]models.RFMT, 0, batchSize)
	for _, record := range records {
		if len(record) < 10 {
			atomic.AddInt64(&rfmtImportProgress.failed, 1)
			continue
		}

		// Parse CSV row to RFMT model
		// Expected format: PN;Nama Lengkap;JG;ESGDESC;Kanca;Uker;Uker Tujuan;Keterangan;Kelompok Jabatan RMFT Baru
		rfmt := models.RFMT{
			PN:                  record[0],
			NamaLengkap:         record[1],
			JG:                  record[2],
			ESGDESC:             record[3],
			Kanca:               record[4],
			Uker:                record[5],
			UkerTujuan:          record[6],
			Keterangan:          record[7],
			KelompokJabatanRMFT: record[8],
		}

		batch = append(batch, rfmt)

		// Send batch when it reaches the size limit
		if len(batch) >= batchSize {
			jobs <- batch
			batch = make([]models.RFMT, 0, batchSize)
		}
	}

	// Send remaining records
	if len(batch) > 0 {
		jobs <- batch
	}

	// Close jobs channel and wait for workers
	close(jobs)
	wg.Wait()

	// Update final status
	duration := time.Since(startTime)
	rfmtImportProgress.Lock()
	rfmtImportProgress.status = "completed"
	rfmtImportProgress.Unlock()

	log.Printf("✅ RFMT Import completed in %v", duration)
	log.Printf("📊 Processed: %d, Failed: %d", rfmtImportProgress.processed, rfmtImportProgress.failed)
}

// insertRFMTBatch inserts a batch of RFMT records
func (c *RFMTController) insertRFMTBatch(batch []models.RFMT) error {
	if len(batch) == 0 {
		return nil
	}

	// Use CreateInBatches for efficient batch insert
	return c.DB.CreateInBatches(batch, len(batch)).Error
}

// GetRFMTImportProgress returns the current import progress
func (c *RFMTController) GetImportProgress(ctx *fiber.Ctx) error {
	rfmtImportProgress.RLock()
	defer rfmtImportProgress.RUnlock()

	progress := float64(0)
	if rfmtImportProgress.total > 0 {
		progress = float64(rfmtImportProgress.processed) / float64(rfmtImportProgress.total) * 100
	}

	duration := time.Since(rfmtImportProgress.startTime)
	recordsPerSecond := float64(0)
	if duration.Seconds() > 0 {
		recordsPerSecond = float64(rfmtImportProgress.processed) / duration.Seconds()
	}

	return ctx.JSON(fiber.Map{
		"total":              rfmtImportProgress.total,
		"processed":          rfmtImportProgress.processed,
		"failed":             rfmtImportProgress.failed,
		"progress":           progress,
		"status":             rfmtImportProgress.status,
		"duration_seconds":   duration.Seconds(),
		"records_per_second": recordsPerSecond,
	})
}

// DeleteAllRFMTs deletes all RFMT records
func (c *RFMTController) DeleteAll(ctx *fiber.Ctx) error {
	// Count before delete
	var totalBefore int64
	c.DB.Model(&models.RFMT{}).Count(&totalBefore)

	// Hard delete all records (including soft deleted)
	result := c.DB.Unscoped().Where("1 = 1").Delete(&models.RFMT{})

	if result.Error != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to delete all RFMTs",
		})
	}

	// Reset auto increment
	c.DB.Exec("ALTER TABLE rfmts AUTO_INCREMENT = 1")

	return ctx.JSON(fiber.Map{
		"message":       "All RFMTs deleted successfully",
		"deleted_count": totalBefore,
	})
}
