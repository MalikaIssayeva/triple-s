package main

import (
	"log"
	"net/http"
	"os"

	"triple-s/handle"
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

	http.HandleFunc("/", handle.Handler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
}
