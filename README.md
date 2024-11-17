# Project: Triple-S - Simple Storage Service

## Overview

This project involves building a simple storage service (Triple-S) with basic features similar to Amazon S3. The service will support creating, listing, and deleting buckets, as well as uploading, retrieving, and deleting objects within those buckets. It will expose a RESTful API for interacting with the storage system. The backend will store metadata in CSV files and manage files in the filesystem.

### Key Features:

1. **HTTP Server**: Set up a basic HTTP server to handle incoming requests.
2. **Bucket Management**: Implement APIs for managing storage containers (buckets).
3. **Object Operations**: Implement APIs for uploading, retrieving, and deleting objects within buckets.
4. **Metadata Storage**: Use CSV files to store metadata for buckets and objects.

---

## Project Initialization

### Task: Set Up a Basic HTTP Server

Before implementing the full storage system, the first step is to create a basic HTTP server using Go’s `net/http` package. This server will serve as the foundation for the REST API functionalities.

### Requirements:
1. **Server Configuration**: The server must listen on a configurable port.
2. **Request Handling**: The server should handle incoming HTTP requests and provide appropriate responses.
3. **Error Handling**: Ensure the server handles errors gracefully, with proper logging and shutdown procedures.

### Arguments:
- **Port Number**: The server should accept a port number to run on.
- **Directory Path**: The server should accept a directory path where the files (buckets and objects) will be stored.

---

## Bucket Management

In this phase, you will implement functionalities for managing storage containers, or "buckets." Each bucket will store objects (files), and we will interact with them using REST API endpoints.

### API Endpoints for Bucket Management

#### 1. Create a Bucket

- **HTTP Method**: `PUT`
- **Endpoint**: `/buckets/{BucketName}`
- **Request Body**: Empty.
- **Behavior**:
    - Validate the bucket name to ensure it follows S3 naming rules.
    - Check if the bucket name is unique.
    - If valid and unique, create the bucket and store its metadata.
    - **Response**: 
      - `200 OK` if successful.
      - `400 Bad Request` for invalid names.
      - `409 Conflict` if the bucket already exists.

#### 2. List All Buckets

- **HTTP Method**: `GET`
- **Endpoint**: `/buckets`
- **Behavior**:
    - Read the bucket metadata from a CSV file.
    - Respond with an XML list of all buckets, including metadata like creation time, last modified time, etc.
    - **Response**: 
      - `200 OK` with XML data.

#### 3. Delete a Bucket

- **HTTP Method**: `DELETE`
- **Endpoint**: `/buckets/{BucketName}`
- **Behavior**:
    - Verify the existence of the bucket.
    - Ensure the bucket is empty before deletion.
    - **Response**: 
      - `204 No Content` if successful.
      - `404 Not Found` if the bucket doesn’t exist.
      - `409 Conflict` if the bucket is not empty.

### Bucket Naming Rules

- Bucket names must be unique across the system.
- Bucket names must be between 3 and 63 characters long.
- Only lowercase letters, numbers, hyphens, and periods are allowed.
- Names cannot be formatted as an IP address (e.g., `192.168.0.1`).
- Must not begin or end with a hyphen.
- Must not contain consecutive periods or hyphens.

Use regular expressions to enforce these rules.

### Example Scenario

1. **Create a Bucket**:  
   A client sends a `PUT` request to `/buckets/my-bucket`. The server checks the name, validates it, and creates the bucket in the metadata CSV file.

2. **List Buckets**:  
   A client sends a `GET` request to `/buckets`. The server reads the CSV file and responds with the list of all buckets.

3. **Delete a Bucket**:  
   A client sends a `DELETE` request to `/buckets/my-bucket`. The server checks if the bucket exists and is empty, then deletes it.

---

## Object Operations

Once buckets are managed, the next step is handling objects (files) within those buckets. You will implement APIs for uploading, retrieving, and deleting objects.

### API Endpoints for Object Operations

#### 1. Upload a New Object

- **HTTP Method**: `PUT`
- **Endpoint**: `/buckets/{BucketName}/objects/{ObjectKey}`
- **Request Body**: Binary data (file content).
- **Headers**:
    - `Content-Type`: The MIME type of the object (e.g., `image/png`).
    - `Content-Length`: The length of the file.
- **Behavior**:
    - Verify the bucket exists.
    - Validate the object key.
    - Save the object data to the disk in the appropriate bucket folder.
    - Update the metadata in the `objects.csv` file.
    - **Response**: 
      - `200 OK` on success.
      - Appropriate error messages for failures (e.g., `404 Not Found` if the bucket doesn’t exist).

#### 2. Retrieve an Object

- **HTTP Method**: `GET`
- **Endpoint**: `/buckets/{BucketName}/objects/{ObjectKey}`
- **Behavior**:
    - Check if the bucket exists.
    - Check if the object exists within the bucket.
    - Serve the object content if it exists, with the appropriate `Content-Type` header.
    - **Response**:
      - `200 OK` with object content if found.
      - `404 Not Found` if the object or bucket does not exist.

#### 3. Delete an Object

- **HTTP Method**: `DELETE`
- **Endpoint**: `/buckets/{BucketName}/objects/{ObjectKey}`
- **Behavior**:
    - Verify if both the bucket and the object exist.
    - Delete the object from the disk and update the `objects.csv` metadata.
    - **Response**:
      - `204 No Content` on success.
      - `404 Not Found` if the object or bucket does not exist.

### Example Scenarios

1. **Object Upload**:  
   A client sends a `PUT` request to `/buckets/photos/holiday.jpg` with the image data. The server stores the file in the `data/photos/` directory and updates the `objects.csv`.

2. **Object Retrieval**:  
   A client sends a `GET` request to `/buckets/photos/holiday.jpg`. The server checks if the file exists and returns the binary content.

3. **Object Deletion**:  
   A client sends a `DELETE` request to `/buckets/photos/holiday.jpg`. The server deletes the object from disk and removes the entry from `objects.csv`.

---

## Directory Structure

- **Data Directory**: The base directory for storing all data files.
  - `data/`: The main data folder.
    - `data/{bucket-name}/`: Subfolder for each bucket.
      - `objects.csv`: Metadata for the objects in that bucket.
      - Files (objects) stored inside the bucket folder.
      
## Usage Instructions

### Command-Line Options

- **--help**: Display usage information.
- **--port <N>**: The port number the server will listen on.
- **--dir <S>**: Path to the directory where data (buckets and objects) will be stored.

### Example Command:

To start the server on port `8080` and use `data/` as the base directory:

```bash
$ ./triple-s --port 8080 --dir /path/to/data
```

---

## Error Handling

Your server should handle various errors, such as:

- **Bucket Not Found**: Return `404 Not Found` if a bucket or object does not exist.
- **Invalid Bucket/Object Name**: Return `400 Bad Request` for invalid names or formats.
- **Conflict**: Return `409 Conflict` for actions that conflict with existing data (e.g., trying to create a bucket that already exists, or deleting a non-empty bucket).
- **Internal Server Errors**: Return `500 Internal Server Error` for unexpected server issues.

---

## Final Remarks

By following this structure, you will implement a robust and scalable storage service that can handle basic file storage, metadata management, and provide a foundation for building more advanced features later on.