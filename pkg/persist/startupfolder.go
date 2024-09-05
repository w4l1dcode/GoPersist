package persist

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DropFileToStartup moves a file to the Startup folder
func DropFileToStartup(filePath, fileName string) error {
	startupFolder := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	destinationPath := filepath.Join(startupFolder, fileName)

	// Copy the file to the Startup folder
	err := copyFile(filePath, destinationPath)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	fmt.Printf("File successfully added to Startup folder: %s\n", destinationPath)
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// CreateStartupBatchFile creates a batch file in the Startup folder
func CreateStartupBatchFile(command, arguments, fileName string) error {
	startupFolder := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	batchFilePath := filepath.Join(startupFolder, fileName+".bat")

	// Create the batch file content
	batchContent := fmt.Sprintf(`@echo off
%s %s
`, command, arguments)

	// Write the batch content to the file
	err := os.WriteFile(batchFilePath, []byte(batchContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create batch file: %w", err)
	}

	fmt.Printf("Batch file successfully created in Startup folder: %s\n", batchFilePath)
	return nil
}

// RemoveFileFromStartup deletes a file from the Startup folder
func RemoveFileFromStartup(fileName string) error {
	startupFolder := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	destinationPath := filepath.Join(startupFolder, fileName+".bat")

	// Check if the file exists
	if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", destinationPath)
	}

	fmt.Printf(destinationPath)

	// Remove the file from the Startup folder
	err := os.Remove(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	fmt.Printf("File successfully removed from Startup folder: %s\n", destinationPath)
	return nil
}
