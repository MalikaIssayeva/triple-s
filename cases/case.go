package cases

import (
	"fmt"
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
