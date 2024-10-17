package cases

import (
	"fmt"
	"log"
)

func HelpCase(args []string) {
	fmt.Println(`Simple Storage Service

Usage:
    triple-s [-port <N>] [-dir <S>]
    triple-s --help

Options:
    --help     Show this screen.
    --port N   Port number.
    --dir S    Path to the directory.`)
}

func ParseArgs(args []string) (string, string) {
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
		case "--help":
			HelpCase(args)
		default:
			log.Fatalf("Unknown command: %s", args[i])
		}
	}

	if port == "" || dir == "" {
		log.Fatalf("Not enough arguments! Example: --port <port> --dir <path>")
	}

	return port, dir
}
