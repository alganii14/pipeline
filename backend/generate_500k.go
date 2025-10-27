package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	startTime := time.Now()
	fmt.Println("🚀 Generating 500,000 rows CSV file...")

	// Create file
	file, err := os.Create("sample_500k.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"PN", "NAMA_RMFT", "KODE_UKER", "KC", "PROD", "NO_REK", "DUP", "NAMA", "TGL", "STRATEGY", "SEGMENT", "PIPELINE", "PROYEKSI"}
	writer.Write(header)

	// Data templates
	strategies := []string{
		"Kolaborasi Perusahaan Anak",
		"Optimalisasi Business Cluster",
		"Optimalisasi Digital Channel",
		"Optimalisasi Nasabah Prio, BOD, BOC",
		"Penguatan Produk & Fungsi RM",
		"Peningkatan Payroll Berkualitas",
		"Reaktivasi Rek Dormant",
		"Rekening Trx Debitur",
	}

	segments := []string{
		"KONSUMER",
		"Merchant",
		"Mikro",
		"Prioritas",
		"RITEL BADAN USAHA",
		"RITEL INDIVIDU",
		"RITEL NON INDIVIDU",
		"Ritel Perusahaan",
		"SME",
		"WEALTH",
	}

	cities := []string{
		"Jakarta", "Bandung", "Surabaya", "Medan", "Makassar",
		"Palembang", "Denpasar", "Semarang", "Yogyakarta", "Malang",
		"Bogor", "Tangerang", "Bekasi", "Depok", "Batam",
		"Balikpapan", "Pontianak", "Manado", "Padang", "Pekanbaru",
	}

	regions := []string{"Pusat", "Utara", "Selatan", "Timur", "Barat", "Tengah"}

	names := []string{
		"Budi Santoso", "Siti Nurhaliza", "Ahmad Dahlan", "Dewi Lestari", "Rizki Pratama",
		"Linda Wijaya", "Eko Prasetyo", "Maya Safitri", "Bambang Susilo", "Sari Indah",
		"Agus Setiawan", "Rina Wijaya", "Dedi Kurniawan", "Fitri Handayani", "Joko Widodo",
		"Ani Yudhoyono", "Hendra Gunawan", "Tuti Mariani", "Yanto Sukirman", "Mega Sari",
	}

	pipelines := []string{"Pipeline A", "Pipeline B", "Pipeline C"}

	// Seed random
	rand.Seed(time.Now().UnixNano())

	// Generate 500,000 rows
	totalRows := 500000
	batchSize := 10000

	for i := 1; i <= totalRows; i++ {
		city := cities[rand.Intn(len(cities))]
		region := regions[rand.Intn(len(regions))]

		// Generate random date in 2025
		month := rand.Intn(12) + 1
		day := rand.Intn(28) + 1
		date := fmt.Sprintf("2025-%02d-%02d", month, day)

		// Random proyeksi between 1 million to 10 million
		proyeksi := float64(rand.Intn(9000000)+1000000) + float64(rand.Intn(100))/100

		row := []string{
			fmt.Sprintf("PN%06d", i),                  // PN
			fmt.Sprintf("RMFT %s", city),              // NAMA_RMFT
			fmt.Sprintf("UK%06d", i),                  // KODE_UKER
			fmt.Sprintf("KC %s %s", city, region),     // KC
			fmt.Sprintf("PROD%03d", rand.Intn(999)+1), // PROD
			fmt.Sprintf("REK%09d", i),                 // NO_REK
			fmt.Sprintf("DUP%03d", rand.Intn(999)+1),  // DUP
			names[rand.Intn(len(names))],              // NAMA
			date,                                      // TGL
			strategies[rand.Intn(len(strategies))],    // STRATEGY
			segments[rand.Intn(len(segments))],        // SEGMENT
			pipelines[rand.Intn(len(pipelines))],      // PIPELINE
			fmt.Sprintf("%.2f", proyeksi),             // PROYEKSI
		}

		writer.Write(row)

		// Progress indicator
		if i%batchSize == 0 {
			fmt.Printf("✅ Generated %d / %d rows (%.1f%%)\n", i, totalRows, float64(i)/float64(totalRows)*100)
		}
	}

	duration := time.Since(startTime)
	fileInfo, _ := os.Stat("sample_500k.csv")
	fileSizeMB := float64(fileInfo.Size()) / 1024 / 1024

	fmt.Println("\n🎉 Generation completed!")
	fmt.Printf("   📊 Total rows: %d\n", totalRows)
	fmt.Printf("   📁 File size: %.2f MB\n", fileSizeMB)
	fmt.Printf("   ⏱️  Duration: %s\n", duration)
	fmt.Printf("   🚀 Speed: %.0f rows/second\n", float64(totalRows)/duration.Seconds())
}
