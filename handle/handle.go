package handle

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"triple-s/bo"
	"triple-s/name"
)

func sendXMLResponse(w http.ResponseWriter, status int, data string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	w.Write([]byte(data))
}

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
		sendXMLResponse(w, http.StatusInternalServerError, "<Error><Code>InternalError</Code><Message>Failed to read bucket metadata</Message></Error>")
		return
	}

	uniqueBuckets := make(map[string]bool)
	response := "<ListAllBucketsResult>\n  <Buckets>\n"

	for _, record := range records {
		if !uniqueBuckets[record[0]] {
			response += fmt.Sprintf("    <Bucket>\n      <Name>%s</Name>\n      <CreationDate>%s</CreationDate>\n      <LastModified>%s</LastModified>\n    </Bucket>\n", record[0], record[1], record[2])
			uniqueBuckets[record[0]] = true
		}
	}

	response += "  </Buckets>\n</ListAllBucketsResult>\n"
	sendXMLResponse(w, http.StatusOK, response)
}

func handlePut(w http.ResponseWriter, r *http.Request, bucketName, objectKey, dirPath string) {
	if !name.ValidateBucketName(bucketName) {
		log.Printf("Invalid bucket name: %s", bucketName)
		http.Error(w, "Invalid bucket name", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			log.Printf("Failed to create bucket '%s': %v", bucketName, err)
			http.Error(w, "Failed to create bucket", http.StatusInternalServerError)
			return
		}

		err := bo.AppendBucketMetadata("data/buckets.csv", bucketName, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339), "active")
		if err != nil {
			log.Printf("Failed to save metadata for bucket '%s': %v", bucketName, err)
			http.Error(w, "Failed to save bucket metadata", http.StatusInternalServerError)
			return
		}

		log.Printf("Bucket '%s' created successfully", bucketName)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Bucket '%s' created successfully", bucketName)
		return
	}

	if objectKey == "" {
		http.Error(w, "Object key is required", http.StatusBadRequest)
		return
	}

	objectFilePath := filepath.Join(dirPath, objectKey)
	records, err := bo.ReadBucketsMetadata("data/buckets.csv")
	if err != nil {
		log.Printf("Failed to read bucket metadata: %v", err)
		http.Error(w, "Failed to read bucket metadata", http.StatusInternalServerError)
		return
	}
	log.Printf("Current bucket records: %v", records)
	file, err := os.Create(objectFilePath)
	if err != nil {
		http.Error(w, "Failed to create object file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err = io.Copy(file, r.Body); err != nil {
		http.Error(w, "Failed to write object data", http.StatusInternalServerError)
		return
	}

	fileInfo, err := os.Stat(objectFilePath)
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	size := fmt.Sprintf("%d", fileInfo.Size())
	contentType := r.Header.Get("Content-Type")
	lastModified := time.Now().Format(time.RFC3339)

	objectMetadataPath := filepath.Join(dirPath, "objects.csv")
	err = bo.AppendObjectMetadata(objectMetadataPath, objectKey, size, contentType, lastModified)
	if err != nil {
		http.Error(w, "Failed to save object metadata", http.StatusInternalServerError)
		return
	}

	err = bo.UpdateBucketLastModified("data/buckets.csv", bucketName, lastModified)
	if err != nil {
		log.Printf("Failed to update bucket metadata for '%s': %v", bucketName, err)
		http.Error(w, "Failed to update bucket metadata", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully uploaded object '%s' in bucket '%s'", objectKey, bucketName)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Object '%s' uploaded successfully", objectKey)
}

func handleGetObject(w http.ResponseWriter, bucketName, objectKey, dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Bucket '%s' not found", bucketName), http.StatusNotFound)
		return
	}

	objectFilePath := filepath.Join(dirPath, objectKey)
	if _, err := os.Stat(objectFilePath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Object '%s' not found in bucket '%s'", objectKey, bucketName), http.StatusNotFound)
		return
	}

	file, err := os.Open(objectFilePath)
	if err != nil {
		http.Error(w, "Failed to open object", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", getContentType(objectKey))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}

func getContentType(objectKey string) string {
	switch {
	case strings.HasSuffix(objectKey, ".txt"):
		return "text/plain"
	case strings.HasSuffix(objectKey, ".jpg"), strings.HasSuffix(objectKey, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(objectKey, ".png"):
		return "image/png"
	case strings.HasSuffix(objectKey, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(objectKey, ".json"):
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

func handleDeleteBucket(w http.ResponseWriter, bucketName, dirPath string) {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		response := fmt.Sprintf("<Error><Code>NoSuchBucket</Code><Message>Bucket '%s' not found</Message></Error>", bucketName)
		sendXMLResponse(w, http.StatusNotFound, response)
		return
	}

	if bo.HasObjects(dirPath) {
		response := fmt.Sprintf("<Error><Code>BucketNotEmpty</Code><Message>Bucket '%s' is not empty</Message></Error>", bucketName)
		sendXMLResponse(w, http.StatusConflict, response)
		return
	}

	if err := os.RemoveAll(dirPath); err != nil {
		response := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to delete bucket '%s'</Message></Error>", bucketName)
		sendXMLResponse(w, http.StatusInternalServerError, response)
		return
	}

	if err := bo.RemoveBucketMetadata("data/buckets.csv", bucketName); err != nil {
		response := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to remove metadata for bucket '%s'</Message></Error>", bucketName)
		sendXMLResponse(w, http.StatusInternalServerError, response)
		return
	}

	response := fmt.Sprintf("<DeleteBucketResult><BucketName>%s</BucketName><Message>Bucket '%s' deleted successfully</Message></DeleteBucketResult>", bucketName, bucketName)
	sendXMLResponse(w, http.StatusOK, response)
}

func handleDeleteObject(w http.ResponseWriter, bucketName, objectKey, dirPath string) {
	objectFilePath := filepath.Join(dirPath, objectKey)
	if _, err := os.Stat(objectFilePath); os.IsNotExist(err) {
		response := fmt.Sprintf("<Error><Code>NoSuchKey</Code><Message>Object '%s' not found in bucket '%s'</Message></Error>", objectKey, bucketName)
		sendXMLResponse(w, http.StatusNotFound, response)
		return
	}

	if err := os.Remove(objectFilePath); err != nil {
		response := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to delete object '%s'</Message></Error>", objectKey)
		sendXMLResponse(w, http.StatusInternalServerError, response)
		return
	}

	objectMetadataPath := filepath.Join(dirPath, "objects.csv")
	if err := bo.RemoveObjectMetadata(objectMetadataPath, objectKey); err != nil {
		response := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to update object metadata for '%s'</Message></Error>", objectKey)
		sendXMLResponse(w, http.StatusInternalServerError, response)
		return
	}

	sendXMLResponse(w, http.StatusNoContent, "") // 204 No Content
}
