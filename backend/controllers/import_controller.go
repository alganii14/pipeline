package controllers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"pipeline-backend/config"
	"pipeline-backend/models"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	BatchSize     = 4000 // Bulk insert per 4,000 rows (Max safe: 65,535 ÷ 13 columns = 5,041, using 80% = 4,000)
	MaxWorkers    = 8    // Maximum parallel workers
	ChannelBuffer = 100  // Channel buffer size
)

type CSVRecord struct {
	PN       string
	NamaRMFT string
	KodeUker string
	KC       string
	Prod     string
	NoRek    string
	Dup      string
	Nama     string
	TGL      string
	Strategy string
	Segment  string
	Pipeline string
	Proyeksi float64
}

// Progress tracker untuk real-time update
type ImportProgress struct {
	TotalRows     int64   `json:"total_rows"`
	ImportedRows  int64   `json:"imported_rows"`
	FailedRows    int64   `json:"failed_rows"`
	Progress      float64 `json:"progress"`
	Speed         float64 `json:"speed"`
	ElapsedTime   string  `json:"elapsed_time"`
	EstimatedTime string  `json:"estimated_time"`
	IsCompleted   bool    `json:"is_completed"`
}

var (
	currentProgress ImportProgress
	progressMutex   sync.RWMutex
)

// GetImportProgress mengembalikan progress import saat ini
func GetImportProgress(c *fiber.Ctx) error {
	progressMutex.RLock()
	defer progressMutex.RUnlock()
	return c.JSON(currentProgress)
}

// ImportCSV handles CSV file upload and import with parallel processing
func ImportCSV(c *fiber.Ctx) error {
	startTime := time.Now()

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Check file extension
	if file.Header.Get("Content-Type") != "text/csv" &&
		file.Header.Get("Content-Type") != "application/vnd.ms-excel" {
		log.Println("File type:", file.Header.Get("Content-Type"))
		// Still allow, just warn
	}

	// Open file
	fileHandle, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer fileHandle.Close()

	// Create CSV reader
	reader := csv.NewReader(fileHandle)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Skip header row
	header, err := reader.Read()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid CSV file (no header)",
		})
	}

	log.Printf("📋 CSV Header: %v\n", header)
	log.Printf("📁 Starting import from file: %s (%.2f MB)\n", file.Filename, float64(file.Size)/(1024*1024))

	// Reset progress tracker
	progressMutex.Lock()
	currentProgress = ImportProgress{
		TotalRows:    0,
		ImportedRows: 0,
		FailedRows:   0,
		Progress:     0,
		IsCompleted:  false,
	}
	progressMutex.Unlock()

	// Create channels
	recordsChan := make(chan []CSVRecord, ChannelBuffer)
	errorsChan := make(chan error, MaxWorkers)
	doneChan := make(chan bool, MaxWorkers)

	var wg sync.WaitGroup
	var successCount, failedCount, totalRows int64
	var mu sync.Mutex

	// Start progress updater (update setiap detik)
	stopProgress := make(chan bool)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond) // Update setiap 0.5 detik
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				elapsed := time.Since(startTime).Seconds()
				speed := float64(successCount) / elapsed
				remaining := float64(totalRows-successCount) / speed

				progressMutex.Lock()
				currentProgress.ImportedRows = successCount
				currentProgress.FailedRows = failedCount
				currentProgress.TotalRows = totalRows
				if totalRows > 0 {
					currentProgress.Progress = float64(successCount) / float64(totalRows) * 100
				}
				currentProgress.Speed = speed
				currentProgress.ElapsedTime = fmt.Sprintf("%.1fs", elapsed)
				if speed > 0 && totalRows > successCount {
					currentProgress.EstimatedTime = fmt.Sprintf("%.1fs", remaining)
				} else {
					currentProgress.EstimatedTime = "calculating..."
				}
				progressMutex.Unlock()
				mu.Unlock()
			case <-stopProgress:
				return
			}
		}
	}()

	// Start workers
	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go worker(i, recordsChan, &successCount, &failedCount, &mu, &wg, errorsChan, doneChan)
	}

	// Read and batch CSV records
	go func() {
		batch := make([]CSVRecord, 0, BatchSize)
		rowCount := 0

		for {
			record, err := reader.Read()
			if err == io.EOF {
				// Update total rows
				mu.Lock()
				totalRows = int64(rowCount)
				mu.Unlock()

				// Send remaining batch
				if len(batch) > 0 {
					recordsChan <- batch
				}
				close(recordsChan)
				break
			}

			if err != nil {
				log.Printf("⚠️  Error reading row %d: %v\n", rowCount+1, err)
				mu.Lock()
				failedCount++
				totalRows = int64(rowCount)
				mu.Unlock()
				continue
			}

			rowCount++

			// Update total rows immediately (estimasi sementara)
			mu.Lock()
			totalRows = int64(rowCount)
			mu.Unlock()

			// Parse CSV record
			csvRecord, err := parseCSVRecord(record)
			if err != nil {
				log.Printf("⚠️  Error parsing row %d: %v\n", rowCount+1, err)
				mu.Lock()
				failedCount++
				mu.Unlock()
				continue
			}

			batch = append(batch, csvRecord)

			// Send batch when it reaches BatchSize
			if len(batch) >= BatchSize {
				recordsChan <- batch
				log.Printf("📦 Sent batch %d (Total rows processed: %d)\n", rowCount/BatchSize, rowCount)
				batch = make([]CSVRecord, 0, BatchSize)
			}
		}

		log.Printf("✅ Finished reading CSV: %d rows\n", rowCount)
	}()

	// Wait for all workers to complete
	wg.Wait()
	close(doneChan)
	close(errorsChan)
	close(stopProgress) // Stop progress updater

	duration := time.Since(startTime)
	endTime := time.Now()
	total := successCount + failedCount
	speed := float64(successCount) / duration.Seconds()

	// Final progress update
	progressMutex.Lock()
	currentProgress.ImportedRows = successCount
	currentProgress.FailedRows = failedCount
	currentProgress.TotalRows = totalRows
	currentProgress.Progress = 100
	currentProgress.Speed = speed
	currentProgress.ElapsedTime = duration.String()
	currentProgress.EstimatedTime = "0s"
	currentProgress.IsCompleted = true
	progressMutex.Unlock()

	log.Printf("\n🎉 Import completed!\n")
	log.Printf("   ✅ Success: %d\n", successCount)
	log.Printf("   ❌ Failed: %d\n", failedCount)
	log.Printf("   📊 Total: %d\n", total)
	log.Printf("   ⏱️  Duration: %s\n", duration)
	log.Printf("   🚀 Speed: %.0f rows/second\n\n", speed)

	response := models.ImportResponse{
		Success:    int(successCount),
		Failed:     int(failedCount),
		Total:      int(total),
		Duration:   duration.String(),
		DurationMs: duration.Milliseconds(),
		Speed:      speed,
		Message:    "Import completed successfully",
		StartTime:  startTime.Format("2006-01-02 15:04:05"),
		EndTime:    endTime.Format("2006-01-02 15:04:05"),
	}

	return c.JSON(response)
}

// worker processes batches of records
func worker(id int, recordsChan <-chan []CSVRecord, successCount, failedCount *int64, mu *sync.Mutex, wg *sync.WaitGroup, errorsChan chan<- error, doneChan chan<- bool) {
	defer wg.Done()

	db := config.GetDB()
	batchNum := 0

	for records := range recordsChan {
		batchNum++
		batchStart := time.Now()

		// Convert CSVRecords to Pipeline models
		pipelines := make([]models.Pipeline, len(records))
		for i, record := range records {
			pipelines[i] = models.Pipeline{
				PN:       record.PN,
				NamaRMFT: record.NamaRMFT,
				KodeUker: record.KodeUker,
				KC:       record.KC,
				Prod:     record.Prod,
				NoRek:    record.NoRek,
				Dup:      record.Dup,
				Nama:     record.Nama,
				TGL:      record.TGL,
				Strategy: record.Strategy,
				Segment:  record.Segment,
				Pipeline: record.Pipeline,
				Proyeksi: record.Proyeksi,
			}
		}

		// Bulk insert with transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			return tx.CreateInBatches(pipelines, BatchSize).Error
		})

		if err != nil {
			log.Printf("❌ Worker %d - Batch %d failed: %v\n", id, batchNum, err)
			mu.Lock()
			*failedCount += int64(len(records))
			mu.Unlock()
			errorsChan <- err
		} else {
			mu.Lock()
			*successCount += int64(len(records))
			currentTotal := *successCount
			mu.Unlock()
			log.Printf("✅ Worker %d - Batch %d inserted %d records in %s | 📊 Total imported: %d\n",
				id, batchNum, len(records), time.Since(batchStart), currentTotal)
		}
	}

	doneChan <- true
}

// parseCSVRecord parses a CSV row into CSVRecord struct
func parseCSVRecord(record []string) (CSVRecord, error) {
	if len(record) < 13 {
		return CSVRecord{}, fmt.Errorf("invalid record length: expected 13 fields, got %d", len(record))
	}

	proyeksi, err := strconv.ParseFloat(record[12], 64)
	if err != nil {
		proyeksi = 0 // Default to 0 if parsing fails
	}

	return CSVRecord{
		PN:       record[0],
		NamaRMFT: record[1],
		KodeUker: record[2],
		KC:       record[3],
		Prod:     record[4],
		NoRek:    record[5],
		Dup:      record[6],
		Nama:     record[7],
		TGL:      record[8],
		Strategy: record[9],
		Segment:  record[10],
		Pipeline: record[11],
		Proyeksi: proyeksi,
	}, nil
}
