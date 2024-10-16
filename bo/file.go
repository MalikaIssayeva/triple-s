package bo

import (
	"encoding/csv"
	"fmt"
	"os"
)

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

func HasObjects(bucketPath string) bool {
	files, err := os.ReadDir(bucketPath)
	if err != nil {
		return false
	}
	return len(files) > 0
}

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

func RemoveObjectMetadata(objectMetadataPath, objectKey string) error {
	file, err := os.OpenFile(objectMetadataPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	tempFile, err := os.CreateTemp("", "objects.csv")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	writer := csv.NewWriter(tempFile)
	defer writer.Flush()

	for _, record := range records {
		if record[0] != objectKey {
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	if err := os.Rename(tempFile.Name(), objectMetadataPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func UpdateBucketLastModified(filePath, bucketName, lastModified string) error {
	records, err := ReadBucketsMetadata(filePath)
	if err != nil {
		return err
	}

	for i, record := range records {
		if record[0] == bucketName {
			records[i][2] = lastModified
			break
		}
	}
	file, err := os.OpenFile(filePath, os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
