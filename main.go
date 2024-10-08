package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"triple-s/cases"
)

//go run main.go --port 8080 --dir .

//curl -X POST "http://localhost:8080?filename=example.txt" --data-binary @example.txt

// http://localhost:8080

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintf(w, "processing GET request")

	case http.MethodPost:
		filename := r.URL.Query().Get("filename")
		fmt.Println("received filename:", filename)

		if filename == "" {
			http.Error(w, "filename is required", http.StatusBadRequest)
			return
		}

		filePath := "uploads/" + filename // You might want to use the 'dir' variable here
		fmt.Println("creating file at:", filePath)

		file, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "unable to create file: "+err.Error(), http.StatusInternalServerError)
			fmt.Println("error creating file:", err)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, r.Body)
		if err != nil {
			http.Error(w, "unable to write file: "+err.Error(), http.StatusInternalServerError)
			fmt.Println("error writing file:", err)
			return
		}

		fmt.Fprintf(w, "file is successfully uploaded")
		fmt.Println("file uploaded successfully.")

	default:
		http.Error(w, "method is not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Incorrect input! Try again.")
		os.Exit(1)
	}

	var port, dir string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			if i+1 < len(args) {
				port = args[i+1]
				i++
			} else {
				fmt.Println("Port value does not exist")
				os.Exit(1)
			}
		case "--dir":
			if i+1 < len(args) {
				dir = args[i+1]
				i++
			} else {
				fmt.Println("Directory does not exist")
				os.Exit(1)
			}
		case "--help":
			cases.HelpCase(args)
		default:
			fmt.Println("Unknown command.")
			os.Exit(1)
		}
	}

	if port == "" || dir == "" {
		fmt.Println("Not enough arguments! Example: --port /number of port/ --dir /path to dir/")
		os.Exit(1)
	}

	// Set the upload directory
	if err := os.MkdirAll(dir+"/uploads", os.ModePerm); err != nil {
		fmt.Println("Failed to create uploads directory:", err)
		os.Exit(1)
	}

	http.HandleFunc("/", Handler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
	

