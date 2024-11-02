package main

import (
	"fmt"
	"os"
)

func main() {
	// Define the paths of the files you want to delete
	filesToDelete := []string{
		"Dockerfile",
		"template.yaml",
		"samconfig.toml",
	}

	// Iterate over each file and attempt to delete it
	for _, filePath := range filesToDelete {
		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("Error cleaning up deploy file file %s: %v\n", filePath, err)
			continue
		}
	}
}
