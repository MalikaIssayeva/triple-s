package file

import (
	"encoding/csv"
	"fmt"
	"os"
)

// AppendBucketMetadata добавляет метаданные нового bucket в CSV файл.
func AppendBucketMetadata(filePath, name, creationDate, lastModified string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{name, creationDate, lastModified}); err != nil {
		return fmt.Errorf("unable to write to file: %w", err)
	}

	return nil
}

// ReadBucketsMetadata читает метаданные всех bucket'ов из CSV файла.
func ReadBucketsMetadata(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	return records, nil
}
