package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"triple-s/flags"
	"triple-s/models"
)

func UpdateBucketCSV(bucketName string, updatedData models.Bucket) error {
	csvFilePath := filepath.Join(flags.StorageDir, "buckets.csv")
	file, err := os.OpenFile(csvFilePath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("could not open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read CSV file: %w", err)
	}

	updated := false
	for i, record := range records {
		if len(record) > 0 && record[0] == bucketName {
			records[i] = []string{
				updatedData.Name,
				updatedData.CreationDate.Format(time.RFC3339),
				updatedData.ContentStatus,
				updatedData.LastModified.Format(time.RFC3339),
			}
			updated = true
			break
		}
	}

	if !updated {
		records = append(records, []string{
			updatedData.Name,
			updatedData.CreationDate.Format(time.RFC3339),
			updatedData.ContentStatus,
			updatedData.LastModified.Format(time.RFC3339),
		})
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("could not truncate file: %w", err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek to beginning of file: %w", err)
	}

	writer := csv.NewWriter(file)
	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("could not write CSV file: %w", err)
	}
	writer.Flush()

	return nil
}

func RemoveBucketCSV(bucketName string) error {
	csvPath := filepath.Join(flags.StorageDir, "buckets.csv")
	file, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	file, err = os.OpenFile(csvPath, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	for _, record := range records {
		if record[0] != bucketName {
			writer.Write(record)
		}
	}
	writer.Flush()
	return writer.Error()
}

func IsBucketEmptyIgnoringObjectsCSV(bucketPath string) (bool, error) {
	files, err := os.ReadDir(bucketPath)
	if err != nil {
		return false, err
	}
	for _, file := range files {
		if file.Name() != "objects.csv" {
			return false, nil
		}
	}

	return true, nil
}
