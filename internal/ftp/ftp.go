package ftp

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type FTPRecord struct {
	Date time.Time
	FTP  float64
}

type FTPManager struct {
	filePath string
	records  []FTPRecord
}

func NewFTPManager(filePath string) *FTPManager {
	return &FTPManager{
		filePath: filePath,
		records:  make([]FTPRecord, 0),
	}
}

func (f *FTPManager) LoadFTPData() error {
	file, err := os.Open(f.filePath)
	if err != nil {
		return fmt.Errorf("failed to open FTP file: %w", err)
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV: %w", err)
	}

	f.records = make([]FTPRecord, 0, len(records))

	for i, record := range records {
		if i == 0 {
			// Skip header row if it exists
			if len(record) >= 2 && record[0] == "date" {
				continue
			}
		}

		if len(record) < 2 {
			continue
		}

		date, err := time.Parse("2006-01-02", record[0])
		if err != nil {
			slog.Warn("Failed to parse date", "date", record[0], "error", err)
			continue
		}

		ftp, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			slog.Warn("Failed to parse FTP", "ftp", record[1], "error", err)
			continue
		}

		f.records = append(f.records, FTPRecord{
			Date: date,
			FTP:  ftp,
		})
	}

	slog.Info("Loaded FTP data", "records", len(f.records))
	return nil
}

func (f *FTPManager) GetFTPForDate(date time.Time) float64 {
	if len(f.records) == 0 {
		return 0
	}

	// Find the most recent FTP record before or on the given date
	var latestFTP float64
	var latestDate time.Time

	for _, record := range f.records {
		if record.Date.Before(date) || record.Date.Equal(date) {
			if latestDate.IsZero() || record.Date.After(latestDate) {
				latestDate = record.Date
				latestFTP = record.FTP
			}
		}
	}

	return latestFTP
}

func (f *FTPManager) GetCurrentFTP() float64 {
	return f.GetFTPForDate(time.Now())
}

func (f *FTPManager) GetAllRecords() []FTPRecord {
	return f.records
}
