package test

// import (
// 	"encoding/csv"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"regexp"
// 	"time"

// 	"triple-s/file"
// )

// // GET — это когда ты спрашиваешь: "Покажи мне данные!"
// // POST — это когда ты отправляешь что-то, чтобы сервер это сохранил.
// // PUT — это как сказать: "Создай или обнови что-то на сервере".
// // DELETE — это как сказать: "Удали это!"

// // ./triple-s --port 8080 --dir .
// // http://localhost:8080

// func validateBucketName(name string) bool {
// 	re := regexp.MustCompile(`^(?!.*[.-]{2})([a-z0-9]([a-z0-9.-]{1,61}[a-z0-9])?)$`)
// 	return len(name) >= 3 && len(name) <= 63 && re.MatchString(name) && !regexp.MustCompile(`^[.-]|[.-]$`).MatchString(name)
// }

// // Обработчик HTTP-запросов
// // Обработчик HTTP-запросов
// func Handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Printf("Received %s request for %s\n", r.Method, r.URL.Path)
// 	bucketName := r.URL.Path[1:] // Получаем имя bucket из URL

// 	if bucketName == "" {
// 		http.Error(w, "Bucket name is required", http.StatusBadRequest)
// 		return
// 	}

// 	dirPath := fmt.Sprintf("data/%s", bucketName) // Путь к директории bucket

// 	switch r.Method {
// 	case http.MethodPut:
// 		if !validateBucketName(bucketName) {
// 			http.Error(w, "Invalid bucket name", http.StatusBadRequest)
// 			return
// 		}

// 		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
// 			http.Error(w, "Bucket already exists", http.StatusConflict)
// 			return
// 		}

// 		// Создаем директорию для нового bucket
// 		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
// 			http.Error(w, "Failed to create bucket", http.StatusInternalServerError)
// 			return
// 		}

// 		// Сохраняем метаданные нового bucket в CSV
// 		err := file.AppendBucketMetadata("data/buckets.csv", bucketName, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
// 		if err != nil {
// 			http.Error(w, "Failed to save bucket metadata", http.StatusInternalServerError)
// 			return
// 		}

// 		fmt.Fprintf(w, "Bucket '%s' created successfully", bucketName)

// 	case http.MethodGet:
// 		records, err := file.ReadBucketsMetadata("data/buckets.csv")
// 		if err != nil {
// 			http.Error(w, "Failed to read bucket metadata", http.StatusInternalServerError)
// 			return
// 		}

// 		w.Header().Set("Content-Type", "application/xml")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("<ListAllBucketsResult>\n  <Buckets>\n"))

// 		for _, record := range records {
// 			xmlData := fmt.Sprintf("    <Bucket>\n      <Name>%s</Name>\n      <CreationDate>%s</CreationDate>\n      <LastModified>%s</LastModified>\n    </Bucket>\n", record[0], record[1], record[2])
// 			w.Write([]byte(xmlData))
// 		}

// 		w.Write([]byte("  </Buckets>\n</ListAllBucketsResult>\n"))

// 	case http.MethodDelete:
// 		// Проверяем, существует ли bucket
// 		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
// 			http.Error(w, "Bucket not found", http.StatusNotFound)
// 			return
// 		}

// 		// Проверяем, есть ли объекты в bucket
// 		if hasObjects(dirPath) {
// 			http.Error(w, "Bucket is not empty", http.StatusConflict)
// 			return
// 		}

// 		// Удаляем bucket
// 		if err := os.RemoveAll(dirPath); err != nil {
// 			http.Error(w, "Failed to delete bucket", http.StatusInternalServerError)
// 			return
// 		}

// 		// Удаляем метаданные bucket из CSV
// 		if err := removeBucketMetadata("data/buckets.csv", bucketName); err != nil {
// 			http.Error(w, "Failed to remove bucket metadata", http.StatusInternalServerError)
// 			return
// 		}

// 		fmt.Fprintf(w, "Bucket '%s' deleted successfully", bucketName)

// 	default:
// 		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
// 	}
// }

// func hasObjects(bucketPath string) bool {
// 	files, err := os.ReadDir(bucketPath)
// 	if err != nil {
// 		return false // Обработка ошибки при чтении директории
// 	}
// 	return len(files) > 0 // Возвращает true, если в bucket есть файлы
// }

// // Удаление метаданных из CSV-файла
// func removeBucketMetadata(filePath, bucketName string) error {
// 	records, err := file.ReadBucketsMetadata(filePath)
// 	if err != nil {
// 		return err
// 	}

// 	var updatedRecords [][]string
// 	for _, record := range records {
// 		if record[0] != bucketName { // Сохраняем все, кроме удаляемого
// 			updatedRecords = append(updatedRecords, record)
// 		}
// 	}

// 	// Перезаписываем файл с обновленными записями
// 	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_RDWR, 0o644)
// 	if err != nil {
// 		return fmt.Errorf("unable to open file for truncation: %w", err)
// 	}
// 	defer file.Close()

// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()

// 	for _, record := range updatedRecords {
// 		if err := writer.Write(record); err != nil {
// 			return fmt.Errorf("unable to write to file: %w", err)
// 		}
// 	}

// 	return nil
// }

// // Основная функция
// func main() {
// 	args := os.Args[1:]

// 	if len(args) == 0 {
// 		log.Fatalf("Incorrect input! Try again.")
// 	}

// 	var port, dir string

// 	for i := 0; i < len(args); i++ {
// 		switch args[i] {
// 		case "--port":
// 			if i+1 < len(args) {
// 				port = args[i+1]
// 				i++
// 			} else {
// 				log.Fatalf("Port value does not exist")
// 			}
// 		case "--dir":
// 			if i+1 < len(args) {
// 				dir = args[i+1]
// 				i++
// 			} else {
// 				log.Fatalf("Directory does not exist")
// 			}
// 		default:
// 			log.Fatalf("Unknown command.")
// 		}
// 	}

// 	if port == "" || dir == "" {
// 		log.Fatalf("Not enough arguments! Example: --port /number of port/ --dir /path to dir/")
// 	}

// 	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
// 		log.Fatalf("Failed to create base data directory: %v", err)
// 	}

// 	http.HandleFunc("/", Handler)

// 	err := http.ListenAndServe(":"+port, nil)
// 	if err != nil {
// 		log.Fatalf("Error starting server: %s\n", err)
// 	}
// }

// package file

// import (
// 	"encoding/csv"
// 	"fmt"
// 	"os"
// )

// // AppendBucketMetadata добавляет метаданные нового bucket в CSV файл.
// func AppendBucketMetadata(filePath, name, creationDate, lastModified string) error {
// 	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
// 	if err != nil {
// 		return fmt.Errorf("unable to open file: %w", err)
// 	}
// 	defer file.Close()

// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()

// 	if err := writer.Write([]string{name, creationDate, lastModified}); err != nil {
// 		return fmt.Errorf("unable to write to file: %w", err)
// 	}

// 	return nil
// }

// // ReadBucketsMetadata читает метаданные всех bucket'ов из CSV файла.
// func ReadBucketsMetadata(filePath string) ([][]string, error) {
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to open file: %w", err)
// 	}
// 	defer file.Close()

// 	reader := csv.NewReader(file)
// 	records, err := reader.ReadAll()
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to read file: %w", err)
// 	}

// 	return records, nil
// }
