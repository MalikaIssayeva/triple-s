package bucket

import (
	"fmt"
	"regexp"
	"sync"
)

var (
	bucketStore = make(map[string]struct{})
	mu          sync.Mutex
)

func CreateBucket(name string) error {
	mu.Lock()
	defer mu.Unlock()

	if err := validateBucketName(name); err != nil {
		return err
	}

	// Check if the bucket already exists
	if _, exists := bucketStore[name]; exists {
		return fmt.Errorf("bucket name already exists: %s", name)
	}

	// Create the bucket
	bucketStore[name] = struct{}{}
	return nil
}

func validateBucketName(name string) error {
	// Validate bucket name according to S3 rules
	if len(name) < 3 || len(name) > 63 {
		return fmt.Errorf("bucket name must be between 3 and 63 characters")
	}

	// Only lowercase letters, numbers, hyphens, and periods are allowed
	re := regexp.MustCompile(`^[a-z0-9.-]+$`)
	if !re.MatchString(name) {
		return fmt.Errorf("bucket name must consist of lowercase letters, numbers, hyphens, and periods only")
	}

	return nil
}


// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"regexp"
// 	"sync"

// 	"triple-s/cases"
// 	"triple-s/file"
// 	"triple-s/bucket"
// )

// var (
// 	bucketStore = make(map[string]struct{})
// 	mu          sync.Mutex
// )

// func Handler(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case http.MethodGet:
// 		fmt.Fprintf(w, "processing GET request")
// 	case http.MethodPost:
// 		filename := r.URL.Query().Get("filename")
// 		fmt.Println("received filename:", filename)

// 		if filename == "" {
// 			http.Error(w, "filename is required", http.StatusBadRequest)
// 			return
// 		}

// 		filePath := "uploads/" + filename

// 		if err := file.CreateFile(filePath, r.Body); err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			fmt.Println("error creating file:", err)
// 			return
// 		}

// 		fmt.Fprintf(w, "file is successfully uploaded")
// 		fmt.Println("file uploaded successfully.")
// 	case http.MethodPut:
// 		bucketName := r.URL.Path[1:] // Get bucket name from URL path
// 		if err := bucket.CreateBucket(bucketName); err != nil {
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}
// 		fmt.Fprintf(w, "Bucket '%s' created successfully", bucketName)
// 	default:
// 		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
// 	}
// }

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
// 		case "--help":
// 			cases.HelpCase(args)
// 		default:
// 			log.Fatalf("Unknown command.")
// 		}
// 	}

// 	if port == "" || dir == "" {
// 		log.Fatalf("Not enough arguments! Example: --port /number of port/ --dir /path to dir/")
// 	}

// 	if err := os.MkdirAll(dir+"/uploads", os.ModePerm); err != nil {
// 		log.Fatalf("Failed to create uploads directory: %v", err)
// 	}

// 	http.HandleFunc("/", Handler)

// 	err := http.ListenAndServe(":"+port, nil)
// 	if err != nil {
// 		log.Fatalf("Error starting server: %s\n", err)
// 	}
// }
