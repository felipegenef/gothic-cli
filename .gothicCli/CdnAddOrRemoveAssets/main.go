package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	gothicCliShared "{{.GoModName}}/.gothicCli"
)

func main() {
	// Define command-line flags
	action := flag.String("action", "", "Specify the action: add or delete")
	stage := flag.String("stage", "default", "Specify the deployment stage (default, dev, staging, prod)")
	flag.Parse()

	// Open the configuration file
	file, err := os.Open("gothic-config.json")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Create a variable to store the configuration
	var config gothicCliShared.Config

	// Decode the JSON from the file
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	var stageValue string
	if stage != nil {
		stageValue = *stage
	} else {
		stageValue = "default" // or a default value
	}

	// Read the app ID from the file
	content, err := os.ReadFile(".gothicCli/app-id.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Convert the content to string
	appID := string(content)

	// Construct the S3 bucket name
	bucketPublicFolderName := "s3://" + config.ProjectName + "-" + stageValue + "-" + appID + "/public"

	// Check the action and execute the corresponding command
	switch *action {
	case "add":
		addFilesCmd := exec.Command("aws", "s3", "cp", "public", bucketPublicFolderName, "--recursive", "--region", config.Deploy.Region, "--profile", config.Deploy.Profile)
		addFilesCmd.Stdout = os.Stdout
		addFilesCmd.Stdin = os.Stdin
		addFilesCmd.Stderr = os.Stderr

		// Run the command
		if err := addFilesCmd.Run(); err != nil {
			log.Fatalf("Error executing add command: %v", err)
		}
		fmt.Println("S3 Files added successfully.")

	case "delete":
		removeFilesCmd := exec.Command("aws", "s3", "rm", bucketPublicFolderName, "--recursive", "--region", config.Deploy.Region, "--profile", config.Deploy.Profile)
		removeFilesCmd.Stdout = os.Stdout
		removeFilesCmd.Stdin = os.Stdin
		removeFilesCmd.Stderr = os.Stderr

		// Run the command
		if err := removeFilesCmd.Run(); err != nil {
			log.Fatalf("Error executing delete command: %v", err)
		}
		fmt.Println("S3 Files deleted successfully.")

	default:
		log.Fatalf("Invalid action specified: %s. Use 'add' or 'delete'.", *action)
	}
}
