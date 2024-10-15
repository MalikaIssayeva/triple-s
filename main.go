package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"triple-s/file"
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
	bucketName := r.URL.Path[1:] // Получаем имя bucket из URL

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
		if !name.ValidateBucketName(bucketName) {
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
	case http.MethodDelete:
		w.Header().Set("Content-Type", "application/xml")

		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			xmlData := fmt.Sprintf("<Error><Code>NoSuchBucket</Code><Message>Bucket '%s' not found</Message></Error>", bucketName)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(xmlData))
			return
		}

		if file.HasObjects(dirPath) {
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

		if err := file.RemoveBucketMetadata("data/buckets.csv", bucketName); err != nil {
			xmlData := fmt.Sprintf("<Error><Code>InternalError</Code><Message>Failed to remove metadata for bucket '%s'</Message></Error>", bucketName)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(xmlData))
			return
		}
		xmlData := fmt.Sprintf("    <DeleteBucketResult>\n        <BucketName>%s</BucketName>\n            <Message>Bucket '%s' deleted successfully</Message>\n    </DeleteBucketResult>\n", bucketName, bucketName)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(xmlData))

	default:
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
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
