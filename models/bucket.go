package models

import (
	"encoding/xml"
	"time"
)

type Bucket struct {
	Name          string    `xml:"Name"`
	CreationDate  time.Time `xml:"CreationDate"`
	ContentStatus string    `xml:"ContentStatus"`
	LastModified  time.Time `xml:"LastModified"`
}

type ListAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

type ErrorResponse struct {
	XMLName xml.Name `xml:"error"`
	Status  int      `xml:"status"`
	Message string   `xml:"message"`
}

type SuccessResponse struct {
	XMLName xml.Name `xml:"success"`
	Message string   `xml:"message"`
}
