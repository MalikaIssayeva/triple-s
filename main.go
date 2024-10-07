package main

import (
	"fmt"
	"net/http"
	"os"

	"triple-s/cases"
)

// http://localhost:8080
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
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
				fmt.Println("port value does not exist")
				os.Exit(1)
			}
		case "--dir":
			if i+1 < len(args) {
				dir = args[i+1]
				i++
			} else {
				fmt.Println("dir does not exist")
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

	http.HandleFunc("/", helloHandler)

	// Запускаем HTTP-сервер
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
