package bo

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

// HasObjects проверяет, есть ли объекты в bucket.
func HasObjects(bucketPath string) bool {
	files, err := os.ReadDir(bucketPath)
	if err != nil {
		return false
	}
	return len(files) > 0
}

// RemoveBucketMetadata удаляет метаданные bucket из CSV файла.
func RemoveBucketMetadata(filePath, bucketName string) error {
	records, err := ReadBucketsMetadata(filePath)
	if err != nil {
		return err
	}

	var updatedRecords [][]string
	for _, record := range records {
		if record[0] != bucketName {
			updatedRecords = append(updatedRecords, record)
		}
	}

	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file for truncation: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range updatedRecords {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("unable to write to file: %w", err)
		}
	}

	return nil
}

// AppendObjectMetadata добавляет метаданные объекта в CSV файл.
func AppendObjectMetadata(filePath, objectKey, creationDate string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{objectKey, creationDate}); err != nil {
		return fmt.Errorf("unable to write to file: %w", err)
	}

	return nil
}
