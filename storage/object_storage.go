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

func UpdateObjectMetadata(bucketName string, metadata models.ObjectCSV) error {
	csvPath := filepath.Join(flags.StorageDir, bucketName, "objects.csv")
	file, err := os.OpenFile(csvPath, os.O_RDWR|os.O_CREATE, 0o644)
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
		if len(record) > 0 && record[0] == metadata.ObjectKey {
			originalCreationDate := record[1]

			records[i] = []string{
				metadata.ObjectKey,
				originalCreationDate,
				metadata.ContentType,
				metadata.LastModified.Format(time.RFC3339),
			}
			updated = true
			break
		}
	}

	if !updated {
		records = append(records, []string{
			metadata.ObjectKey,
			fmt.Sprintf("%d", metadata.ObjectSize),
			metadata.ContentType,
			metadata.LastModified.Format(time.RFC3339),
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

func RemoveObjectMetadata(bucketName, objectKey string) error {
	csvPath := filepath.Join(flags.StorageDir, bucketName, "objects.csv")
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
		if record[0] != objectKey {
			writer.Write(record)
		}
	}
	writer.Flush()
	deleteCSV(csvPath)
	return writer.Error()
}

func deleteCSV(csvPath string) error {
	isEmpty, err := IsBucketEmptyIgnoringObjectsCSV(csvPath)
	if err != nil {
		return fmt.Errorf("error checking bucket: %v", err)
	}

	if !isEmpty {
		return fmt.Errorf("bucket is not empty")
	}

	err = os.RemoveAll(csvPath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}
