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

// func main() {
//     args := os.Args[1:]

//     if len(args) == 0 {
//         fmt.Println("Incorrect input! Try again.")
//         os.Exit(1)
//     }

//     // Create a map to store arguments
//     flags := make(map[string]string)
    
//     // Processing arguments
//     for i := 0; i < len(args); i++ {
//         if i+1 < len(args) && (args[i] == "--port" || args[i] == "--dir") {
//             flags[args[i]] = args[i+1]
//             i++ // Skip the next element since it has already been processed
//         } else if args[i] == "--help" {
//             cases.HelpCase(args)
//             return
//         } else {
//             fmt.Println("Unknown command:", args[i])
//             os.Exit(1)
//         }
//     }

//     // Check for mandatory flags
//     port, hasPort := flags["--port"]
//     dir, hasDir := flags["--dir"]

//     if !hasPort || !hasDir {
//         fmt.Println("Not enough arguments! Use: --port [port number] --dir [directory path]")
//         os.Exit(1)
//     }

//     fmt.Printf("Server started on port %s, data will be stored in directory %s\n", port, dir)

//     http.HandleFunc("/", helloHandler)

//     // Start the server
//     err := http.ListenAndServe(":"+port, nil)
//     if err != nil {
//         fmt.Printf("Error starting server: %s\n", err)
//         os.Exit(1)
//     }
// }
