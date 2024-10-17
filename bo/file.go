package bo

import (
	"encoding/csv"
	"fmt"
	"os"
)

func AppendBucketMetadata(filePath, name, creationDate, lastModified, status string) error {
	return appendToCSV(filePath, []string{name, creationDate, lastModified, status})
}

func ReadBucketsMetadata(filePath string) ([][]string, error) {
	return readCSV(filePath)
}

func HasObjects(bucketPath string) bool {
	files, err := os.ReadDir(bucketPath)
	return err == nil && len(files) > 0
}

func RemoveBucketMetadata(filePath, bucketName string) error {
	records, err := readCSV(filePath)
	if err != nil {
		return err
	}

	updatedRecords := filterRecords(records, func(record []string) bool {
		return record[0] != bucketName
	})

	return writeCSV(filePath, updatedRecords)
}

func AppendObjectMetadata(filePath, objectKey, size, contentType, lastModified string) error {
	return appendToCSV(filePath, []string{objectKey, size, contentType, lastModified})
}

func RemoveObjectMetadata(objectMetadataPath, objectKey string) error {
	records, err := readCSV(objectMetadataPath)
	if err != nil {
		return err
	}

	updatedRecords := filterRecords(records, func(record []string) bool {
		return record[0] != objectKey
	})

	return writeCSV(objectMetadataPath, updatedRecords)
}

func UpdateBucketLastModified(filePath, bucketName, lastModified string) error {
	records, err := readCSV(filePath)
	if err != nil {
		return err
	}

	for i, record := range records {
		if record[0] == bucketName {
			records[i][2] = lastModified
			break
		}
	}

	return writeCSV(filePath, records)
}

func appendToCSV(filePath string, record []string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(record); err != nil {
		return fmt.Errorf("unable to write to file: %w", err)
	}

	return nil
}

func readCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

func filterRecords(records [][]string, filterFunc func([]string) bool) [][]string {
	var filtered [][]string
	for _, record := range records {
		if filterFunc(record) {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func writeCSV(filePath string, records [][]string) error {
	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("unable to open file for truncation: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("unable to write to file: %w", err)
		}
	}
	return nil
}
