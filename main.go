package main

import (
    "fmt"
    "net/http"
	"os"
	"triple-s/cases"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!")
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Incorrect input! Try again.")
		os.Exit(1)
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			fmt.Println("lol")
		case "--dir":
			fmt.Println("lol")
		case "--help" :
			cases.HelpCase(args)
		default:
			fmt.Println("Unknown command.")
			os.Exit(1)
		}
	}
}

