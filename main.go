package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"triple-s/file"
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

func validateBucketName(name string) bool {
	if len(name) < 3 || len(name) > 63 {
		return false
	}

	// Check if it starts and ends with a letter or digit
	if !(isLowerLetterOrDigit(name[0]) && isLowerLetterOrDigit(name[len(name)-1])) {
		return false
	}

	// Check for valid characters
	re := regexp.MustCompile(`^[a-z0-9.-]+$`)
	if !re.MatchString(name) {
		return false
	}

	// Ensure it doesn't start with a dot or a hyphen
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "-") {
		return false
	}

	// Check for relative paths like './' or '../'
	if strings.Contains(name, "./") || strings.Contains(name, "../") {
		return false
	}

	// Check for adjacent periods or hyphens
	if strings.Contains(name, "..") || strings.Contains(name, "--") {
		return false
	}

	// Check for IP address format
	ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if ipRegex.MatchString(name) {
		return false
	}

	return true
}

// Helper function to check if a character is a lowercase letter or digit
func isLowerLetterOrDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'z')
}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received %s request for %s\n", r.Method, r.URL.Path)
	bucketName := r.URL.Path[1:] // Получаем имя bucket из URL

	// Если путь пуст, это запрос на получение списка всех buckets
	if bucketName == "" && r.Method == http.MethodGet {
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
		return
	}

	if bucketName == "" {
		http.Error(w, "Bucket name is required", http.StatusBadRequest)
		return
	}

	dirPath := fmt.Sprintf("data/%s", bucketName) // Путь к директории bucket

	switch r.Method {
	case http.MethodPut:
		fmt.Println("Validating bucket name")
		if !validateBucketName(bucketName) {
			http.Error(w, "Invalid bucket name", http.StatusBadRequest)
			return
		}

		fmt.Println("Checking if bucket exists")
		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			http.Error(w, "Bucket already exists", http.StatusConflict)
			return
		}

		fmt.Println("Creating bucket directory")
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			http.Error(w, "Failed to create bucket", http.StatusInternalServerError)
			return
		}

		fmt.Println("Appending bucket metadata")
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
		// xml format to change!!!!!
	case http.MethodDelete:
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			http.Error(w, "		<Error>\n		<Code>BucketNotEmpty</Code>\n		<Message>Bucket is not empty.</Message>\n		</Error>\n", http.StatusNotFound)
			return
		}

		if hasObjects(dirPath) {
			http.Error(w, "      <Error>\n      <Code>BucketNotEmpty</Code>\n      <Message>Bucket is not empty.</Message>\n     </Error>\n", http.StatusConflict)
			return
		}

		if err := os.RemoveAll(dirPath); err != nil {
			http.Error(w, "Failed to delete bucket", http.StatusInternalServerError)
			return
		}
		if err := removeBucketMetadata("data/buckets.csv", bucketName); err != nil {
			http.Error(w, "Failed to remove bucket metadata", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Bucket '%s' deleted successfully", bucketName)

	default:
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
	}
}

func hasObjects(bucketPath string) bool {
	files, err := os.ReadDir(bucketPath)
	if err != nil {
		return false
	}
	return len(files) > 0
}

func removeBucketMetadata(filePath, bucketName string) error {
	records, err := file.ReadBucketsMetadata(filePath)
	if err != nil {
		return err
	}

	var updatedRecords [][]string
	for _, record := range records {
		if record[0] != bucketName {
			updatedRecords = append(updatedRecords, record)
		}
	}

	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file for truncation: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range updatedRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("unable to write to file: %w", err)
		}
	}

	return nil
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
