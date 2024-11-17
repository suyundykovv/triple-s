package models

import "time"

type ObjectCSV struct {
	ObjectKey    string // The key/name of the object
	ObjectSize   int64
	ContentType  string    // The MIME type of the object
	LastModified time.Time // The last modified time of the object
}
