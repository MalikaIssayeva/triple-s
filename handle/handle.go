package handle

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"triple-s/bo"
	"triple-s/name"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received %s request for %s\n", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[1:], "/", 2)
	bucketName := parts[0]
	var objectKey string

	if len(parts) > 1 {
		objectKey = parts[1]
	}

	dirPath := fmt.Sprintf("data/%s", bucketName)

	switch r.Method {
	case http.MethodPut:
		handlePut(w, r, bucketName, objectKey, dirPath)
	case http.MethodGet:
		if objectKey == "" {
			handleGetBuckets(w)
		} else {
			handleGetObject(w, bucketName, objectKey, dirPath)
		}
	case http.MethodDelete:
		if objectKey == "" {
			handleDeleteBucket(w, bucketName, dirPath)
		} else {
			handleDeleteObject(w, bucketName, objectKey, dirPath)
		}
	default:
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetBuckets(w http.ResponseWriter) {
	records, err := bo.ReadBucketsMetadata("data/buckets.csv")
	if err != nil {
		http.Error(w, "Failed to read bucket metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<ListAllBucketsResult>\n  <Buckets>\n"))

	for _, record := range records {
		xmlData := fmt.Sprintf("    <Bucket>\n      <Name>%s</Name>\n      <CreationDate>%s</CreationDate>\n      <LastModified>%s</LastModified>\n    </Bucket>\n", record[0], record[1], record[2])
		w.Write([]byte(xmlData))
	}

	w.Write([]byte("  </Buckets>\n</ListAllBucketsResult>\n"))
}

func handlePut(w http.ResponseWriter, r *http.Request, bucketName, objectKey, dirPath string) {
	if !name.ValidateBucketName(bucketName) {
		http.Error(w, "Invalid bucket name", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			http.Error(w, "Failed to create bucket", http.StatusInternalServerError)
			return
		}

		err := bo.AppendBucketMetadata("data/buckets.csv", bucketName, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
		if err != nil {
			http.Error(w, "Failed to save bucket metadata", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Bucket '%s' created successfully\n", bucketName)
	}

	if objectKey == "" {
		http.Error(w, "Object key is required", http.StatusBadRequest)
		return
	}

	objectFilePath := filepath.Join(dirPath, objectKey)
	file, err := os.Create(objectFilePath)
	if err != nil {
		http.Error(w, "Failed to create object file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, "Failed to write object data", http.StatusInternalServerError)
		return
	}

	objectMetadataPath := filepath.Join(dirPath, "objects.csv")
	creationDate := time.Now().Format(time.RFC3339)

	err = bo.AppendObjectMetadata(objectMetadataPath, objectKey, creationDate)
	if err != nil {
		http.Error(w, "Failed to save object metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Object '%s' uploaded successfully", objectKey)
}

func handleGetObject(w http.ResponseWriter, bucketName, objectKey, dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		xmlData := fmt.Sprintf("<Error><Code>NoSuchBucket</Code><Message>Bucket '%s' not found</Message></Error>", bucketName)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(xmlData))
		return
	}

	objectFilePath := filepath.Join(dirPath, objectKey)
	if _, err := os.Stat(objectFilePath); os.IsNotExist(err) {
		xmlData := fmt.Sprintf("<Error>\n  <Code>NoSuchKey</Code>\n    <Message>Object '%s' not found in bucket '%s'</Message>\n</Error>\n", objectKey, bucketName)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(xmlData))
		return
	}

	file, err := os.Open(objectFilePath)
	if err != nil {
		http.Error(w, "Failed to open object", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if strings.HasSuffix(objectKey, ".txt") {
		w.Header().Set("Content-Type", "text/plain")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}

func handleDeleteBucket(w http.ResponseWriter, bucketName, dirPath string) {
	w.Header().Set("Content-Type", "application/xml")

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		xmlData := fmt.Sprintf("<Error><Code>NoSuchBucket</Code><Message>Bucket '%s' not found</Message></Error>", bucketName)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(xmlData))
		return
	}

	if bo.HasObjects(dirPath) {
		xmlData := fmt.Sprintf("<Error><Code>BucketNotEmpty</Code><Message>Bucket '%s' is not empty</Message></Error>", bucketName)
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(xmlData))
		return
	}

	if err := os.RemoveAll(dirPath); err != nil {
		xmlData := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to delete bucket '%s'</Message></Error>", bucketName)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(xmlData))
		return
	}

	if err := bo.RemoveBucketMetadata("data/buckets.csv", bucketName); err != nil {
		xmlData := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to remove metadata for bucket '%s'</Message></Error>", bucketName)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(xmlData))
		return
	}

	xmlData := fmt.Sprintf("    <DeleteBucketResult>\n        <BucketName>%s</BucketName>\n            <Message>Bucket '%s' deleted successfully</Message>\n    </DeleteBucketResult>\n", bucketName, bucketName)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xmlData))
}

func handleDeleteObject(w http.ResponseWriter, bucketName, objectKey, dirPath string) {
	objectFilePath := filepath.Join(dirPath, objectKey)
	if _, err := os.Stat(objectFilePath); os.IsNotExist(err) {
		xmlData := fmt.Sprintf("<Error><Code>NoSuchKey</Code><Message>Object '%s' not found in bucket '%s'</Message></Error>", objectKey, bucketName)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(xmlData))
		return
	}

	if err := os.Remove(objectFilePath); err != nil {
		xmlData := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to delete object '%s'</Message></Error>", objectKey)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(xmlData))
		return
	}

	objectMetadataPath := filepath.Join(dirPath, "objects.csv")
	if err := bo.RemoveObjectMetadata(objectMetadataPath, objectKey); err != nil {
		xmlData := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to update object metadata for '%s'</Message></Error>", objectKey)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(xmlData))
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
