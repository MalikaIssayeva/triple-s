package main

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

// GET — это когда ты спрашиваешь: "Покажи мне данные!"
// POST — это когда ты отправляешь что-то, чтобы сервер это сохранил.
// PUT — это как сказать: "Создай или обнови что-то на сервере".
// DELETE — это как сказать: "Удали это!"
// # Создание нового bucket
// curl -i -X PUT http://localhost:8080/my-bucket

// # Получение списка всех buckets
// curl -i -X GET http://localhost:8080/

// # Удаление bucket
// curl -i -X DELETE http://localhost:8080/my-bucket

// ./triple-s --port 8080 --dir .
// http://localhost:8080

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received %s request for %s\n", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[1:], "/", 2)
	bucketName := parts[0]
	var objectKey string

	if len(parts) > 1 {
		objectKey = parts[1]
	}

	if bucketName == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	dirPath := fmt.Sprintf("data/%s", bucketName)

	switch r.Method {
	case http.MethodPut:
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

	case http.MethodGet:
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			http.Error(w, "Bucket not found", http.StatusNotFound)
			return
		}

		if objectKey == "" {
			http.Error(w, "Object key is required", http.StatusBadRequest)
			return
		}

		objectFilePath := filepath.Join(dirPath, objectKey)
		file, err := os.Open(objectFilePath)
		if err != nil {
			http.Error(w, "Object not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)

	case http.MethodDelete:
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			http.Error(w, "Bucket not found", http.StatusNotFound)
			return
		}

		if objectKey == "" {
			http.Error(w, "Object key is required", http.StatusBadRequest)
			return
		}

		objectFilePath := filepath.Join(dirPath, objectKey)
		if err := os.Remove(objectFilePath); err != nil {
			http.Error(w, "Failed to delete object", http.StatusInternalServerError)
			return
		}

		objectMetadataPath := filepath.Join(dirPath, "objects.csv")
		if err := bo.RemoveObjectMetadata(objectMetadataPath, objectKey); err != nil {
			http.Error(w, "Failed to update object metadata", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatalf("Incorrect input! Try again.")
	}

	var port, dir string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			if i+1 < len(args) {
				port = args[i+1]
				i++
			} else {
				log.Fatalf("Port value does not exist")
			}
		case "--dir":
			if i+1 < len(args) {
				dir = args[i+1]
				i++
			} else {
				log.Fatalf("Directory does not exist")
			}
		default:
			log.Fatalf("Unknown command.")
		}
	}

	if port == "" || dir == "" {
		log.Fatalf("Not enough arguments! Example: --port /number of port/ --dir /path to dir/")
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create base data directory: %v", err)
	}

	http.HandleFunc("/", Handler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}
