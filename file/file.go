package file

import (
	"fmt"
	"io"
	"os"
)

func CreateFile(filePath string, body io.Reader) error {
	fmt.Println("creating file at:", filePath)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, body); err != nil {
		return fmt.Errorf("unable to write file: %w", err)
	}

	return nil
}

