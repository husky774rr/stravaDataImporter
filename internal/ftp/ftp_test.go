package ftp

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFTPManager(t *testing.T) {
	// Create temporary CSV file for testing
	tempDir := t.TempDir()
	csvFile := filepath.Join(tempDir, "test_ftp.csv")

	csvContent := `date,ftp
2024-01-01,170
2024-08-29,191
2024-10-27,217
2025-02-05,248`

	err := os.WriteFile(csvFile, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	manager := NewFTPManager(csvFile)

	// Test LoadFTPData
	err = manager.LoadFTPData()
	if err != nil {
		t.Fatalf("LoadFTPData() error = %v", err)
	}

	records := manager.GetAllRecords()
	if len(records) != 4 {
		t.Errorf("Expected 4 records, got %d", len(records))
	}

	// Test GetFTPForDate
	testDate := time.Date(2024, 9, 15, 0, 0, 0, 0, time.UTC)
	ftp := manager.GetFTPForDate(testDate)
	if ftp != 191 {
		t.Errorf("Expected FTP 191 for date 2024-09-15, got %f", ftp)
	}

	// Test GetFTPForDate before any records
	earlyDate := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)
	ftp = manager.GetFTPForDate(earlyDate)
	if ftp != 0 {
		t.Errorf("Expected FTP 0 for early date, got %f", ftp)
	}

	// Test GetFTPForDate exact match
	exactDate := time.Date(2024, 10, 27, 0, 0, 0, 0, time.UTC)
	ftp = manager.GetFTPForDate(exactDate)
	if ftp != 217 {
		t.Errorf("Expected FTP 217 for exact date 2024-10-27, got %f", ftp)
	}
}

func TestFTPManagerEmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	csvFile := filepath.Join(tempDir, "empty_ftp.csv")

	err := os.WriteFile(csvFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty CSV file: %v", err)
	}

	manager := NewFTPManager(csvFile)
	err = manager.LoadFTPData()
	if err != nil {
		t.Fatalf("LoadFTPData() error = %v", err)
	}

	ftp := manager.GetCurrentFTP()
	if ftp != 0 {
		t.Errorf("Expected FTP 0 for empty file, got %f", ftp)
	}
}

func TestFTPManagerInvalidFile(t *testing.T) {
	manager := NewFTPManager("nonexistent.csv")
	err := manager.LoadFTPData()
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}
