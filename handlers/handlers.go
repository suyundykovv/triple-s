package handlers

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strings"
	"triple-s/models"
)

func MyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")
	bucketName, objectName, err := splitPath(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: err.Error()})
		return
	}
	if bucketName == "" {
		switch r.Method {
		case http.MethodGet:
			listBucketsHandler(w, r)
		default:
			w.WriteHeader(http.StatusBadRequest)
			xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: "Bad request"})
		}
	} else if objectName == "" {
		switch r.Method {
		case http.MethodPut:
			createBucketHandler(w, r, bucketName)
		case http.MethodDelete:
			deleteBucketHandler(w, r, bucketName)
		default:
			w.WriteHeader(http.StatusBadRequest)
			xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: "Bad request"})
		}
	} else {
		switch r.Method {
		case http.MethodGet:
			retrieveObjectHandler(w, r)
		case http.MethodPut:
			uploadObjectHandler(w, r)
		case http.MethodDelete:
			deleteObjectHandler(w, r, bucketName, objectName)
		default:
			w.WriteHeader(http.StatusBadRequest)
			xml.NewEncoder(w).Encode(models.ErrorResponse{Status: 400, Message: "Bad request"})
		}
	}
}

func splitPath(path string) (string, string, error) {
	components := strings.Split(strings.Trim(path, "/"), "/")

	if len(components) == 1 {
		return components[0], "", nil
	} else if len(components) > 1 {
		return components[0], components[1], nil
	}
	return "", "", fmt.Errorf("invalid path, no bucket name provided")
}

func isBucketExists(baseDir, bucketName string) bool {
	bucketPath := fmt.Sprintf("%s/%s", baseDir, bucketName)
	_, err := os.Stat(bucketPath)
	return !os.IsNotExist(err)
}

func islowercaseLetterorDigit(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the triple-s storage service!")
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}
