package handlers

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"triple-s/flags"
	"triple-s/models"
	"triple-s/storage"
)

func listBucketsHandler(w http.ResponseWriter, r *http.Request) {
	csvFilePath := filepath.Join(flags.StorageDir, "buckets.csv")
	file, err := os.Open(csvFilePath)
	if err != nil {
		log.Printf("Error opening CSV file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error reading buckets"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error reading buckets"})
		return
	}

	var buckets []models.Bucket
	for _, record := range records {
		if len(record) > 0 {
			creationDate, _ := time.Parse(time.RFC3339, record[1])
			lastModified, _ := time.Parse(time.RFC3339, record[3])
			buckets = append(buckets, models.Bucket{
				Name:          record[0],
				CreationDate:  creationDate,
				LastModified:  lastModified,
				ContentStatus: record[2],
			})
		}
	}

	result := models.ListAllMyBucketsResult{Buckets: buckets}
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	if err := xml.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Error encoding XML response: %v\n", err)
	}
}

func createBucketHandler(w http.ResponseWriter, r *http.Request, bucketName string) {
	if !isValidBucketName(bucketName) {
		w.WriteHeader(http.StatusBadRequest)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: "Invalid bucket name"})
		return
	}
	if isBucketExists(flags.StorageDir, bucketName) {
		w.WriteHeader(http.StatusConflict)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 409, Message: "Bucket already exists"})
		return
	}

	err := os.Mkdir(fmt.Sprintf("%s/%s", flags.StorageDir, bucketName), 0o755)
	if err != nil {
		log.Printf("Error creating bucket: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error creating bucket"})
		return
	}

	contentStatus := "inactive"
	csvdata := models.Bucket{
		Name:          bucketName,
		CreationDate:  time.Now(),
		ContentStatus: contentStatus,
		LastModified:  time.Now(),
	}

	err = storage.UpdateBucketCSV(bucketName, csvdata)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error updating metadata"})
		return
	}

	w.WriteHeader(http.StatusOK)
	xml.NewEncoder(w).Encode(models.SuccessResponse{Message: fmt.Sprintf("Bucket %s created successfully", bucketName)})
}

func isValidBucketName(bucketName string) bool {
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return false
	}
	if !islowercaseLetterorDigit(bucketName[0]) || !islowercaseLetterorDigit(bucketName[len(bucketName)-1]) {
		return false
	}

	for i := 0; i < len(bucketName); i++ {
		char := bucketName[i]
		if !(islowercaseLetterorDigit(char) || char == '-' || char == '.') {
			return false
		}
	}
	return true
}

func deleteBucketHandler(w http.ResponseWriter, r *http.Request, bucketName string) {
	bucketPath := filepath.Join(flags.StorageDir, bucketName)

	if !isBucketExists(flags.StorageDir, bucketName) {
		w.WriteHeader(http.StatusNotFound)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Bucket not found"})
		return
	}
	bucketExists, err := checkBucketInCSV(bucketName)
	if err != nil {
		log.Printf("Error checking object in objects.csv: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error checking object existence"})
		return
	}
	if !bucketExists {
		w.WriteHeader(http.StatusNotFound)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 404, Message: "Object not found in objects.csv"})
		return
	}
	isEmpty, err := storage.IsBucketEmptyIgnoringObjectsCSV(bucketPath)
	if err != nil {
		log.Printf("Error checking if bucket is empty: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error checking bucket status"})
		return
	}

	if !isEmpty {
		w.WriteHeader(http.StatusConflict)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 409, Message: "Bucket is not empty"})
		return
	}

	err = os.RemoveAll(bucketPath)
	if err != nil {
		log.Printf("Error deleting bucket %s: %v\n", bucketName, err)
		w.WriteHeader(http.StatusInternalServerError)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 500, Message: "Error deleting bucket"})
		return
	}

	storage.RemoveBucketCSV(bucketName)

	w.WriteHeader(http.StatusNoContent)
}

func isBucketEmpty(bucketPath string) (bool, error) {
	f, err := os.Open(bucketPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	entries, err := f.Readdir(1)
	if err == io.EOF {
		return true, nil
	} else if err != nil {
		return false, err
	}

	return len(entries) == 0, nil
}

func checkBucketInCSV(bucketName string) (bool, error) {
	csvFilePath := filepath.Join(flags.StorageDir, "buckets.csv")
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
		if record[0] == bucketName { // Assuming ObjectKey is the first column
			return true, nil
		}
	}

	return false, nil
}
