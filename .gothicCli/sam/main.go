package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	gothicCliShared "{{.GoModName}}/.gothicCli"
)

func main() {

	// Define command-line flags
	action := flag.String("action", "deploy", "Specify the action: deploy, delete or build")
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

	switch *action {
	case "deploy":

		samDeployCMD := exec.Command("sam", "deploy", "--stack-name", config.ProjectName+"-"+stageValue, "--parameter-overrides", "Stage="+stageValue, "--profile", config.Deploy.Profile)
		samDeployCMD.Stdout = os.Stdout
		samDeployCMD.Stdin = os.Stdin
		samDeployCMD.Stderr = os.Stderr

		// Run the command
		if err := samDeployCMD.Run(); err != nil {
			log.Fatalf("Error deploying app:%v", err)
		}
		defer cleanupCache(config, stageValue)
		defer runCleanUp()

	case "delete":
		samDeleteCMD := exec.Command("sam", "delete", "--stack-name", config.ProjectName+"-"+stageValue, "--profile", config.Deploy.Profile)
		samDeleteCMD.Stdout = os.Stdout
		samDeleteCMD.Stdin = os.Stdin
		samDeleteCMD.Stderr = os.Stderr

		// Run the command
		if err := samDeleteCMD.Run(); err != nil {
			log.Fatalf("Error deleting app:%v", err)
		}
		defer runCleanUp()
	case "build":
		samBuildCMD := exec.Command("sam", "build")
		samBuildCMD.Stdout = os.Stdout
		samBuildCMD.Stdin = os.Stdin
		samBuildCMD.Stderr = os.Stderr

		// Run the command
		if err := samBuildCMD.Run(); err != nil {
			log.Fatalf("Error building app:%v", err)
		}

	default:
		log.Fatalf("Invalid action specified: %s. Use 'deploy', 'delete' or 'build'.", *action)
	}
}

func cleanupCache(config gothicCliShared.Config, stage string) {
	// Execute the command to get the CloudFront distribution ID
	getDistributionIdCMD := exec.Command("aws", "cloudformation", "describe-stacks", "--stack-name", config.ProjectName+"-"+stage, "--query", "Stacks[0].Outputs[?OutputKey=='CloudFrontId'].OutputValue", "--output", "text", "--region", config.Deploy.Region, "--profile", config.Deploy.Profile)

	// Capture the output of the command
	var out bytes.Buffer
	getDistributionIdCMD.Stdout = &out
	if err := getDistributionIdCMD.Run(); err != nil {
		log.Fatalf("Error getting CloudFront Id: %v", err)
	}

	// The result of the command will be the CloudFront distribution ID
	distributionId := strings.TrimSpace(out.String()) // Remove any extra spaces
	if distributionId == "" {
		log.Fatal("CloudFront ID not found")
	}

	// Now, use the distribution ID in the command to create the invalidation
	cleanCachesCmd := exec.Command("aws", "cloudfront", "create-invalidation", "--distribution-id", distributionId, "--paths", "/*", "--region", config.Deploy.Region, "--profile", config.Deploy.Profile)

	// Execute the cache cleanup command
	if err := cleanCachesCmd.Run(); err != nil {
		log.Fatalf("Error cleaning up deploy files: %v", err)
	}

	// Print the distribution ID and confirm the cache cleanup
	fmt.Printf("Successfully reset CloudFront cache for distribution: %s\n", distributionId)
}

func runCleanUp() {
	cleanUpCMD := exec.Command("go", "run", ".gothicCli/buildSamTemplate/cleanup/main.go")
	if err := cleanUpCMD.Run(); err != nil {
		log.Fatalf("Error cleaning up deploy files: %v", err)
	}
}
