package handlers

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"triple-s/flags"
	"triple-s/models"
	"triple-s/storage"
	"triple-s/utils"
)

func retrieveObjectHandler(w http.ResponseWriter, r *http.Request) {
	bucketName, objectKey, err := splitPath(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: "Invalid bucket or object key"})
		return
	}

	bucketPath := filepath.Join(flags.StorageDir, bucketName)
	if !isBucketExists(flags.StorageDir, bucketName) {
		w.WriteHeader(http.StatusNotFound)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Bucket not found"})
		return
	}
	csvPath := filepath.Join(bucketPath, "objects.csv")
	file, err := os.Open(csvPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error opening objects metadata file"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error reading objects metadata file"})
		return
	}

	objectExists := false
	for _, record := range records {
		if len(record) > 0 && record[0] == objectKey {
			objectExists = true
			break
		}
	}

	if !objectExists {
		w.WriteHeader(http.StatusNotFound)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Object not found in metadata"})
		return
	}

	objectPath := filepath.Join(bucketPath, objectKey)
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Object not found on filesystem"})
		return
	}

	content, err := os.ReadFile(objectPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Failed to read object data"})
		return
	}

	contentType := getObjectContentType(objectPath, content)
	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}

func getObjectContentType(objectPath string, content []byte) string {
	contentType := mime.TypeByExtension(filepath.Ext(objectPath))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	return contentType
}

func updateBucketStatus(bucketName, status string) error {
	csvPath := filepath.Join(flags.StorageDir, "buckets.csv")

	file, err := os.OpenFile(csvPath, os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("could not open buckets CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read buckets CSV: %w", err)
	}

	updated := false
	for i, record := range records {
		if len(record) > 0 && record[0] == bucketName {
			originalCreationDate := record[1]

			records[i] = []string{
				bucketName,
				originalCreationDate,
				status,
				time.Now().Format(time.RFC3339),
			}
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("bucket %s not found in metadata", bucketName)
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("could not truncate file: %w", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek to beginning of file: %w", err)
	}

	writer := csv.NewWriter(file)
	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("could not write updated CSV: %w", err)
	}
	writer.Flush()

	return writer.Error()
}

func uploadObjectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	bucketName, objectKey, err := splitPath(r.URL.Path)
	if objectKey == "objects.csv" {
		w.WriteHeader(http.StatusBadRequest)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: "Invalid object key"})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: "Invalid bucket or object key"})
		return
	}

	bucketPath := filepath.Join(flags.StorageDir, bucketName)
	if !isBucketExists(flags.StorageDir, bucketName) {
		w.WriteHeader(http.StatusNotFound)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Bucket not found"})
		return
	}

	if !utils.IsValidObjectKey(objectKey) {
		w.WriteHeader(http.StatusBadRequest)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: "Invalid object key"})
		return
	}

	content, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Failed to read object data"})
		return
	}
	defer r.Body.Close()

	objectPath := filepath.Join(bucketPath, objectKey)
	err = os.WriteFile(objectPath, content, 0o644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error saving object"})
		return
	}
	fileInfo, err := os.Stat(objectPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error retrieving file information"})
		return
	}

	contentType := r.Header.Get("Content-Type")
	csvdata := models.ObjectCSV{
		ObjectKey:    objectKey,
		ObjectSize:   fileInfo.Size(),
		ContentType:  contentType,
		LastModified: time.Now(),
	}

	err = storage.UpdateObjectMetadata(bucketName, csvdata)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error updating metadata"})
		return
	}

	err = updateBucketStatus(bucketName, "active")
	if err != nil {
		log.Printf("Failed to update bucket metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error updating bucket metadata"})
		return
	}

	updateBucketLastModified(bucketName)

	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(models.SuccessResponse{Message: fmt.Sprintf("Object %s uploaded successfully", objectKey)})
}

func deleteObjectHandler(w http.ResponseWriter, r *http.Request, bucketName, objectName string) {
	w.Header().Set("Content-Type", "application/xml")

	bucketPath := filepath.Join(flags.StorageDir, bucketName)
	if !isBucketExists(flags.StorageDir, bucketName) {
		w.WriteHeader(http.StatusNotFound)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Bucket not found"})
		return
	}

	objectExists, err := checkObjectInCSV(bucketName, objectName)
	if err != nil {
		log.Printf("Error checking object in objects.csv: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error checking object existence"})
		return
	}
	if !objectExists {
		w.WriteHeader(http.StatusNotFound)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Object not found in objects.csv"})
		return
	}

	objectPath := filepath.Join(bucketPath, objectName)
	err = os.Remove(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Object not found"})
		} else {
			log.Printf("Error deleting object: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error deleting object"})
		}
		return
	}

	isEmpty, err := storage.IsBucketEmptyIgnoringObjectsCSV(bucketPath)
	if err != nil {
		log.Printf("Failed to check if bucket is empty: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error checking bucket status"})
		return
	}

	newStatus := "inactive"
	if !isEmpty {
		newStatus = "active"
	}

	err = updateBucketStatus(bucketName, newStatus)
	if err != nil {
		log.Printf("Failed to update bucket metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error updating bucket metadata"})
		return
	}

	storage.RemoveObjectMetadata(bucketName, objectName)
	w.WriteHeader(http.StatusNoContent)
}

func checkObjectInCSV(bucketName, objectName string) (bool, error) {
	csvFilePath := filepath.Join(flags.StorageDir, bucketName, "objects.csv")
	file, err := os.Open(csvFilePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return false, err
	}

	for _, record := range records {
		if record[0] == objectName { // Assuming ObjectKey is the first column
			return true, nil
		}
	}

	return false, nil
}

func updateBucketLastModified(bucketName string) error {
	csvPath := filepath.Join(flags.StorageDir, "buckets.csv")
	file, err := os.OpenFile(csvPath, os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("could not open buckets CSV: %v", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("could not read buckets CSV: %v", err)
	}
	updated := false
	for i, record := range records {
		if len(record) > 0 && record[0] == bucketName {
			originalCreationDate := record[1]
			records[i] = []string{
				bucketName,
				originalCreationDate,
				record[2],
				time.Now().Format(time.RFC3339),
			}
			updated = true
			break
		}
	}
	if !updated {
		return fmt.Errorf("bucket %s not found in metadata", bucketName)
	}
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("could not truncate file: %v", err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek to beginning of file: %v", err)
	}
	writer := csv.NewWriter(file)
	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("could not write updated CSV: %v", err)
	}
	writer.Flush()
	return writer.Error()
}
