package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"triple-s/file"
)

// ./triple-s --port 8080 --dir .
// http://localhost:8080

func validateBucketName(name string) bool {
	re := regexp.MustCompile(`^(?!.*[.-]{2})([a-z0-9]([a-z0-9.-]{1,61}[a-z0-9])?)$`)
	return len(name) >= 3 && len(name) <= 63 && re.MatchString(name) && !regexp.MustCompile(`^[.-]|[.-]$`).MatchString(name)
}

// Обработчик HTTP-запросов
func Handler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.URL.Path[1:]

	if bucketName == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	dirPath := fmt.Sprintf("data/%s", bucketName)

	switch r.Method {
	case http.MethodPut:
		if !validateBucketName(bucketName) {
			http.Error(w, "Invalid bucket name", http.StatusBadRequest)
			return
		}

		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			http.Error(w, "Bucket already exists", http.StatusConflict)
			return
		}

		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			http.Error(w, "Failed to create bucket", http.StatusInternalServerError)
			return
		}

		err := file.AppendBucketMetadata("data/buckets.csv", bucketName, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
		if err != nil {
			http.Error(w, "Failed to save bucket metadata", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Bucket '%s' created successfully", bucketName)

	case http.MethodGet:
		records, err := file.ReadBucketsMetadata("data/buckets.csv")
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

	case http.MethodDelete:
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			http.Error(w, "Bucket not found", http.StatusNotFound)
			return
		}

		// Проверка на наличие объектов в bucket
		if hasObjects(dirPath) {
			http.Error(w, "Bucket is not empty", http.StatusConflict)
			return
		}

		if err := os.RemoveAll(dirPath); err != nil {
			http.Error(w, "Failed to delete bucket", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Bucket '%s' deleted successfully", bucketName)

	default:
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
	}
}

// Проверка, есть ли объекты в bucket
func hasObjects(bucketPath string) bool {
	// Ваша логика для проверки наличия объектов
	return false // Пример, измените в зависимости от вашего кода
}

// Основная функция
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
